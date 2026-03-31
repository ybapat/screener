package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

type IngestionService struct {
	screenTime repository.ScreenTimeRepository
}

func NewIngestionService(st repository.ScreenTimeRepository) *IngestionService {
	return &IngestionService{screenTime: st}
}

type UploadResult struct {
	BatchID          uuid.UUID `json:"batch_id"`
	RecordsAccepted  int       `json:"records_accepted"`
	RecordsRejected  int       `json:"records_rejected"`
	ValidationErrors []string  `json:"validation_errors"`
}

func (s *IngestionService) Upload(ctx context.Context, userID uuid.UUID, input []models.ScreenTimeRecordInput) (*UploadResult, error) {
	if len(input) == 0 {
		return nil, apierror.BadRequest("no records provided")
	}
	if len(input) > 1000 {
		return nil, apierror.BadRequest("max 1000 records per upload")
	}

	batchID := uuid.New()
	var accepted []models.ScreenTimeRecord
	var validationErrors []string

	var earliest, latest time.Time
	for i, rec := range input {
		if errs := validateRecord(rec); len(errs) > 0 {
			for _, e := range errs {
				validationErrors = append(validationErrors, fmt.Sprintf("record[%d]: %s", i, e))
			}
			continue
		}

		id := uuid.New()
		accepted = append(accepted, models.ScreenTimeRecord{
			ID:          id,
			UserID:      userID,
			BatchID:     &batchID,
			AppName:     rec.AppName,
			AppCategory: rec.AppCategory,
			DurationSec: rec.DurationSec,
			StartedAt:   rec.StartedAt,
			EndedAt:     rec.EndedAt,
			DeviceType:  rec.DeviceType,
			OS:          rec.OS,
			Status:      models.DataStatusRaw,
		})

		if earliest.IsZero() || rec.StartedAt.Before(earliest) {
			earliest = rec.StartedAt
		}
		if latest.IsZero() || rec.EndedAt.After(latest) {
			latest = rec.EndedAt
		}
	}

	if len(accepted) == 0 {
		return &UploadResult{
			BatchID:          batchID,
			RecordsAccepted:  0,
			RecordsRejected:  len(input),
			ValidationErrors: validationErrors,
		}, nil
	}

	batch := &models.DataBatch{
		ID:             batchID,
		UserID:         userID,
		RecordCount:    len(accepted),
		DateRangeStart: &earliest,
		DateRangeEnd:   &latest,
		Status:         models.DataStatusRaw,
	}

	if err := s.screenTime.CreateBatch(ctx, batch); err != nil {
		return nil, apierror.Internal("failed to create batch")
	}

	if err := s.screenTime.InsertRecords(ctx, accepted); err != nil {
		return nil, apierror.Internal("failed to insert records")
	}

	return &UploadResult{
		BatchID:          batchID,
		RecordsAccepted:  len(accepted),
		RecordsRejected:  len(input) - len(accepted),
		ValidationErrors: validationErrors,
	}, nil
}

func (s *IngestionService) GetBatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DataBatch, int, error) {
	return s.screenTime.GetBatchesByUser(ctx, userID, limit, offset)
}

func (s *IngestionService) GetBatch(ctx context.Context, batchID uuid.UUID) (*models.DataBatch, error) {
	return s.screenTime.GetBatchByID(ctx, batchID)
}

func (s *IngestionService) GetBatchRecords(ctx context.Context, batchID uuid.UUID) ([]models.ScreenTimeRecord, error) {
	return s.screenTime.GetRecordsByBatch(ctx, batchID)
}

func (s *IngestionService) WithdrawBatch(ctx context.Context, batchID uuid.UUID, userID uuid.UUID) error {
	batch, err := s.screenTime.GetBatchByID(ctx, batchID)
	if err != nil {
		return apierror.Internal("failed to get batch")
	}
	if batch.UserID != userID {
		return apierror.Forbidden("not your batch")
	}
	if batch.Status == models.DataStatusSold {
		return apierror.BadRequest("cannot withdraw sold data")
	}
	return s.screenTime.UpdateBatchStatus(ctx, batchID, models.DataStatusWithdrawn)
}

func validateRecord(rec models.ScreenTimeRecordInput) []string {
	var errs []string
	if rec.DurationSec <= 0 || rec.DurationSec >= 86400 {
		errs = append(errs, "duration must be between 1 and 86399 seconds")
	}
	if !rec.EndedAt.After(rec.StartedAt) {
		errs = append(errs, "ended_at must be after started_at")
	}
	if rec.StartedAt.After(time.Now().Add(time.Minute)) {
		errs = append(errs, "started_at cannot be in the future")
	}
	actualDuration := rec.EndedAt.Sub(rec.StartedAt).Seconds()
	diff := actualDuration - float64(rec.DurationSec)
	if diff > 60 || diff < -60 {
		errs = append(errs, "duration_secs does not match started_at/ended_at (>60s tolerance)")
	}
	return errs
}
