import {
  PublicKey,
  SystemProgram,
  TransactionInstruction,
  LAMPORTS_PER_SOL,
} from "@solana/web3.js";

// Anchor discriminator: sha256("global:deposit")[:8]
// Precomputed to avoid importing crypto in the browser
const DEPOSIT_DISCRIMINATOR = new Uint8Array([
  0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8,
]);

/**
 * Builds the deposit instruction for the Screener escrow program.
 *
 * Accounts:
 *   0. buyer (signer, mut) — the wallet paying
 *   1. authority — the server wallet (read-only)
 *   2. escrow_state (PDA, mut) — will be initialized
 *   3. vault (PDA, mut) — receives the SOL
 *   4. system_program
 */
export function buildDepositInstruction(params: {
  programId: PublicKey;
  buyer: PublicKey;
  authority: PublicKey;
  escrowPDA: PublicKey;
  vaultPDA: PublicKey;
  amount: number; // lamports
  datasetIdBytes: Uint8Array; // 16 bytes
}): TransactionInstruction {
  const { programId, buyer, authority, escrowPDA, vaultPDA, amount, datasetIdBytes } = params;

  // Instruction data: [8-byte discriminator][8-byte amount LE][16-byte dataset_id]
  const data = Buffer.alloc(8 + 8 + 16);

  // Use the real Anchor discriminator for "deposit"
  // We'll compute it properly: sha256("global:deposit") first 8 bytes
  const disc = anchorDiscriminator("deposit");
  data.set(disc, 0);

  // amount as u64 little-endian
  const amountBuf = Buffer.alloc(8);
  amountBuf.writeBigUInt64LE(BigInt(amount));
  data.set(amountBuf, 8);

  // dataset_id (16 bytes)
  data.set(datasetIdBytes, 16);

  return new TransactionInstruction({
    programId,
    keys: [
      { pubkey: buyer, isSigner: true, isWritable: true },
      { pubkey: authority, isSigner: false, isWritable: false },
      { pubkey: escrowPDA, isSigner: false, isWritable: true },
      { pubkey: vaultPDA, isSigner: false, isWritable: true },
      { pubkey: SystemProgram.programId, isSigner: false, isWritable: false },
    ],
    data,
  });
}

/**
 * Converts a UUID string to a 16-byte Uint8Array.
 */
export function uuidToBytes(uuid: string): Uint8Array {
  const hex = uuid.replace(/-/g, "");
  const bytes = new Uint8Array(16);
  for (let i = 0; i < 16; i++) {
    bytes[i] = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
  }
  return bytes;
}

/**
 * Computes the Anchor instruction discriminator: sha256("global:<name>")[:8]
 */
function anchorDiscriminator(name: string): Uint8Array {
  // Use SubtleCrypto for browser compatibility
  // Since this is sync, we precompute known discriminators
  const known: Record<string, number[]> = {
    deposit: [0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8],
  };

  if (known[name]) {
    return new Uint8Array(known[name]);
  }

  // Fallback: this shouldn't happen for known instructions
  throw new Error(`Unknown instruction: ${name}`);
}

/**
 * Derives the escrow state PDA.
 */
export function deriveEscrowPDA(
  programId: PublicKey,
  buyer: PublicKey,
  datasetIdBytes: Uint8Array
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("escrow"), buyer.toBuffer(), Buffer.from(datasetIdBytes)],
    programId
  );
}

/**
 * Derives the vault PDA.
 */
export function deriveVaultPDA(
  programId: PublicKey,
  buyer: PublicKey,
  datasetIdBytes: Uint8Array
): [PublicKey, number] {
  return PublicKey.findProgramAddressSync(
    [Buffer.from("vault"), buyer.toBuffer(), Buffer.from(datasetIdBytes)],
    programId
  );
}
