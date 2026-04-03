package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateCredits(ctx context.Context, id uuid.UUID, amount int64) (int64, error)
	UpdateEpsilon(ctx context.Context, id uuid.UUID, epsilonDelta float64) error
	LinkWallet(ctx context.Context, userID uuid.UUID, wallet string) error
	GetByWallet(ctx context.Context, wallet string) (*models.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type ScreenTimeRepository interface {
	CreateBatch(ctx context.Context, batch *models.DataBatch) error
	InsertRecords(ctx context.Context, records []models.ScreenTimeRecord) error
	GetBatchesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DataBatch, int, error)
	GetBatchByID(ctx context.Context, id uuid.UUID) (*models.DataBatch, error)
	GetRecordsByBatch(ctx context.Context, batchID uuid.UUID) ([]models.ScreenTimeRecord, error)
	GetRecordsByUserAndCategories(ctx context.Context, userID uuid.UUID, categories []string) ([]models.ScreenTimeRecord, error)
	UpdateBatchStatus(ctx context.Context, id uuid.UUID, status models.DataStatus) error
	GetAvailableRecords(ctx context.Context, categories []string, minContributors int) ([]models.ScreenTimeRecord, error)
}

type SharingPreferenceRepository interface {
	Upsert(ctx context.Context, pref *models.SharingPreference) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.SharingPreference, error)
	GetByUserAndCategory(ctx context.Context, userID uuid.UUID, category string) (*models.SharingPreference, error)
}

type DatasetRepository interface {
	Create(ctx context.Context, ds *models.Dataset) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Dataset, error)
	ListActive(ctx context.Context, categories []string, limit, offset int) ([]models.Dataset, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.DatasetStatus) error
	UpdatePrice(ctx context.Context, id uuid.UUID, price int64) error
	AddContributor(ctx context.Context, dc *models.DatasetContributor) error
	GetContributors(ctx context.Context, datasetID uuid.UUID) ([]models.DatasetContributor, error)
	CreateSample(ctx context.Context, sample *models.DatasetSample) error
	GetSamples(ctx context.Context, datasetID uuid.UUID) ([]models.DatasetSample, error)
}

type PurchaseRepository interface {
	Create(ctx context.Context, p *models.Purchase) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error)
	GetByBuyer(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]models.Purchase, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.PurchaseStatus) error
	IncrementDownloads(ctx context.Context, id uuid.UUID) error
	HasPurchased(ctx context.Context, buyerID, datasetID uuid.UUID) (bool, error)
}

type PrivacyRepository interface {
	CreateLedgerEntry(ctx context.Context, entry *models.EpsilonLedgerEntry) error
	GetLedgerByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.EpsilonLedgerEntry, int, error)
}

type SolanaRepository interface {
	CreateSolTransaction(ctx context.Context, tx *models.SolTransaction) error
	GetSolTransactionBySignature(ctx context.Context, sig string) (*models.SolTransaction, error)
	UpdateSolTransactionStatus(ctx context.Context, id uuid.UUID, status models.SolTxStatus) error
	GetSolTransactionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SolTransaction, int, error)
	CreateEscrow(ctx context.Context, escrow *models.SolEscrow) error
	GetEscrow(ctx context.Context, id uuid.UUID) (*models.SolEscrow, error)
	GetEscrowByBuyerAndDataset(ctx context.Context, buyerID, datasetID uuid.UUID) (*models.SolEscrow, error)
	UpdateEscrow(ctx context.Context, escrow *models.SolEscrow) error
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value string) error
}

type MarketplaceRepository interface {
	CreateSegment(ctx context.Context, seg *models.DataSegment) error
	GetSegmentByID(ctx context.Context, id uuid.UUID) (*models.DataSegment, error)
	ListSegmentsByBuyer(ctx context.Context, buyerID uuid.UUID) ([]models.DataSegment, error)
	CreateBid(ctx context.Context, bid *models.Bid) error
	GetBidByID(ctx context.Context, id uuid.UUID) (*models.Bid, error)
	ListBidsByBuyer(ctx context.Context, buyerID uuid.UUID) ([]models.Bid, error)
	ListActiveBidsBySegment(ctx context.Context, segmentID uuid.UUID) ([]models.Bid, error)
	UpdateBidStatus(ctx context.Context, id uuid.UUID, status models.BidStatus) error
	ExpireOldBids(ctx context.Context) (int, error)
	CountActiveBidsForCategories(ctx context.Context, categories []string) (int, error)
	CreateCreditTransaction(ctx context.Context, tx *models.CreditTransaction) error
	GetCreditHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.CreditTransaction, int, error)
	CreatePriceHistory(ctx context.Context, ph *models.PriceHistory) error
}
