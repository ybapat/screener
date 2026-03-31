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

type datasetRepo struct {
	pool *pgxpool.Pool
}

func NewDatasetRepository(pool *pgxpool.Pool) DatasetRepository {
	return &datasetRepo{pool: pool}
}

func (r *datasetRepo) Create(ctx context.Context, ds *models.Dataset) error {
	query := `
		INSERT INTO datasets (id, title, description, category_filter, contributor_count, record_count,
			date_range_start, date_range_end, k_anonymity_k, epsilon_per_query, noise_mechanism,
			base_price_credits, current_price_credits, age_ranges, countries, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		ds.ID, ds.Title, ds.Description, ds.CategoryFilter, ds.ContributorCount, ds.RecordCount,
		ds.DateRangeStart, ds.DateRangeEnd, ds.KAnonymityK, ds.EpsilonPerQuery, ds.NoiseMechanism,
		ds.BasePriceCredits, ds.CurrentPriceCredits, ds.AgeRanges, ds.Countries, ds.Status,
	).Scan(&ds.CreatedAt, &ds.UpdatedAt)
}

func (r *datasetRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Dataset, error) {
	ds := &models.Dataset{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, title, description, category_filter, contributor_count, record_count,
			date_range_start, date_range_end, k_anonymity_k, epsilon_per_query, noise_mechanism,
			base_price_credits, current_price_credits, age_ranges, countries, status, created_at, updated_at
		FROM datasets WHERE id = $1`, id).Scan(
		&ds.ID, &ds.Title, &ds.Description, &ds.CategoryFilter, &ds.ContributorCount, &ds.RecordCount,
		&ds.DateRangeStart, &ds.DateRangeEnd, &ds.KAnonymityK, &ds.EpsilonPerQuery, &ds.NoiseMechanism,
		&ds.BasePriceCredits, &ds.CurrentPriceCredits, &ds.AgeRanges, &ds.Countries, &ds.Status,
		&ds.CreatedAt, &ds.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get dataset: %w", err)
	}
	return ds, nil
}

func (r *datasetRepo) ListActive(ctx context.Context, categories []string, limit, offset int) ([]models.Dataset, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM datasets WHERE status = 'active'`
	args := []any{}

	if len(categories) > 0 {
		countQuery += ` AND category_filter && $1`
		args = append(args, categories)
	}
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT id, title, description, category_filter, contributor_count, record_count,
			date_range_start, date_range_end, k_anonymity_k, epsilon_per_query, noise_mechanism,
			base_price_credits, current_price_credits, age_ranges, countries, status, created_at, updated_at
		FROM datasets WHERE status = 'active'`

	listArgs := []any{}
	argIdx := 1
	if len(categories) > 0 {
		listQuery += fmt.Sprintf(` AND category_filter && $%d`, argIdx)
		listArgs = append(listArgs, categories)
		argIdx++
	}
	listQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	listArgs = append(listArgs, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var datasets []models.Dataset
	for rows.Next() {
		var ds models.Dataset
		if err := rows.Scan(
			&ds.ID, &ds.Title, &ds.Description, &ds.CategoryFilter, &ds.ContributorCount, &ds.RecordCount,
			&ds.DateRangeStart, &ds.DateRangeEnd, &ds.KAnonymityK, &ds.EpsilonPerQuery, &ds.NoiseMechanism,
			&ds.BasePriceCredits, &ds.CurrentPriceCredits, &ds.AgeRanges, &ds.Countries, &ds.Status,
			&ds.CreatedAt, &ds.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		datasets = append(datasets, ds)
	}
	return datasets, total, nil
}

func (r *datasetRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.DatasetStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE datasets SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

func (r *datasetRepo) UpdatePrice(ctx context.Context, id uuid.UUID, price int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE datasets SET current_price_credits = $2, updated_at = NOW() WHERE id = $1`, id, price)
	return err
}

func (r *datasetRepo) AddContributor(ctx context.Context, dc *models.DatasetContributor) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO dataset_contributors (id, dataset_id, user_id, epsilon_charged, records_included, earning_credits)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		dc.ID, dc.DatasetID, dc.UserID, dc.EpsilonCharged, dc.RecordsIncluded, dc.EarningCredits)
	return err
}

func (r *datasetRepo) GetContributors(ctx context.Context, datasetID uuid.UUID) ([]models.DatasetContributor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, dataset_id, user_id, epsilon_charged, records_included, earning_credits
		FROM dataset_contributors WHERE dataset_id = $1`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []models.DatasetContributor
	for rows.Next() {
		var dc models.DatasetContributor
		if err := rows.Scan(&dc.ID, &dc.DatasetID, &dc.UserID, &dc.EpsilonCharged, &dc.RecordsIncluded, &dc.EarningCredits); err != nil {
			return nil, err
		}
		contributors = append(contributors, dc)
	}
	return contributors, nil
}

func (r *datasetRepo) CreateSample(ctx context.Context, sample *models.DatasetSample) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO dataset_samples (id, dataset_id, app_category, duration_range, time_of_day, device_type, contributor_age_range, contributor_country)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		sample.ID, sample.DatasetID, sample.AppCategory, sample.DurationRange, sample.TimeOfDay,
		sample.DeviceType, sample.ContributorAgeRange, sample.ContributorCountry)
	return err
}

func (r *datasetRepo) GetSamples(ctx context.Context, datasetID uuid.UUID) ([]models.DatasetSample, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, dataset_id, app_category, duration_range, time_of_day, device_type, contributor_age_range, contributor_country
		FROM dataset_samples WHERE dataset_id = $1`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []models.DatasetSample
	for rows.Next() {
		var s models.DatasetSample
		if err := rows.Scan(&s.ID, &s.DatasetID, &s.AppCategory, &s.DurationRange, &s.TimeOfDay,
			&s.DeviceType, &s.ContributorAgeRange, &s.ContributorCountry); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}
	return samples, nil
}
