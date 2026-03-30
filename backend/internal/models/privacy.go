package models

import (
	"time"

	"github.com/google/uuid"
)

type BudgetEventType string

const (
	BudgetEventDatasetSale     BudgetEventType = "dataset_sale"
	BudgetEventQueryResponse   BudgetEventType = "query_response"
	BudgetEventSampleGen       BudgetEventType = "sample_generation"
	BudgetEventRefund          BudgetEventType = "budget_refund"
)

type EpsilonLedgerEntry struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	EventType        BudgetEventType `json:"event_type"`
	EpsilonSpent     float64         `json:"epsilon_spent"`
	EpsilonRemaining float64         `json:"epsilon_remaining"`
	DatasetID        *uuid.UUID      `json:"dataset_id,omitempty"`
	Description      *string         `json:"description,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}
