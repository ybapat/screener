package models

import (
	"time"

	"github.com/google/uuid"
)

type BidStatus string

const (
	BidStatusActive   BidStatus = "active"
	BidStatusAccepted BidStatus = "accepted"
	BidStatusRejected BidStatus = "rejected"
	BidStatusExpired  BidStatus = "expired"
	BidStatusCancelled BidStatus = "cancelled"
)

type DataSegment struct {
	ID              uuid.UUID  `json:"id"`
	BuyerID         uuid.UUID  `json:"buyer_id"`
	AppCategories   []string   `json:"app_categories"`
	DateRangeStart  *time.Time `json:"date_range_start,omitempty"`
	DateRangeEnd    *time.Time `json:"date_range_end,omitempty"`
	AgeRanges       []string   `json:"age_ranges,omitempty"`
	Countries       []string   `json:"countries,omitempty"`
	DeviceTypes     []string   `json:"device_types,omitempty"`
	MinContributors int        `json:"min_contributors"`
	MinRecords      int        `json:"min_records"`
	DesiredK        int        `json:"desired_k_anonymity"`
	MaxEpsilon      float64    `json:"max_epsilon"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Bid struct {
	ID         uuid.UUID  `json:"id"`
	SegmentID  uuid.UUID  `json:"segment_id"`
	BuyerID    uuid.UUID  `json:"buyer_id"`
	BidCredits int64      `json:"bid_credits"`
	Status     BidStatus  `json:"status"`
	ExpiresAt  time.Time  `json:"expires_at"`
	DatasetID  *uuid.UUID `json:"dataset_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreditTransaction struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Amount       int64      `json:"amount"`
	BalanceAfter int64      `json:"balance_after"`
	TxType       string     `json:"tx_type"`
	ReferenceID  *uuid.UUID `json:"reference_id,omitempty"`
	Description  *string    `json:"description,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PriceHistory struct {
	ID           uuid.UUID  `json:"id"`
	DatasetID    *uuid.UUID `json:"dataset_id,omitempty"`
	AppCategory  string     `json:"app_category"`
	PriceCredits int64      `json:"price_credits"`
	DemandScore  float64    `json:"demand_score"`
	RarityScore  float64    `json:"rarity_score"`
	RecordedAt   time.Time  `json:"recorded_at"`
}
