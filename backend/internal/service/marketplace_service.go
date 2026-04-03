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

type MarketplaceService struct {
	datasets    repository.DatasetRepository
	purchases   repository.PurchaseRepository
	marketplace repository.MarketplaceRepository
	credits     *CreditService
}

func NewMarketplaceService(
	ds repository.DatasetRepository,
	p repository.PurchaseRepository,
	mp repository.MarketplaceRepository,
	credits *CreditService,
) *MarketplaceService {
	return &MarketplaceService{
		datasets:    ds,
		purchases:   p,
		marketplace: mp,
		credits:     credits,
	}
}

func (s *MarketplaceService) ListDatasets(ctx context.Context, categories []string, limit, offset int) ([]models.Dataset, int, error) {
	return s.datasets.ListActive(ctx, categories, limit, offset)
}

func (s *MarketplaceService) GetDataset(ctx context.Context, id uuid.UUID) (*models.Dataset, error) {
	return s.datasets.GetByID(ctx, id)
}

func (s *MarketplaceService) GetSamples(ctx context.Context, datasetID uuid.UUID) ([]models.DatasetSample, error) {
	return s.datasets.GetSamples(ctx, datasetID)
}

// Purchase handles the full purchase flow: check balance, debit buyer, credit sellers, create purchase record.
func (s *MarketplaceService) Purchase(ctx context.Context, buyerID, datasetID uuid.UUID) (*models.Purchase, error) {
	dataset, err := s.datasets.GetByID(ctx, datasetID)
	if err != nil || dataset == nil {
		return nil, apierror.NotFound("dataset not found")
	}
	if dataset.Status != models.DatasetStatusActive {
		return nil, apierror.BadRequest("dataset is not available for purchase")
	}

	already, err := s.purchases.HasPurchased(ctx, buyerID, datasetID)
	if err != nil {
		return nil, apierror.Internal("failed to check purchase history")
	}
	if already {
		return nil, apierror.Conflict("you already purchased this dataset")
	}

	price := dataset.CurrentPriceCredits
	purchaseID := uuid.New()

	// Debit buyer
	_, err = s.credits.Debit(ctx, buyerID, price, "purchase", &purchaseID)
	if err != nil {
		return nil, err
	}

	// Credit sellers proportionally
	contributors, err := s.datasets.GetContributors(ctx, datasetID)
	if err == nil && len(contributors) > 0 {
		perContributor := price / int64(len(contributors))
		for _, c := range contributors {
			s.credits.Credit(ctx, c.UserID, perContributor, "data_sale", &purchaseID)
		}
	}

	purchase := &models.Purchase{
		ID:           purchaseID,
		BuyerID:      buyerID,
		DatasetID:    datasetID,
		PriceCredits: price,
		Status:       models.PurchaseStatusCompleted,
	}

	if err := s.purchases.Create(ctx, purchase); err != nil {
		return nil, apierror.Internal("failed to create purchase record")
	}

	return purchase, nil
}

// CompleteSolPurchase creates a purchase record for a SOL-paid transaction.
// The buyer already paid on-chain, so no credit debit is needed.
// Returns the purchase and the list of contributors for SOL payout.
func (s *MarketplaceService) CompleteSolPurchase(ctx context.Context, buyerID, datasetID uuid.UUID) (*models.Purchase, []models.DatasetContributor, error) {
	dataset, err := s.datasets.GetByID(ctx, datasetID)
	if err != nil || dataset == nil {
		return nil, nil, apierror.NotFound("dataset not found")
	}
	if dataset.Status != models.DatasetStatusActive {
		return nil, nil, apierror.BadRequest("dataset is not available for purchase")
	}

	already, err := s.purchases.HasPurchased(ctx, buyerID, datasetID)
	if err != nil {
		return nil, nil, apierror.Internal("failed to check purchase history")
	}
	if already {
		return nil, nil, apierror.Conflict("you already purchased this dataset")
	}

	purchaseID := uuid.New()
	purchase := &models.Purchase{
		ID:           purchaseID,
		BuyerID:      buyerID,
		DatasetID:    datasetID,
		PriceCredits: dataset.CurrentPriceCredits,
		Status:       models.PurchaseStatusCompleted,
	}

	if err := s.purchases.Create(ctx, purchase); err != nil {
		return nil, nil, apierror.Internal("failed to create purchase record")
	}

	contributors, _ := s.datasets.GetContributors(ctx, datasetID)
	return purchase, contributors, nil
}

func (s *MarketplaceService) GetPurchases(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]models.Purchase, int, error) {
	return s.purchases.GetByBuyer(ctx, buyerID, limit, offset)
}

// Segments and Bids

func (s *MarketplaceService) CreateSegment(ctx context.Context, buyerID uuid.UUID, seg *models.DataSegment) error {
	seg.ID = uuid.New()
	seg.BuyerID = buyerID
	seg.CreatedAt = time.Now()
	return s.marketplace.CreateSegment(ctx, seg)
}

func (s *MarketplaceService) ListSegments(ctx context.Context, buyerID uuid.UUID) ([]models.DataSegment, error) {
	return s.marketplace.ListSegmentsByBuyer(ctx, buyerID)
}

func (s *MarketplaceService) PlaceBid(ctx context.Context, buyerID, segmentID uuid.UUID, credits int64, duration time.Duration) (*models.Bid, error) {
	seg, err := s.marketplace.GetSegmentByID(ctx, segmentID)
	if err != nil || seg == nil {
		return nil, apierror.NotFound("segment not found")
	}
	if seg.BuyerID != buyerID {
		return nil, apierror.Forbidden("not your segment")
	}

	bid := &models.Bid{
		ID:         uuid.New(),
		SegmentID:  segmentID,
		BuyerID:    buyerID,
		BidCredits: credits,
		Status:     models.BidStatusActive,
		ExpiresAt:  time.Now().Add(duration),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.marketplace.CreateBid(ctx, bid); err != nil {
		return nil, apierror.Internal(fmt.Sprintf("failed to create bid: %v", err))
	}
	return bid, nil
}

func (s *MarketplaceService) ListBids(ctx context.Context, buyerID uuid.UUID) ([]models.Bid, error) {
	return s.marketplace.ListBidsByBuyer(ctx, buyerID)
}

func (s *MarketplaceService) CancelBid(ctx context.Context, buyerID, bidID uuid.UUID) error {
	bid, err := s.marketplace.GetBidByID(ctx, bidID)
	if err != nil || bid == nil {
		return apierror.NotFound("bid not found")
	}
	if bid.BuyerID != buyerID {
		return apierror.Forbidden("not your bid")
	}
	if bid.Status != models.BidStatusActive {
		return apierror.BadRequest("bid is not active")
	}
	return s.marketplace.UpdateBidStatus(ctx, bidID, models.BidStatusCancelled)
}
