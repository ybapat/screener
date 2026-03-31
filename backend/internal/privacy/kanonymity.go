package privacy

import (
	"github.com/google/uuid"
)

// QuasiIdentifier represents the set of attributes used for k-anonymity grouping.
// Each combination of these values defines an equivalence class.
type QuasiIdentifier struct {
	AppCategory   string `json:"app_category"`
	TimeOfDay     string `json:"time_of_day"`
	DurationRange string `json:"duration_range"`
	AgeRange      string `json:"age_range"`
	Country       string `json:"country"`
	DeviceType    string `json:"device_type"`
}

// AnonymizedRecord is a screen time record after generalization.
type AnonymizedRecord struct {
	QI          QuasiIdentifier `json:"qi"`
	DurationSec int             `json:"duration_secs"` // raw value kept for aggregation, not exported
	UserID      uuid.UUID       `json:"-"`             // never exported
}

// EquivalenceClass groups anonymized records sharing the same quasi-identifier.
type EquivalenceClass struct {
	QI      QuasiIdentifier    `json:"qi"`
	Records []AnonymizedRecord `json:"-"`
	UserIDs map[uuid.UUID]bool `json:"-"` // distinct contributors
}

// KAnonymizer groups records by quasi-identifiers and suppresses groups
// with fewer than k distinct contributors.
type KAnonymizer struct {
	K int
}

// NewKAnonymizer creates a new k-anonymity engine with the given k value.
func NewKAnonymizer(k int) *KAnonymizer {
	if k < 2 {
		k = 2
	}
	return &KAnonymizer{K: k}
}

// Anonymize takes a set of anonymized records and groups them into equivalence classes.
// Returns two slices: kept classes (>= k distinct users) and suppressed records.
func (a *KAnonymizer) Anonymize(records []AnonymizedRecord) (kept []EquivalenceClass, suppressed []AnonymizedRecord) {
	classMap := make(map[QuasiIdentifier]*EquivalenceClass)

	for _, rec := range records {
		ec, exists := classMap[rec.QI]
		if !exists {
			ec = &EquivalenceClass{
				QI:      rec.QI,
				UserIDs: make(map[uuid.UUID]bool),
			}
			classMap[rec.QI] = ec
		}
		ec.Records = append(ec.Records, rec)
		ec.UserIDs[rec.UserID] = true
	}

	for _, ec := range classMap {
		if len(ec.UserIDs) >= a.K {
			kept = append(kept, *ec)
		} else {
			suppressed = append(suppressed, ec.Records...)
		}
	}

	return kept, suppressed
}

// ContributorCount returns the number of distinct users in an equivalence class.
func (ec *EquivalenceClass) ContributorCount() int {
	return len(ec.UserIDs)
}

// TotalDuration returns the sum of all durations in the equivalence class.
func (ec *EquivalenceClass) TotalDuration() int {
	sum := 0
	for _, r := range ec.Records {
		sum += r.DurationSec
	}
	return sum
}

// MeanDuration returns the average duration in the equivalence class.
func (ec *EquivalenceClass) MeanDuration() float64 {
	if len(ec.Records) == 0 {
		return 0
	}
	return float64(ec.TotalDuration()) / float64(len(ec.Records))
}
