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

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, display_name, role, age_range, country, timezone, credit_balance, global_epsilon_budget, epsilon_spent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.DisplayName, user.Role,
		user.AgeRange, user.Country, user.Timezone,
		user.CreditBalance, user.GlobalEpsilonBudget, user.EpsilonSpent,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, role, age_range, country, timezone,
		       credit_balance, global_epsilon_budget, epsilon_spent, created_at, updated_at
		FROM users WHERE id = $1`

	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
		&u.AgeRange, &u.Country, &u.Timezone,
		&u.CreditBalance, &u.GlobalEpsilonBudget, &u.EpsilonSpent,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, role, age_range, country, timezone,
		       credit_balance, global_epsilon_budget, epsilon_spent, created_at, updated_at
		FROM users WHERE email = $1`

	u := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
		&u.AgeRange, &u.Country, &u.Timezone,
		&u.CreditBalance, &u.GlobalEpsilonBudget, &u.EpsilonSpent,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET display_name = $2, age_range = $3, country = $4, timezone = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, user.ID, user.DisplayName, user.AgeRange, user.Country, user.Timezone)
	return err
}

func (r *userRepo) UpdateCredits(ctx context.Context, id uuid.UUID, amount int64) (int64, error) {
	query := `
		UPDATE users SET credit_balance = credit_balance + $2, updated_at = NOW()
		WHERE id = $1
		RETURNING credit_balance`

	var balance int64
	err := r.pool.QueryRow(ctx, query, id, amount).Scan(&balance)
	return balance, err
}

func (r *userRepo) UpdateEpsilon(ctx context.Context, id uuid.UUID, epsilonDelta float64) error {
	query := `
		UPDATE users SET epsilon_spent = epsilon_spent + $2, updated_at = NOW()
		WHERE id = $1 AND epsilon_spent + $2 <= global_epsilon_budget`

	tag, err := r.pool.Exec(ctx, query, id, epsilonDelta)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient epsilon budget")
	}
	return nil
}
