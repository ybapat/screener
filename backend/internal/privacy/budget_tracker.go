package privacy

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ybapat/screener/backend/internal/models"
)

// BudgetTracker enforces per-user epsilon limits using sequential composition.
// Every time a user's data is included in a dataset, epsilon is debited atomically.
type BudgetTracker struct {
	pool *pgxpool.Pool
}

// NewBudgetTracker creates a new privacy budget tracker.
func NewBudgetTracker(pool *pgxpool.Pool) *BudgetTracker {
	return &BudgetTracker{pool: pool}
}

// CanSpend checks if the user has enough epsilon budget remaining.
// Returns (canSpend, remaining, error).
func (bt *BudgetTracker) CanSpend(ctx context.Context, userID uuid.UUID, epsilon float64) (bool, float64, error) {
	var budget, spent float64
	err := bt.pool.QueryRow(ctx,
		`SELECT global_epsilon_budget, epsilon_spent FROM users WHERE id = $1`, userID,
	).Scan(&budget, &spent)
	if err != nil {
		return false, 0, fmt.Errorf("query budget: %w", err)
	}

	remaining := budget - spent
	return remaining >= epsilon, remaining, nil
}

// Spend atomically debits epsilon from a user's budget and writes an audit log entry.
// Uses a database transaction to ensure the budget check and debit happen atomically.
func (bt *BudgetTracker) Spend(ctx context.Context, userID uuid.UUID, epsilon float64, datasetID *uuid.UUID, description string) error {
	tx, err := bt.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Atomically check and debit
	var remaining float64
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET epsilon_spent = epsilon_spent + $2, updated_at = NOW()
		WHERE id = $1 AND epsilon_spent + $2 <= global_epsilon_budget
		RETURNING global_epsilon_budget - epsilon_spent`,
		userID, epsilon,
	).Scan(&remaining)
	if err != nil {
		return fmt.Errorf("insufficient privacy budget or user not found")
	}

	// Write audit log
	ledgerID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO epsilon_ledger (id, user_id, event_type, epsilon_spent, epsilon_remaining, dataset_id, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ledgerID, userID, models.BudgetEventDatasetSale, epsilon, remaining, datasetID, description,
	)
	if err != nil {
		return fmt.Errorf("write ledger: %w", err)
	}

	return tx.Commit(ctx)
}

// GetRemainingBudget returns the user's remaining epsilon budget.
func (bt *BudgetTracker) GetRemainingBudget(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	var budget, spent float64
	err := bt.pool.QueryRow(ctx,
		`SELECT global_epsilon_budget, epsilon_spent FROM users WHERE id = $1`, userID,
	).Scan(&budget, &spent)
	if err != nil {
		return 0, 0, err
	}
	return budget, spent, nil
}

// GetLedger returns the epsilon audit trail for a user.
func (bt *BudgetTracker) GetLedger(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.EpsilonLedgerEntry, int, error) {
	var total int
	err := bt.pool.QueryRow(ctx, `SELECT COUNT(*) FROM epsilon_ledger WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := bt.pool.Query(ctx, `
		SELECT id, user_id, event_type, epsilon_spent, epsilon_remaining, dataset_id, description, created_at
		FROM epsilon_ledger WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []models.EpsilonLedgerEntry
	for rows.Next() {
		var e models.EpsilonLedgerEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &e.EpsilonSpent, &e.EpsilonRemaining, &e.DatasetID, &e.Description, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, nil
}
