package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

type CreditService struct {
	users       repository.UserRepository
	marketplace repository.MarketplaceRepository
}

func NewCreditService(users repository.UserRepository, mp repository.MarketplaceRepository) *CreditService {
	return &CreditService{users: users, marketplace: mp}
}

// TopUp adds mock credits to a user's account.
func (s *CreditService) TopUp(ctx context.Context, userID uuid.UUID, amount int64) (int64, error) {
	if amount <= 0 || amount > 100000 {
		return 0, apierror.BadRequest("amount must be between 1 and 100000")
	}

	newBalance, err := s.users.UpdateCredits(ctx, userID, amount)
	if err != nil {
		return 0, apierror.Internal("failed to update credits")
	}

	desc := fmt.Sprintf("top up %d credits", amount)
	s.marketplace.CreateCreditTransaction(ctx, &models.CreditTransaction{
		ID:           uuid.New(),
		UserID:       userID,
		Amount:       amount,
		BalanceAfter: newBalance,
		TxType:       "topup",
		Description:  &desc,
	})

	return newBalance, nil
}

// Debit subtracts credits for a purchase. Returns new balance.
func (s *CreditService) Debit(ctx context.Context, userID uuid.UUID, amount int64, txType string, refID *uuid.UUID) (int64, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return 0, apierror.Internal("user not found")
	}

	if user.CreditBalance < amount {
		return 0, apierror.BadRequest("insufficient credits")
	}

	newBalance, err := s.users.UpdateCredits(ctx, userID, -amount)
	if err != nil {
		return 0, apierror.Internal("failed to debit credits")
	}

	desc := fmt.Sprintf("debit %d credits for %s", amount, txType)
	s.marketplace.CreateCreditTransaction(ctx, &models.CreditTransaction{
		ID:           uuid.New(),
		UserID:       userID,
		Amount:       -amount,
		BalanceAfter: newBalance,
		TxType:       txType,
		ReferenceID:  refID,
		Description:  &desc,
	})

	return newBalance, nil
}

// Credit adds credits to a seller's account for a data sale.
func (s *CreditService) Credit(ctx context.Context, userID uuid.UUID, amount int64, txType string, refID *uuid.UUID) (int64, error) {
	newBalance, err := s.users.UpdateCredits(ctx, userID, amount)
	if err != nil {
		return 0, apierror.Internal("failed to credit account")
	}

	desc := fmt.Sprintf("earned %d credits from %s", amount, txType)
	s.marketplace.CreateCreditTransaction(ctx, &models.CreditTransaction{
		ID:           uuid.New(),
		UserID:       userID,
		Amount:       amount,
		BalanceAfter: newBalance,
		TxType:       txType,
		ReferenceID:  refID,
		Description:  &desc,
	})

	return newBalance, nil
}

func (s *CreditService) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.CreditTransaction, int, error) {
	return s.marketplace.GetCreditHistory(ctx, userID, limit, offset)
}
