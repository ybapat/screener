package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ybapat/screener/backend/internal/models"
)

type screenTimeRepo struct {
	pool *pgxpool.Pool
}

func NewScreenTimeRepository(pool *pgxpool.Pool) ScreenTimeRepository {
	return &screenTimeRepo{pool: pool}
}

func (r *screenTimeRepo) CreateBatch(ctx context.Context, batch *models.DataBatch) error {
	query := `
		INSERT INTO data_batches (id, user_id, record_count, date_range_start, date_range_end, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	return r.pool.QueryRow(ctx, query,
		batch.ID, batch.UserID, batch.RecordCount,
		batch.DateRangeStart, batch.DateRangeEnd, batch.Status,
	).Scan(&batch.CreatedAt)
}

func (r *screenTimeRepo) InsertRecords(ctx context.Context, records []models.ScreenTimeRecord) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, rec := range records {
		_, err := tx.Exec(ctx, `
			INSERT INTO screentime_records (id, user_id, batch_id, app_name, app_category, duration_secs, started_at, ended_at, device_type, os, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			rec.ID, rec.UserID, rec.BatchID, rec.AppName, rec.AppCategory,
			rec.DurationSec, rec.StartedAt, rec.EndedAt, rec.DeviceType, rec.OS, rec.Status,
		)
		if err != nil {
			return fmt.Errorf("insert record: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *screenTimeRepo) GetBatchesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DataBatch, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM data_batches WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, record_count, date_range_start, date_range_end, status, created_at
		FROM data_batches WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var batches []models.DataBatch
	for rows.Next() {
		var b models.DataBatch
		if err := rows.Scan(&b.ID, &b.UserID, &b.RecordCount, &b.DateRangeStart, &b.DateRangeEnd, &b.Status, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		batches = append(batches, b)
	}
	return batches, total, nil
}

func (r *screenTimeRepo) GetBatchByID(ctx context.Context, id uuid.UUID) (*models.DataBatch, error) {
	b := &models.DataBatch{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, record_count, date_range_start, date_range_end, status, created_at
		FROM data_batches WHERE id = $1`, id).
		Scan(&b.ID, &b.UserID, &b.RecordCount, &b.DateRangeStart, &b.DateRangeEnd, &b.Status, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *screenTimeRepo) GetRecordsByBatch(ctx context.Context, batchID uuid.UUID) ([]models.ScreenTimeRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, batch_id, app_name, app_category, duration_secs, started_at, ended_at, device_type, os, status, created_at
		FROM screentime_records WHERE batch_id = $1 ORDER BY started_at`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ScreenTimeRecord
	for rows.Next() {
		var rec models.ScreenTimeRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.BatchID, &rec.AppName, &rec.AppCategory,
			&rec.DurationSec, &rec.StartedAt, &rec.EndedAt, &rec.DeviceType, &rec.OS, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *screenTimeRepo) GetRecordsByUserAndCategories(ctx context.Context, userID uuid.UUID, categories []string) ([]models.ScreenTimeRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, batch_id, app_name, app_category, duration_secs, started_at, ended_at, device_type, os, status, created_at
		FROM screentime_records WHERE user_id = $1 AND app_category = ANY($2) AND status = 'validated'
		ORDER BY started_at`, userID, categories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ScreenTimeRecord
	for rows.Next() {
		var rec models.ScreenTimeRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.BatchID, &rec.AppName, &rec.AppCategory,
			&rec.DurationSec, &rec.StartedAt, &rec.EndedAt, &rec.DeviceType, &rec.OS, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *screenTimeRepo) UpdateBatchStatus(ctx context.Context, id uuid.UUID, status models.DataStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE data_batches SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *screenTimeRepo) GetAvailableRecords(ctx context.Context, categories []string, minContributors int) ([]models.ScreenTimeRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.user_id, r.batch_id, r.app_name, r.app_category, r.duration_secs, r.started_at, r.ended_at, r.device_type, r.os, r.status, r.created_at
		FROM screentime_records r
		JOIN users u ON r.user_id = u.id
		WHERE r.app_category = ANY($1)
		  AND r.status IN ('validated', 'raw')
		  AND u.epsilon_spent < u.global_epsilon_budget
		ORDER BY r.started_at`, categories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ScreenTimeRecord
	for rows.Next() {
		var rec models.ScreenTimeRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.BatchID, &rec.AppName, &rec.AppCategory,
			&rec.DurationSec, &rec.StartedAt, &rec.EndedAt, &rec.DeviceType, &rec.OS, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}
