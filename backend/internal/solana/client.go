package solana

import (
	"context"
	"fmt"
	"log/slog"

	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/programs/system"
)

// Client wraps the Solana RPC client and server keypair.
type Client struct {
	rpc       *rpc.Client
	serverKey solanago.PrivateKey
	programID solanago.PublicKey
}

// NewClient creates a new Solana RPC client.
func NewClient(rpcURL, keypairPath, programID string) (*Client, error) {
	rpcClient := rpc.New(rpcURL)

	key, err := LoadOrGenerateKeypair(keypairPath)
	if err != nil {
		return nil, fmt.Errorf("load keypair: %w", err)
	}

	var progID solanago.PublicKey
	if programID != "" {
		progID = solanago.MustPublicKeyFromBase58(programID)
	}

	return &Client{
		rpc:       rpcClient,
		serverKey: key,
		programID: progID,
	}, nil
}

// ServerPublicKey returns the server wallet's public key.
func (c *Client) ServerPublicKey() solanago.PublicKey {
	return c.serverKey.PublicKey()
}

// ProgramID returns the escrow program's public key.
func (c *Client) ProgramID() solanago.PublicKey {
	return c.programID
}

// RPC returns the underlying RPC client for direct queries.
func (c *Client) RPC() *rpc.Client {
	return c.rpc
}

// GetBalance returns the lamport balance of a public key.
func (c *Client) GetBalance(ctx context.Context, pubkey solanago.PublicKey) (uint64, error) {
	out, err := c.rpc.GetBalance(ctx, pubkey, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return out.Value, nil
}

// RequestAirdrop requests devnet SOL for a public key.
func (c *Client) RequestAirdrop(ctx context.Context, pubkey solanago.PublicKey, lamports uint64) (string, error) {
	sig, err := c.rpc.RequestAirdrop(ctx, pubkey, lamports, rpc.CommitmentConfirmed)
	if err != nil {
		return "", fmt.Errorf("request airdrop: %w", err)
	}
	slog.Info("airdrop requested", "pubkey", pubkey.String(), "lamports", lamports, "sig", sig.String())
	return sig.String(), nil
}

// VerifyTransfer verifies that an on-chain transaction is a SystemProgram.Transfer
// from expectedFrom to expectedTo for expectedLamports, and is confirmed.
func (c *Client) VerifyTransfer(ctx context.Context, signature string, expectedFrom, expectedTo solanago.PublicKey, expectedLamports uint64) error {
	sig := solanago.MustSignatureFromBase58(signature)

	maxVersion := uint64(0)
	tx, err := c.rpc.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment:                     rpc.CommitmentConfirmed,
		MaxSupportedTransactionVersion: &maxVersion,
	})
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}
	if tx.Meta != nil && tx.Meta.Err != nil {
		return fmt.Errorf("transaction failed on-chain")
	}

	// Parse the transaction to find SystemProgram.Transfer instructions
	decoded, err := tx.Transaction.GetTransaction()
	if err != nil {
		return fmt.Errorf("decode transaction: %w", err)
	}

	for _, inst := range decoded.Message.Instructions {
		progKey, err := decoded.Message.Program(inst.ProgramIDIndex)
		if err != nil {
			continue
		}
		if !progKey.Equals(solanago.SystemProgramID) {
			continue
		}

		// Decode system instruction
		if len(inst.Data) < 4 {
			continue
		}

		// SystemProgram.Transfer instruction type is 2 (little-endian u32)
		instructionType := uint32(inst.Data[0]) | uint32(inst.Data[1])<<8 | uint32(inst.Data[2])<<16 | uint32(inst.Data[3])<<24
		if instructionType != 2 {
			continue
		}

		if len(inst.Data) < 12 {
			continue
		}

		// Lamports is a u64 at offset 4
		lamports := uint64(inst.Data[4]) | uint64(inst.Data[5])<<8 | uint64(inst.Data[6])<<16 | uint64(inst.Data[7])<<24 |
			uint64(inst.Data[8])<<32 | uint64(inst.Data[9])<<40 | uint64(inst.Data[10])<<48 | uint64(inst.Data[11])<<56

		if len(inst.Accounts) < 2 {
			continue
		}

		fromKey, err := decoded.Message.Account(inst.Accounts[0])
		if err != nil {
			continue
		}
		toKey, err := decoded.Message.Account(inst.Accounts[1])
		if err != nil {
			continue
		}

		if fromKey.Equals(expectedFrom) && toKey.Equals(expectedTo) && lamports == expectedLamports {
			return nil // Verified!
		}
	}

	return fmt.Errorf("no matching transfer instruction found")
}

// SendSOL builds, signs, and sends a SystemProgram.Transfer from the server keypair.
func (c *Client) SendSOL(ctx context.Context, to solanago.PublicKey, lamports uint64) (string, error) {
	recent, err := c.rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("get blockhash: %w", err)
	}

	tx, err := solanago.NewTransaction(
		[]solanago.Instruction{
			system.NewTransferInstruction(lamports, c.serverKey.PublicKey(), to).Build(),
		},
		recent.Value.Blockhash,
		solanago.TransactionPayer(c.serverKey.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("build transaction: %w", err)
	}

	_, err = tx.Sign(func(key solanago.PublicKey) *solanago.PrivateKey {
		if key.Equals(c.serverKey.PublicKey()) {
			return &c.serverKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("sign transaction: %w", err)
	}

	sig, err := c.rpc.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("send transaction: %w", err)
	}

	slog.Info("sent SOL", "to", to.String(), "lamports", lamports, "sig", sig.String())
	return sig.String(), nil
}

// DeriveEscrowPDA computes the escrow state PDA for a buyer and dataset.
func (c *Client) DeriveEscrowPDA(buyer solanago.PublicKey, datasetIDBytes [16]byte) (solanago.PublicKey, uint8, error) {
	addr, bump, err := solanago.FindProgramAddress(
		[][]byte{
			[]byte("escrow"),
			buyer.Bytes(),
			datasetIDBytes[:],
		},
		c.programID,
	)
	return addr, bump, err
}

// DeriveVaultPDA computes the vault PDA for a buyer and dataset.
func (c *Client) DeriveVaultPDA(buyer solanago.PublicKey, datasetIDBytes [16]byte) (solanago.PublicKey, uint8, error) {
	addr, bump, err := solanago.FindProgramAddress(
		[][]byte{
			[]byte("vault"),
			buyer.Bytes(),
			datasetIDBytes[:],
		},
		c.programID,
	)
	return addr, bump, err
}

// ReleaseEscrow sends a release instruction to the escrow program, transferring
// SOL from the vault PDA to a seller. Only callable with the server keypair (authority).
func (c *Client) ReleaseEscrow(ctx context.Context, escrowPDA, vaultPDA, seller, buyer solanago.PublicKey, amount uint64) (string, error) {
	recent, err := c.rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("get blockhash: %w", err)
	}

	// Build the release instruction
	// Accounts: authority (signer), escrow_state (mut), vault (mut), seller (mut), system_program
	data := make([]byte, 8+8)
	// Anchor discriminator for "release": sha256("global:release")[:8]
	disc := anchorDiscriminator("release")
	copy(data[:8], disc[:])
	// amount as u64 little-endian
	putUint64LE(data[8:], amount)

	inst := solanago.NewInstruction(
		c.programID,
		solanago.AccountMetaSlice{
			solanago.NewAccountMeta(c.serverKey.PublicKey(), false, true), // authority (signer)
			solanago.NewAccountMeta(escrowPDA, true, false),              // escrow_state (mut)
			solanago.NewAccountMeta(vaultPDA, true, false),               // vault (mut)
			solanago.NewAccountMeta(seller, true, false),                 // seller (mut)
			solanago.NewAccountMeta(solanago.SystemProgramID, false, false),
		},
		data,
	)

	tx, err := solanago.NewTransaction(
		[]solanago.Instruction{inst},
		recent.Value.Blockhash,
		solanago.TransactionPayer(c.serverKey.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("build release tx: %w", err)
	}

	_, err = tx.Sign(func(key solanago.PublicKey) *solanago.PrivateKey {
		if key.Equals(c.serverKey.PublicKey()) {
			return &c.serverKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("sign release tx: %w", err)
	}

	sig, err := c.rpc.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("send release tx: %w", err)
	}

	slog.Info("escrow released", "seller", seller.String(), "amount", amount, "sig", sig.String())
	return sig.String(), nil
}

// RefundEscrow sends a refund instruction to return remaining SOL to the buyer.
func (c *Client) RefundEscrow(ctx context.Context, escrowPDA, vaultPDA, buyer solanago.PublicKey) (string, error) {
	recent, err := c.rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("get blockhash: %w", err)
	}

	disc := anchorDiscriminator("refund")
	data := make([]byte, 8)
	copy(data, disc[:])

	inst := solanago.NewInstruction(
		c.programID,
		solanago.AccountMetaSlice{
			solanago.NewAccountMeta(c.serverKey.PublicKey(), false, true), // authority (signer)
			solanago.NewAccountMeta(escrowPDA, true, false),              // escrow_state (mut)
			solanago.NewAccountMeta(vaultPDA, true, false),               // vault (mut)
			solanago.NewAccountMeta(buyer, true, false),                  // buyer (mut)
			solanago.NewAccountMeta(solanago.SystemProgramID, false, false),
		},
		data,
	)

	tx, err := solanago.NewTransaction(
		[]solanago.Instruction{inst},
		recent.Value.Blockhash,
		solanago.TransactionPayer(c.serverKey.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("build refund tx: %w", err)
	}

	_, err = tx.Sign(func(key solanago.PublicKey) *solanago.PrivateKey {
		if key.Equals(c.serverKey.PublicKey()) {
			return &c.serverKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("sign refund tx: %w", err)
	}

	sig, err := c.rpc.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("send refund tx: %w", err)
	}

	slog.Info("escrow refunded", "buyer", buyer.String(), "sig", sig.String())
	return sig.String(), nil
}
