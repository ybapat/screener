package models

import (
	"time"

	"github.com/google/uuid"
)

type DatasetStatus string

const (
	DatasetStatusDraft     DatasetStatus = "draft"
	DatasetStatusActive    DatasetStatus = "active"
	DatasetStatusPaused    DatasetStatus = "paused"
	DatasetStatusExhausted DatasetStatus = "exhausted"
	DatasetStatusWithdrawn DatasetStatus = "withdrawn"
)

type PurchaseStatus string

const (
	PurchaseStatusPending   PurchaseStatus = "pending"
	PurchaseStatusCompleted PurchaseStatus = "completed"
	PurchaseStatusRefunded  PurchaseStatus = "refunded"
	PurchaseStatusFailed    PurchaseStatus = "failed"
)

type Dataset struct {
	ID                  uuid.UUID     `json:"id"`
	Title               string        `json:"title"`
	Description         *string       `json:"description,omitempty"`
	CategoryFilter      []string      `json:"category_filter"`
	ContributorCount    int           `json:"contributor_count"`
	RecordCount         int           `json:"record_count"`
	DateRangeStart      *time.Time    `json:"date_range_start,omitempty"`
	DateRangeEnd        *time.Time    `json:"date_range_end,omitempty"`
	KAnonymityK         int           `json:"k_anonymity_k"`
	EpsilonPerQuery     float64       `json:"epsilon_per_query"`
	NoiseMechanism      string        `json:"noise_mechanism"`
	BasePriceCredits    int64         `json:"base_price_credits"`
	CurrentPriceCredits int64         `json:"current_price_credits"`
	AgeRanges           []string      `json:"age_ranges"`
	Countries           []string      `json:"countries"`
	Status              DatasetStatus `json:"status"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

type DatasetContributor struct {
	ID              uuid.UUID `json:"id"`
	DatasetID       uuid.UUID `json:"dataset_id"`
	UserID          uuid.UUID `json:"user_id"`
	EpsilonCharged  float64   `json:"epsilon_charged"`
	RecordsIncluded int       `json:"records_included"`
	EarningCredits  int64     `json:"earning_credits"`
}

type DatasetSample struct {
	ID                  uuid.UUID `json:"id"`
	DatasetID           uuid.UUID `json:"dataset_id"`
	AppCategory         string    `json:"app_category"`
	DurationRange       string    `json:"duration_range"`
	TimeOfDay           string    `json:"time_of_day"`
	DeviceType          *string   `json:"device_type,omitempty"`
	ContributorAgeRange *string   `json:"contributor_age_range,omitempty"`
	ContributorCountry  *string   `json:"contributor_country,omitempty"`
}

type Purchase struct {
	ID            uuid.UUID      `json:"id"`
	BuyerID       uuid.UUID      `json:"buyer_id"`
	DatasetID     uuid.UUID      `json:"dataset_id"`
	PriceCredits  int64          `json:"price_credits"`
	Status        PurchaseStatus `json:"status"`
	DownloadURL   *string        `json:"download_url,omitempty"`
	DownloadCount int            `json:"download_count"`
	PurchasedAt   time.Time      `json:"purchased_at"`
}
