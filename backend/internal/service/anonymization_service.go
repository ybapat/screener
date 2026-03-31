package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/privacy"
	"github.com/ybapat/screener/backend/internal/repository"
)

type AnonymizationService struct {
	screenTime    repository.ScreenTimeRepository
	users         repository.UserRepository
	datasets      repository.DatasetRepository
	budgetTracker *privacy.BudgetTracker
}

func NewAnonymizationService(
	st repository.ScreenTimeRepository,
	users repository.UserRepository,
	ds repository.DatasetRepository,
	bt *privacy.BudgetTracker,
) *AnonymizationService {
	return &AnonymizationService{
		screenTime:    st,
		users:         users,
		datasets:      ds,
		budgetTracker: bt,
	}
}

type AssembleParams struct {
	Title      string   `json:"title"`
	Categories []string `json:"categories"`
	K          int      `json:"k_anonymity_k"`
	Epsilon    float64  `json:"epsilon"`
}

type AssembleResult struct {
	Dataset          *models.Dataset        `json:"dataset"`
	ClassCount       int                    `json:"equivalence_classes"`
	SuppressedCount  int                    `json:"suppressed_records"`
	ContributorCount int                    `json:"contributor_count"`
	Aggregations     []ClassAggregation     `json:"aggregations"`
}

type ClassAggregation struct {
	QI          privacy.QuasiIdentifier `json:"quasi_identifier"`
	Aggregates  privacy.NoisyAggregates `json:"aggregates"`
	Contributors int                    `json:"contributors"`
}

// AssembleDataset runs the full anonymization pipeline:
// 1. Fetch eligible records
// 2. Check privacy budgets
// 3. Generalize quasi-identifiers
// 4. Apply k-anonymity
// 5. Add differential privacy noise
// 6. Debit epsilon budgets
// 7. Create dataset record
func (s *AnonymizationService) AssembleDataset(ctx context.Context, params AssembleParams) (*AssembleResult, error) {
	if params.K < 2 {
		params.K = 5
	}
	if params.Epsilon <= 0 || params.Epsilon > 5.0 {
		params.Epsilon = 1.0
	}

	// Step 1: Fetch available records
	records, err := s.screenTime.GetAvailableRecords(ctx, params.Categories, params.K)
	if err != nil {
		return nil, fmt.Errorf("fetch records: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no eligible records found for categories %v", params.Categories)
	}

	// Step 2: Filter by privacy budget — only include users who can afford the epsilon cost
	type userInfo struct {
		ageRange string
		country  string
	}
	eligibleUsers := make(map[uuid.UUID]userInfo)
	for _, rec := range records {
		if _, checked := eligibleUsers[rec.UserID]; checked {
			continue
		}
		canSpend, _, err := s.budgetTracker.CanSpend(ctx, rec.UserID, params.Epsilon)
		if err != nil {
			continue
		}
		if canSpend {
			user, err := s.users.GetByID(ctx, rec.UserID)
			if err != nil || user == nil {
				continue
			}
			info := userInfo{}
			if user.AgeRange != nil {
				info.ageRange = *user.AgeRange
			}
			if user.Country != nil {
				info.country = *user.Country
			}
			eligibleUsers[rec.UserID] = info
		}
	}

	// Step 3: Generalize records
	var anonymized []privacy.AnonymizedRecord
	for _, rec := range records {
		info, eligible := eligibleUsers[rec.UserID]
		if !eligible {
			continue
		}

		anonymized = append(anonymized, privacy.AnonymizedRecord{
			QI: privacy.QuasiIdentifier{
				AppCategory:   privacy.GeneralizeAppName(rec.AppName),
				TimeOfDay:     privacy.GeneralizeTimestamp(rec.StartedAt),
				DurationRange: privacy.GeneralizeDuration(rec.DurationSec),
				AgeRange:      privacy.GeneralizeAge(info.ageRange),
				Country:       info.country,
				DeviceType:    stringVal(rec.DeviceType),
			},
			DurationSec: rec.DurationSec,
			UserID:      rec.UserID,
		})
	}

	if len(anonymized) == 0 {
		return nil, fmt.Errorf("no records passed eligibility filter")
	}

	// Step 4: Apply k-anonymity
	kanon := privacy.NewKAnonymizer(params.K)
	keptClasses, suppressed := kanon.Anonymize(anonymized)

	if len(keptClasses) == 0 {
		return nil, fmt.Errorf("all equivalence classes were suppressed (insufficient diversity for k=%d)", params.K)
	}

	// Step 5: Apply differential privacy noise
	dpAgg := privacy.NewDPAggregator(params.Epsilon, 86400, 100)
	var aggregations []ClassAggregation
	allContributors := make(map[uuid.UUID]int) // userID -> records included

	for _, ec := range keptClasses {
		agg := dpAgg.AggregateClass(ec)
		aggregations = append(aggregations, ClassAggregation{
			QI:           ec.QI,
			Aggregates:   agg,
			Contributors: ec.ContributorCount(),
		})
		for uid := range ec.UserIDs {
			allContributors[uid] += len(ec.Records)
		}
	}

	// Step 6: Debit epsilon budgets
	datasetID := uuid.New()
	for uid := range allContributors {
		desc := fmt.Sprintf("dataset %s assembly", datasetID.String()[:8])
		if err := s.budgetTracker.Spend(ctx, uid, params.Epsilon, &datasetID, desc); err != nil {
			// User couldn't afford it after all (race condition) — skip them
			delete(allContributors, uid)
		}
	}

	// Step 7: Create dataset record
	totalRecords := 0
	for _, ec := range keptClasses {
		totalRecords += len(ec.Records)
	}

	basePrice := int64(len(allContributors) * 100) // 100 credits per contributor
	dataset := &models.Dataset{
		ID:                  datasetID,
		Title:               params.Title,
		CategoryFilter:      params.Categories,
		ContributorCount:    len(allContributors),
		RecordCount:         totalRecords,
		KAnonymityK:         params.K,
		EpsilonPerQuery:     params.Epsilon,
		NoiseMechanism:      "laplace",
		BasePriceCredits:    basePrice,
		CurrentPriceCredits: basePrice,
		Status:              models.DatasetStatusActive,
	}

	if err := s.datasets.Create(ctx, dataset); err != nil {
		return nil, fmt.Errorf("create dataset: %w", err)
	}

	// Record contributors
	for uid, recCount := range allContributors {
		earning := basePrice / int64(len(allContributors))
		s.datasets.AddContributor(ctx, &models.DatasetContributor{
			ID:              uuid.New(),
			DatasetID:       datasetID,
			UserID:          uid,
			EpsilonCharged:  params.Epsilon,
			RecordsIncluded: recCount,
			EarningCredits:  earning,
		})
	}

	// Generate sample rows from the first few kept classes
	for i, agg := range aggregations {
		if i >= 10 {
			break
		}
		s.datasets.CreateSample(ctx, &models.DatasetSample{
			ID:            uuid.New(),
			DatasetID:     datasetID,
			AppCategory:   agg.QI.AppCategory,
			DurationRange: agg.QI.DurationRange,
			TimeOfDay:     agg.QI.TimeOfDay,
		})
	}

	return &AssembleResult{
		Dataset:          dataset,
		ClassCount:       len(keptClasses),
		SuppressedCount:  len(suppressed),
		ContributorCount: len(allContributors),
		Aggregations:     aggregations,
	}, nil
}

func stringVal(s *string) string {
	if s == nil {
		return "unknown"
	}
	return *s
}
