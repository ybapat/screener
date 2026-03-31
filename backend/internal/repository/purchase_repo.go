package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ybapat/screener/backend/internal/models"
)

type purchaseRepo struct {
	pool *pgxpool.Pool
}

func NewPurchaseRepository(pool *pgxpool.Pool) PurchaseRepository {
	return &purchaseRepo{pool: pool}
}

func (r *purchaseRepo) Create(ctx context.Context, p *models.Purchase) error {
	query := `
		INSERT INTO purchases (id, buyer_id, dataset_id, price_credits, status)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING purchased_at`

	return r.pool.QueryRow(ctx, query, p.ID, p.BuyerID, p.DatasetID, p.PriceCredits, p.Status).
		Scan(&p.PurchasedAt)
}

func (r *purchaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error) {
	p := &models.Purchase{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, buyer_id, dataset_id, price_credits, status, download_url, download_count, purchased_at
		FROM purchases WHERE id = $1`, id).Scan(
		&p.ID, &p.BuyerID, &p.DatasetID, &p.PriceCredits, &p.Status,
		&p.DownloadURL, &p.DownloadCount, &p.PurchasedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get purchase: %w", err)
	}
	return p, nil
}

func (r *purchaseRepo) GetByBuyer(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]models.Purchase, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM purchases WHERE buyer_id = $1`, buyerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, buyer_id, dataset_id, price_credits, status, download_url, download_count, purchased_at
		FROM purchases WHERE buyer_id = $1 ORDER BY purchased_at DESC LIMIT $2 OFFSET $3`,
		buyerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var purchases []models.Purchase
	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(&p.ID, &p.BuyerID, &p.DatasetID, &p.PriceCredits, &p.Status,
			&p.DownloadURL, &p.DownloadCount, &p.PurchasedAt); err != nil {
			return nil, 0, err
		}
		purchases = append(purchases, p)
	}
	return purchases, total, nil
}

func (r *purchaseRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.PurchaseStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE purchases SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *purchaseRepo) IncrementDownloads(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE purchases SET download_count = download_count + 1 WHERE id = $1`, id)
	return err
}

func (r *purchaseRepo) HasPurchased(ctx context.Context, buyerID, datasetID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM purchases WHERE buyer_id = $1 AND dataset_id = $2 AND status = 'completed'`,
		buyerID, datasetID).Scan(&count)
	return count > 0, err
}
