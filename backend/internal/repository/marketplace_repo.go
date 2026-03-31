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

type marketplaceRepo struct {
	pool *pgxpool.Pool
}

func NewMarketplaceRepository(pool *pgxpool.Pool) MarketplaceRepository {
	return &marketplaceRepo{pool: pool}
}

func (r *marketplaceRepo) CreateSegment(ctx context.Context, seg *models.DataSegment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO data_segments (id, buyer_id, app_categories, date_range_start, date_range_end,
			age_ranges, countries, device_types, min_contributors, min_records, desired_k_anonymity, max_epsilon)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		seg.ID, seg.BuyerID, seg.AppCategories, seg.DateRangeStart, seg.DateRangeEnd,
		seg.AgeRanges, seg.Countries, seg.DeviceTypes, seg.MinContributors, seg.MinRecords,
		seg.DesiredK, seg.MaxEpsilon)
	return err
}

func (r *marketplaceRepo) GetSegmentByID(ctx context.Context, id uuid.UUID) (*models.DataSegment, error) {
	seg := &models.DataSegment{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, buyer_id, app_categories, date_range_start, date_range_end,
			age_ranges, countries, device_types, min_contributors, min_records,
			desired_k_anonymity, max_epsilon, created_at
		FROM data_segments WHERE id = $1`, id).Scan(
		&seg.ID, &seg.BuyerID, &seg.AppCategories, &seg.DateRangeStart, &seg.DateRangeEnd,
		&seg.AgeRanges, &seg.Countries, &seg.DeviceTypes, &seg.MinContributors, &seg.MinRecords,
		&seg.DesiredK, &seg.MaxEpsilon, &seg.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get segment: %w", err)
	}
	return seg, nil
}

func (r *marketplaceRepo) ListSegmentsByBuyer(ctx context.Context, buyerID uuid.UUID) ([]models.DataSegment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, buyer_id, app_categories, date_range_start, date_range_end,
			age_ranges, countries, device_types, min_contributors, min_records,
			desired_k_anonymity, max_epsilon, created_at
		FROM data_segments WHERE buyer_id = $1 ORDER BY created_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []models.DataSegment
	for rows.Next() {
		var seg models.DataSegment
		if err := rows.Scan(&seg.ID, &seg.BuyerID, &seg.AppCategories, &seg.DateRangeStart, &seg.DateRangeEnd,
			&seg.AgeRanges, &seg.Countries, &seg.DeviceTypes, &seg.MinContributors, &seg.MinRecords,
			&seg.DesiredK, &seg.MaxEpsilon, &seg.CreatedAt); err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}
	return segments, nil
}

func (r *marketplaceRepo) CreateBid(ctx context.Context, bid *models.Bid) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bids (id, segment_id, buyer_id, bid_credits, status, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		bid.ID, bid.SegmentID, bid.BuyerID, bid.BidCredits, bid.Status, bid.ExpiresAt)
	return err
}

func (r *marketplaceRepo) GetBidByID(ctx context.Context, id uuid.UUID) (*models.Bid, error) {
	bid := &models.Bid{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, segment_id, buyer_id, bid_credits, status, expires_at, dataset_id, created_at, updated_at
		FROM bids WHERE id = $1`, id).Scan(
		&bid.ID, &bid.SegmentID, &bid.BuyerID, &bid.BidCredits, &bid.Status,
		&bid.ExpiresAt, &bid.DatasetID, &bid.CreatedAt, &bid.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get bid: %w", err)
	}
	return bid, nil
}

func (r *marketplaceRepo) ListBidsByBuyer(ctx context.Context, buyerID uuid.UUID) ([]models.Bid, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, segment_id, buyer_id, bid_credits, status, expires_at, dataset_id, created_at, updated_at
		FROM bids WHERE buyer_id = $1 ORDER BY created_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var b models.Bid
		if err := rows.Scan(&b.ID, &b.SegmentID, &b.BuyerID, &b.BidCredits, &b.Status,
			&b.ExpiresAt, &b.DatasetID, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, nil
}

func (r *marketplaceRepo) ListActiveBidsBySegment(ctx context.Context, segmentID uuid.UUID) ([]models.Bid, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, segment_id, buyer_id, bid_credits, status, expires_at, dataset_id, created_at, updated_at
		FROM bids WHERE segment_id = $1 AND status = 'active' ORDER BY bid_credits DESC`, segmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var b models.Bid
		if err := rows.Scan(&b.ID, &b.SegmentID, &b.BuyerID, &b.BidCredits, &b.Status,
			&b.ExpiresAt, &b.DatasetID, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, nil
}

func (r *marketplaceRepo) UpdateBidStatus(ctx context.Context, id uuid.UUID, status models.BidStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE bids SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *marketplaceRepo) ExpireOldBids(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE bids SET status = 'expired', updated_at = NOW() WHERE status = 'active' AND expires_at < $1`, time.Now())
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *marketplaceRepo) CountActiveBidsForCategories(ctx context.Context, categories []string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT b.id)
		FROM bids b JOIN data_segments s ON b.segment_id = s.id
		WHERE b.status = 'active' AND s.app_categories && $1`, categories).Scan(&count)
	return count, err
}

func (r *marketplaceRepo) CreateCreditTransaction(ctx context.Context, tx *models.CreditTransaction) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO credit_transactions (id, user_id, amount, balance_after, tx_type, reference_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		tx.ID, tx.UserID, tx.Amount, tx.BalanceAfter, tx.TxType, tx.ReferenceID, tx.Description)
	return err
}

func (r *marketplaceRepo) GetCreditHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.CreditTransaction, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM credit_transactions WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, balance_after, tx_type, reference_id, description, created_at
		FROM credit_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []models.CreditTransaction
	for rows.Next() {
		var t models.CreditTransaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.BalanceAfter, &t.TxType,
			&t.ReferenceID, &t.Description, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		txns = append(txns, t)
	}
	return txns, total, nil
}

func (r *marketplaceRepo) CreatePriceHistory(ctx context.Context, ph *models.PriceHistory) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO price_history (id, dataset_id, app_category, price_credits, demand_score, rarity_score)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		ph.ID, ph.DatasetID, ph.AppCategory, ph.PriceCredits, ph.DemandScore, ph.RarityScore)
	return err
}
