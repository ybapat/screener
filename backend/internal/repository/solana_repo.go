package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ybapat/screener/backend/internal/models"
)

type solanaRepo struct {
	pool *pgxpool.Pool
}

func NewSolanaRepository(pool *pgxpool.Pool) SolanaRepository {
	return &solanaRepo{pool: pool}
}

func (r *solanaRepo) CreateSolTransaction(ctx context.Context, tx *models.SolTransaction) error {
	query := `
		INSERT INTO sol_transactions (id, user_id, tx_signature, tx_type, amount_lamports, from_wallet, to_wallet, status, reference_id, confirmed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.UserID, tx.TxSignature, tx.TxType, tx.AmountLamports,
		tx.FromWallet, tx.ToWallet, tx.Status, tx.ReferenceID, tx.ConfirmedAt, tx.CreatedAt,
	)
	return err
}

func (r *solanaRepo) GetSolTransactionBySignature(ctx context.Context, sig string) (*models.SolTransaction, error) {
	query := `
		SELECT id, user_id, tx_signature, tx_type, amount_lamports, from_wallet, to_wallet, status, reference_id, confirmed_at, created_at
		FROM sol_transactions WHERE tx_signature = $1`

	tx := &models.SolTransaction{}
	err := r.pool.QueryRow(ctx, query, sig).Scan(
		&tx.ID, &tx.UserID, &tx.TxSignature, &tx.TxType, &tx.AmountLamports,
		&tx.FromWallet, &tx.ToWallet, &tx.Status, &tx.ReferenceID, &tx.ConfirmedAt, &tx.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sol tx by sig: %w", err)
	}
	return tx, nil
}

func (r *solanaRepo) UpdateSolTransactionStatus(ctx context.Context, id uuid.UUID, status models.SolTxStatus) error {
	query := `UPDATE sol_transactions SET status = $2, confirmed_at = $3 WHERE id = $1`
	now := time.Now()
	var confirmedAt *time.Time
	if status == models.SolTxConfirmed {
		confirmedAt = &now
	}
	_, err := r.pool.Exec(ctx, query, id, status, confirmedAt)
	return err
}

func (r *solanaRepo) GetSolTransactionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SolTransaction, int, error) {
	countQuery := `SELECT COUNT(*) FROM sol_transactions WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, tx_signature, tx_type, amount_lamports, from_wallet, to_wallet, status, reference_id, confirmed_at, created_at
		FROM sol_transactions WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []models.SolTransaction
	for rows.Next() {
		var tx models.SolTransaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.TxSignature, &tx.TxType, &tx.AmountLamports,
			&tx.FromWallet, &tx.ToWallet, &tx.Status, &tx.ReferenceID, &tx.ConfirmedAt, &tx.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	return txs, total, nil
}

func (r *solanaRepo) CreateEscrow(ctx context.Context, escrow *models.SolEscrow) error {
	query := `
		INSERT INTO sol_escrows (id, buyer_id, dataset_id, escrow_pda, vault_pda, amount_lamports, released_lamports, status, deposit_signature, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		escrow.ID, escrow.BuyerID, escrow.DatasetID, escrow.EscrowPDA, escrow.VaultPDA,
		escrow.AmountLamports, escrow.ReleasedLamports, escrow.Status, escrow.DepositSignature, escrow.CreatedAt,
	)
	return err
}

func (r *solanaRepo) GetEscrow(ctx context.Context, id uuid.UUID) (*models.SolEscrow, error) {
	query := `
		SELECT id, buyer_id, dataset_id, escrow_pda, vault_pda, amount_lamports, released_lamports, status, deposit_signature, created_at
		FROM sol_escrows WHERE id = $1`

	e := &models.SolEscrow{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.BuyerID, &e.DatasetID, &e.EscrowPDA, &e.VaultPDA,
		&e.AmountLamports, &e.ReleasedLamports, &e.Status, &e.DepositSignature, &e.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get escrow: %w", err)
	}
	return e, nil
}

func (r *solanaRepo) GetEscrowByBuyerAndDataset(ctx context.Context, buyerID, datasetID uuid.UUID) (*models.SolEscrow, error) {
	query := `
		SELECT id, buyer_id, dataset_id, escrow_pda, vault_pda, amount_lamports, released_lamports, status, deposit_signature, created_at
		FROM sol_escrows WHERE buyer_id = $1 AND dataset_id = $2`

	e := &models.SolEscrow{}
	err := r.pool.QueryRow(ctx, query, buyerID, datasetID).Scan(
		&e.ID, &e.BuyerID, &e.DatasetID, &e.EscrowPDA, &e.VaultPDA,
		&e.AmountLamports, &e.ReleasedLamports, &e.Status, &e.DepositSignature, &e.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get escrow by buyer/dataset: %w", err)
	}
	return e, nil
}

func (r *solanaRepo) UpdateEscrow(ctx context.Context, escrow *models.SolEscrow) error {
	query := `
		UPDATE sol_escrows SET released_lamports = $2, status = $3
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, escrow.ID, escrow.ReleasedLamports, escrow.Status)
	return err
}

func (r *solanaRepo) GetConfig(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM sol_config WHERE key = $1`
	var val string
	err := r.pool.QueryRow(ctx, query, key).Scan(&val)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return val, err
}

func (r *solanaRepo) SetConfig(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO sol_config (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`
	_, err := r.pool.Exec(ctx, query, key, value)
	return err
}
