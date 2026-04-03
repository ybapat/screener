DROP TABLE IF EXISTS sol_config;
DROP TABLE IF EXISTS sol_escrows;
DROP TYPE IF EXISTS escrow_status;
DROP TABLE IF EXISTS sol_transactions;
DROP TYPE IF EXISTS sol_tx_type;
DROP TYPE IF EXISTS sol_tx_status;
ALTER TABLE users DROP COLUMN IF EXISTS solana_wallet;
