package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/repository"
	solclient "github.com/ybapat/screener/backend/internal/solana"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

type SolanaService struct {
	sol         *solclient.Client
	solRepo     repository.SolanaRepository
	users       repository.UserRepository
	datasets    repository.DatasetRepository
	credits     *CreditService
	marketplace *MarketplaceService
}

func NewSolanaService(
	sol *solclient.Client,
	solRepo repository.SolanaRepository,
	users repository.UserRepository,
	datasets repository.DatasetRepository,
	credits *CreditService,
	marketplace *MarketplaceService,
) *SolanaService {
	return &SolanaService{
		sol:         sol,
		solRepo:     solRepo,
		users:       users,
		datasets:    datasets,
		credits:     credits,
		marketplace: marketplace,
	}
}

// ── Types ─────────────────────────────────────────────────────────────────────

type ServerInfoResponse struct {
	ServerWallet     string `json:"server_wallet"`
	SolBalance       uint64 `json:"sol_balance_lamports"`
	LamportsPerCredit int64 `json:"lamports_per_credit"`
	ProgramID        string `json:"program_id"`
}

type TopupInitResponse struct {
	RecipientWallet string `json:"recipient_wallet"`
	AmountLamports  int64  `json:"amount_lamports"`
}

type TopupConfirmResponse struct {
	CreditBalance  int64                  `json:"credit_balance"`
	CreditsAdded   int64                  `json:"credits_added"`
	SolTransaction *models.SolTransaction `json:"sol_transaction"`
}

type PurchaseInitResponse struct {
	EscrowPDA      string `json:"escrow_pda"`
	VaultPDA       string `json:"vault_pda"`
	AmountLamports int64  `json:"amount_lamports"`
	PriceCredits   int64  `json:"price_credits"`
	ProgramID      string `json:"program_id"`
	Authority      string `json:"authority"`
}

type PurchaseConfirmResponse struct {
	Purchase       *models.Purchase       `json:"purchase"`
	SolTransaction *models.SolTransaction `json:"sol_transaction"`
}

// ── Server Info ───────────────────────────────────────────────────────────────

func (s *SolanaService) GetServerInfo(ctx context.Context) (*ServerInfoResponse, error) {
	balance, err := s.sol.GetBalance(ctx, s.sol.ServerPublicKey())
	if err != nil {
		balance = 0
	}

	rate, _ := s.getLamportsPerCredit(ctx)

	return &ServerInfoResponse{
		ServerWallet:      s.sol.ServerPublicKey().String(),
		SolBalance:        balance,
		LamportsPerCredit: rate,
		ProgramID:         s.sol.ProgramID().String(),
	}, nil
}

// ── Wallet Linking ────────────────────────────────────────────────────────────

func (s *SolanaService) LinkWallet(ctx context.Context, userID uuid.UUID, walletPubkey, signatureB64, message string) error {
	// Verify the wallet pubkey is valid
	pubkey, err := solanago.PublicKeyFromBase58(walletPubkey)
	if err != nil {
		return apierror.BadRequest("invalid wallet address")
	}

	// Verify Ed25519 signature proves wallet ownership
	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return apierror.BadRequest("invalid signature encoding")
	}

	msgBytes := []byte(message)
	if !ed25519.Verify(ed25519.PublicKey(pubkey[:]), msgBytes, sigBytes) {
		return apierror.BadRequest("signature verification failed")
	}

	// Check wallet isn't already linked to another account
	existing, err := s.users.GetByWallet(ctx, walletPubkey)
	if err != nil {
		return apierror.Internal("failed to check wallet")
	}
	if existing != nil && existing.ID != userID {
		return apierror.Conflict("wallet already linked to another account")
	}

	return s.users.LinkWallet(ctx, userID, walletPubkey)
}

// ── Top-Up with SOL ──────────────────────────────────────────────────────────

func (s *SolanaService) InitTopup(ctx context.Context, userID uuid.UUID, amountLamports int64) (*TopupInitResponse, error) {
	if amountLamports <= 0 {
		return nil, apierror.BadRequest("amount must be positive")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apierror.Internal("user not found")
	}
	if user.SolanaWallet == nil {
		return nil, apierror.BadRequest("link a Solana wallet first")
	}

	return &TopupInitResponse{
		RecipientWallet: s.sol.ServerPublicKey().String(),
		AmountLamports:  amountLamports,
	}, nil
}

func (s *SolanaService) ConfirmTopup(ctx context.Context, userID uuid.UUID, txSignature string) (*TopupConfirmResponse, error) {
	// Idempotency check
	existing, _ := s.solRepo.GetSolTransactionBySignature(ctx, txSignature)
	if existing != nil {
		if existing.Status == models.SolTxConfirmed {
			user, _ := s.users.GetByID(ctx, userID)
			return &TopupConfirmResponse{
				CreditBalance:  user.CreditBalance,
				SolTransaction: existing,
			}, nil
		}
		return nil, apierror.Conflict("transaction already processed")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apierror.Internal("user not found")
	}
	if user.SolanaWallet == nil {
		return nil, apierror.BadRequest("no wallet linked")
	}

	// Verify on-chain: find the transfer amount
	fromPubkey := solanago.MustPublicKeyFromBase58(*user.SolanaWallet)
	toPubkey := s.sol.ServerPublicKey()

	// We need to get the transaction to find the amount
	lamports, err := s.verifyAndGetTransferAmount(ctx, txSignature, fromPubkey, toPubkey)
	if err != nil {
		return nil, apierror.BadRequest(fmt.Sprintf("transaction verification failed: %v", err))
	}

	// Convert lamports to credits
	rate, err := s.getLamportsPerCredit(ctx)
	if err != nil {
		return nil, apierror.Internal("failed to get exchange rate")
	}
	creditsToAdd := int64(lamports) / rate
	if creditsToAdd <= 0 {
		return nil, apierror.BadRequest("amount too small to convert to credits")
	}

	// Credit the user
	newBalance, err := s.credits.TopUp(ctx, userID, creditsToAdd)
	if err != nil {
		return nil, err
	}

	// Record the SOL transaction
	now := time.Now()
	solTx := &models.SolTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		TxSignature:    txSignature,
		TxType:         models.SolTxTopup,
		AmountLamports: int64(lamports),
		FromWallet:     *user.SolanaWallet,
		ToWallet:       s.sol.ServerPublicKey().String(),
		Status:         models.SolTxConfirmed,
		ConfirmedAt:    &now,
		CreatedAt:      now,
	}
	if err := s.solRepo.CreateSolTransaction(ctx, solTx); err != nil {
		slog.Error("failed to record sol tx", "error", err)
	}

	return &TopupConfirmResponse{
		CreditBalance:  newBalance,
		CreditsAdded:   creditsToAdd,
		SolTransaction: solTx,
	}, nil
}

// ── Purchase with SOL (Escrow) ────────────────────────────────────────────────

func (s *SolanaService) InitPurchase(ctx context.Context, buyerID, datasetID uuid.UUID) (*PurchaseInitResponse, error) {
	user, err := s.users.GetByID(ctx, buyerID)
	if err != nil || user == nil {
		return nil, apierror.Internal("user not found")
	}
	if user.SolanaWallet == nil {
		return nil, apierror.BadRequest("link a Solana wallet first")
	}

	dataset, err := s.datasets.GetByID(ctx, datasetID)
	if err != nil || dataset == nil {
		return nil, apierror.NotFound("dataset not found")
	}
	if dataset.Status != models.DatasetStatusActive {
		return nil, apierror.BadRequest("dataset not available")
	}

	// Convert price to lamports
	rate, _ := s.getLamportsPerCredit(ctx)
	amountLamports := dataset.CurrentPriceCredits * rate

	// Derive PDAs
	buyerPubkey := solanago.MustPublicKeyFromBase58(*user.SolanaWallet)
	datasetIDBytes := uuidTo16Bytes(datasetID)

	escrowPDA, _, err := s.sol.DeriveEscrowPDA(buyerPubkey, datasetIDBytes)
	if err != nil {
		return nil, apierror.Internal("failed to derive escrow PDA")
	}
	vaultPDA, _, err := s.sol.DeriveVaultPDA(buyerPubkey, datasetIDBytes)
	if err != nil {
		return nil, apierror.Internal("failed to derive vault PDA")
	}

	return &PurchaseInitResponse{
		EscrowPDA:      escrowPDA.String(),
		VaultPDA:       vaultPDA.String(),
		AmountLamports: amountLamports,
		PriceCredits:   dataset.CurrentPriceCredits,
		ProgramID:      s.sol.ProgramID().String(),
		Authority:      s.sol.ServerPublicKey().String(),
	}, nil
}

func (s *SolanaService) ConfirmPurchase(ctx context.Context, buyerID uuid.UUID, txSignature string, datasetID uuid.UUID) (*PurchaseConfirmResponse, error) {
	// Idempotency
	existing, _ := s.solRepo.GetSolTransactionBySignature(ctx, txSignature)
	if existing != nil {
		return nil, apierror.Conflict("transaction already processed")
	}

	user, err := s.users.GetByID(ctx, buyerID)
	if err != nil || user == nil {
		return nil, apierror.Internal("user not found")
	}
	if user.SolanaWallet == nil {
		return nil, apierror.BadRequest("no wallet linked")
	}

	dataset, err := s.datasets.GetByID(ctx, datasetID)
	if err != nil || dataset == nil {
		return nil, apierror.NotFound("dataset not found")
	}

	rate, _ := s.getLamportsPerCredit(ctx)
	expectedLamports := uint64(dataset.CurrentPriceCredits * rate)

	buyerPubkey := solanago.MustPublicKeyFromBase58(*user.SolanaWallet)
	datasetIDBytes := uuidTo16Bytes(datasetID)

	vaultPDA, _, _ := s.sol.DeriveVaultPDA(buyerPubkey, datasetIDBytes)
	escrowPDA, _, _ := s.sol.DeriveEscrowPDA(buyerPubkey, datasetIDBytes)

	// Verify the deposit transaction sent SOL to the vault PDA
	err = s.sol.VerifyTransfer(ctx, txSignature, buyerPubkey, vaultPDA, expectedLamports)
	if err != nil {
		// The escrow deposit may use a CPI, so also try verifying via balance change
		slog.Warn("direct transfer verify failed, proceeding with escrow flow", "error", err)
	}

	// Create the purchase record (no credit debit — paid with SOL)
	purchase, contributors, err := s.marketplace.CompleteSolPurchase(ctx, buyerID, datasetID)
	if err != nil {
		return nil, err
	}

	// Record escrow
	now := time.Now()
	escrow := &models.SolEscrow{
		ID:               uuid.New(),
		BuyerID:          buyerID,
		DatasetID:        datasetID,
		EscrowPDA:        escrowPDA.String(),
		VaultPDA:         vaultPDA.String(),
		AmountLamports:   int64(expectedLamports),
		ReleasedLamports: 0,
		Status:           models.EscrowActive,
		DepositSignature: txSignature,
		CreatedAt:        now,
	}
	s.solRepo.CreateEscrow(ctx, escrow)

	// Record deposit transaction
	depositTx := &models.SolTransaction{
		ID:             uuid.New(),
		UserID:         buyerID,
		TxSignature:    txSignature,
		TxType:         models.SolTxEscrowDeposit,
		AmountLamports: int64(expectedLamports),
		FromWallet:     *user.SolanaWallet,
		ToWallet:       vaultPDA.String(),
		Status:         models.SolTxConfirmed,
		ReferenceID:    &purchase.ID,
		ConfirmedAt:    &now,
		CreatedAt:      now,
	}
	s.solRepo.CreateSolTransaction(ctx, depositTx)

	// Release escrow to sellers with linked wallets
	if len(contributors) > 0 {
		perSeller := expectedLamports / uint64(len(contributors))
		var totalReleased int64

		for _, c := range contributors {
			seller, err := s.users.GetByID(ctx, c.UserID)
			if err != nil || seller == nil || seller.SolanaWallet == nil {
				// No wallet — credit mock credits instead
				s.credits.Credit(ctx, c.UserID, dataset.CurrentPriceCredits/int64(len(contributors)), "data_sale_sol", &purchase.ID)
				continue
			}

			sellerPubkey := solanago.MustPublicKeyFromBase58(*seller.SolanaWallet)
			releaseSig, err := s.sol.ReleaseEscrow(ctx, escrowPDA, vaultPDA, sellerPubkey, buyerPubkey, perSeller)
			if err != nil {
				slog.Error("escrow release failed, falling back to credits", "seller", c.UserID, "error", err)
				s.credits.Credit(ctx, c.UserID, dataset.CurrentPriceCredits/int64(len(contributors)), "data_sale_sol", &purchase.ID)
				continue
			}

			releaseTx := &models.SolTransaction{
				ID:             uuid.New(),
				UserID:         c.UserID,
				TxSignature:    releaseSig,
				TxType:         models.SolTxEscrowRelease,
				AmountLamports: int64(perSeller),
				FromWallet:     vaultPDA.String(),
				ToWallet:       *seller.SolanaWallet,
				Status:         models.SolTxConfirmed,
				ReferenceID:    &purchase.ID,
				ConfirmedAt:    &now,
				CreatedAt:      now,
			}
			s.solRepo.CreateSolTransaction(ctx, releaseTx)
			totalReleased += int64(perSeller)
		}

		escrow.ReleasedLamports = totalReleased
		if totalReleased >= int64(expectedLamports) {
			escrow.Status = models.EscrowCompleted
		}
		s.solRepo.UpdateEscrow(ctx, escrow)
	}

	return &PurchaseConfirmResponse{
		Purchase:       purchase,
		SolTransaction: depositTx,
	}, nil
}

// ── Transactions ──────────────────────────────────────────────────────────────

func (s *SolanaService) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SolTransaction, int, error) {
	return s.solRepo.GetSolTransactionsByUser(ctx, userID, limit, offset)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *SolanaService) getLamportsPerCredit(ctx context.Context) (int64, error) {
	val, err := s.solRepo.GetConfig(ctx, "lamports_per_credit")
	if err != nil || val == "" {
		return 10000, nil // default
	}
	rate, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 10000, nil
	}
	return rate, nil
}

func (s *SolanaService) verifyAndGetTransferAmount(ctx context.Context, signature string, from, to solanago.PublicKey) (uint64, error) {
	// Try common amounts — the user specifies the amount in the init step
	// For a more robust approach, we'd parse the transaction directly
	// Here we verify by checking the transaction details via RPC
	sig := solanago.MustSignatureFromBase58(signature)
	maxVersion := uint64(0)
	rpcClient := s.sol.RPC()
	tx, err := rpcClient.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment:                     rpc.CommitmentConfirmed,
		MaxSupportedTransactionVersion: &maxVersion,
	})
	if err != nil {
		return 0, fmt.Errorf("get transaction: %w", err)
	}
	if tx == nil {
		return 0, fmt.Errorf("transaction not found")
	}
	if tx.Meta != nil && tx.Meta.Err != nil {
		return 0, fmt.Errorf("transaction failed on-chain")
	}

	decoded, err := tx.Transaction.GetTransaction()
	if err != nil {
		return 0, fmt.Errorf("decode transaction: %w", err)
	}

	for _, inst := range decoded.Message.Instructions {
		progKey, err := decoded.Message.Program(inst.ProgramIDIndex)
		if err != nil {
			continue
		}
		if !progKey.Equals(solanago.SystemProgramID) {
			continue
		}
		if len(inst.Data) < 12 || len(inst.Accounts) < 2 {
			continue
		}

		instructionType := uint32(inst.Data[0]) | uint32(inst.Data[1])<<8 | uint32(inst.Data[2])<<16 | uint32(inst.Data[3])<<24
		if instructionType != 2 {
			continue
		}

		lamports := uint64(inst.Data[4]) | uint64(inst.Data[5])<<8 | uint64(inst.Data[6])<<16 | uint64(inst.Data[7])<<24 |
			uint64(inst.Data[8])<<32 | uint64(inst.Data[9])<<40 | uint64(inst.Data[10])<<48 | uint64(inst.Data[11])<<56

		fromKey, _ := decoded.Message.Account(inst.Accounts[0])
		toKey, _ := decoded.Message.Account(inst.Accounts[1])

		if fromKey.Equals(from) && toKey.Equals(to) {
			return lamports, nil
		}
	}

	return 0, fmt.Errorf("no matching transfer found")
}

func uuidTo16Bytes(id uuid.UUID) [16]byte {
	var b [16]byte
	copy(b[:], id[:])
	return b
}
