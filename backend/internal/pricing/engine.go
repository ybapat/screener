package pricing

import (
	"context"
	"math"

	"github.com/ybapat/screener/backend/internal/repository"
)

// Engine computes dynamic prices for datasets based on supply, demand, and quality.
type Engine struct {
	marketplace repository.MarketplaceRepository
	screenTime  repository.ScreenTimeRepository
}

// NewEngine creates a new pricing engine.
func NewEngine(mp repository.MarketplaceRepository, st repository.ScreenTimeRepository) *Engine {
	return &Engine{marketplace: mp, screenTime: st}
}

// Params holds the inputs to the pricing formula.
type Params struct {
	BasePrice        int64
	AppCategories    []string
	RecordCount      int
	ContributorCount int
	KAnonymity       int
	Epsilon          float64
}

// ComputePrice calculates the dynamic price using:
//
//	price = base * rarity * demand * quality
//
// Rarity: inverse of supply ratio (rarer data = more expensive)
// Demand: based on active bids for similar categories
// Quality: higher k-anonymity and lower epsilon = more private = more valuable
func (e *Engine) ComputePrice(ctx context.Context, params Params) (int64, float64, float64) {
	// Rarity multiplier
	rarityScore := 1.0
	availableRecords, err := e.screenTime.GetAvailableRecords(ctx, params.AppCategories, 1)
	if err == nil && len(availableRecords) > 0 {
		supplyRatio := float64(params.RecordCount) / float64(len(availableRecords))
		rarityScore = 1 + math.Log1p(1.0/math.Max(supplyRatio, 0.01))
	}

	// Demand multiplier
	demandScore := 1.0
	activeBids, err := e.marketplace.CountActiveBidsForCategories(ctx, params.AppCategories)
	if err == nil {
		demandScore = 1 + 0.1*float64(activeBids)
	}

	// Quality multiplier: higher k and lower epsilon = more valuable
	qualityMultiplier := 1.0 + 0.2*float64(params.KAnonymity)/5.0 + 0.3*(1.0/math.Max(params.Epsilon, 0.1))

	finalPrice := float64(params.BasePrice) * rarityScore * demandScore * qualityMultiplier
	return int64(math.Round(finalPrice)), demandScore, rarityScore
}
