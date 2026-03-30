package models

import (
	"time"

	"github.com/google/uuid"
)

type DataStatus string

const (
	DataStatusRaw        DataStatus = "raw"
	DataStatusValidated  DataStatus = "validated"
	DataStatusAnonymized DataStatus = "anonymized"
	DataStatusListed     DataStatus = "listed"
	DataStatusSold       DataStatus = "sold"
	DataStatusWithdrawn  DataStatus = "withdrawn"
)

type ScreenTimeRecord struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	BatchID     *uuid.UUID `json:"batch_id,omitempty"`
	AppName     string     `json:"app_name"`
	AppCategory string     `json:"app_category"`
	DurationSec int        `json:"duration_secs"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     time.Time  `json:"ended_at"`
	DeviceType  *string    `json:"device_type,omitempty"`
	OS          *string    `json:"os,omitempty"`
	Status      DataStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DataBatch struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	RecordCount    int        `json:"record_count"`
	DateRangeStart *time.Time `json:"date_range_start,omitempty"`
	DateRangeEnd   *time.Time `json:"date_range_end,omitempty"`
	Status         DataStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SharingPreference struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	AppCategory    string    `json:"app_category"`
	ShareAppName   bool      `json:"share_app_name"`
	ShareDuration  bool      `json:"share_duration"`
	ShareTimeOfDay bool      `json:"share_time_of_day"`
	ShareDevice    bool      `json:"share_device"`
	MinPriceCredits int64    `json:"min_price_credits"`
}

// ScreenTimeUploadRequest is the payload for POST /api/v1/data/upload.
type ScreenTimeUploadRequest struct {
	Records []ScreenTimeRecordInput `json:"records" validate:"required,min=1,max=1000,dive"`
}

type ScreenTimeRecordInput struct {
	AppName     string    `json:"app_name" validate:"required,max=255"`
	AppCategory string    `json:"app_category" validate:"required,max=100"`
	DurationSec int       `json:"duration_secs" validate:"required,gt=0,lt=86400"`
	StartedAt   time.Time `json:"started_at" validate:"required"`
	EndedAt     time.Time `json:"ended_at" validate:"required,gtfield=StartedAt"`
	DeviceType  *string   `json:"device_type,omitempty" validate:"omitempty,oneof=phone tablet desktop"`
	OS          *string   `json:"os,omitempty" validate:"omitempty,oneof=ios android macos windows linux"`
}
