-- Add optional Solana wallet to users
ALTER TABLE users ADD COLUMN solana_wallet TEXT;
CREATE UNIQUE INDEX idx_users_solana_wallet ON users(solana_wallet) WHERE solana_wallet IS NOT NULL;

-- Solana transaction types
CREATE TYPE sol_tx_status AS ENUM ('pending', 'confirmed', 'failed');
CREATE TYPE sol_tx_type AS ENUM ('topup', 'purchase', 'seller_payout', 'escrow_deposit', 'escrow_release', 'escrow_refund');

-- All on-chain Solana transactions
CREATE TABLE sol_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    tx_signature    TEXT NOT NULL UNIQUE,
    tx_type         sol_tx_type NOT NULL,
    amount_lamports BIGINT NOT NULL,
    from_wallet     TEXT NOT NULL,
    to_wallet       TEXT NOT NULL,
    status          sol_tx_status NOT NULL DEFAULT 'pending',
    reference_id    UUID,
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sol_tx_user ON sol_transactions(user_id);

-- Escrow state (mirrors on-chain EscrowState account)
CREATE TYPE escrow_status AS ENUM ('active', 'completed', 'refunded');

CREATE TABLE sol_escrows (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id          UUID NOT NULL REFERENCES users(id),
    dataset_id        UUID NOT NULL REFERENCES datasets(id),
    escrow_pda        TEXT NOT NULL,
    vault_pda         TEXT NOT NULL,
    amount_lamports   BIGINT NOT NULL,
    released_lamports BIGINT NOT NULL DEFAULT 0,
    status            escrow_status NOT NULL DEFAULT 'active',
    deposit_signature TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(buyer_id, dataset_id)
);

CREATE INDEX idx_sol_escrows_buyer ON sol_escrows(buyer_id);

-- Configuration (exchange rate, program ID)
CREATE TABLE sol_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO sol_config (key, value) VALUES
    ('lamports_per_credit', '10000'),
    ('program_id', '');
