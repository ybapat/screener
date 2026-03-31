package privacy

import (
	"math"
	"math/rand"
)

// NoisyAggregates holds the differentially private aggregation results
// for a single equivalence class.
type NoisyAggregates struct {
	NoisyCount        int64   `json:"noisy_count"`
	NoisyMeanDuration float64 `json:"noisy_mean_duration_secs"`
	NoisySumDuration  float64 `json:"noisy_sum_duration_secs"`
}

// DPAggregator computes differentially private aggregations using the Laplace mechanism.
// This is a pure-Go implementation that doesn't depend on the Google DP library,
// making it easier to build and deploy while still providing formal DP guarantees.
type DPAggregator struct {
	// TotalEpsilon is the privacy budget for this aggregation.
	// Split equally across count, mean, and sum queries.
	TotalEpsilon float64
	// MaxDuration is the upper bound for duration clamping (sensitivity bound).
	MaxDuration float64
	// MaxRecordsPerUser caps how many records one user can contribute.
	MaxRecordsPerUser int
}

// NewDPAggregator creates a new aggregator with the given privacy parameters.
func NewDPAggregator(epsilon float64, maxDuration float64, maxRecordsPerUser int) *DPAggregator {
	return &DPAggregator{
		TotalEpsilon:      epsilon,
		MaxDuration:       maxDuration,
		MaxRecordsPerUser: maxRecordsPerUser,
	}
}

// AggregateClass computes noisy count, mean, and sum for an equivalence class.
// Uses the Laplace mechanism with sensitivity calibrated to the contribution bounds.
func (d *DPAggregator) AggregateClass(ec EquivalenceClass) NoisyAggregates {
	// Split epsilon three ways: count, mean, sum
	epsPerQuery := d.TotalEpsilon / 3.0

	// --- Noisy Count ---
	// Sensitivity of count = MaxRecordsPerUser (one user can add at most this many records)
	countSensitivity := float64(d.MaxRecordsPerUser)
	trueCount := float64(len(ec.Records))
	noisyCount := trueCount + laplaceSample(countSensitivity/epsPerQuery)
	if noisyCount < 0 {
		noisyCount = 0
	}

	// --- Clamp durations ---
	var clampedDurations []float64
	for _, r := range ec.Records {
		v := float64(r.DurationSec)
		if v > d.MaxDuration {
			v = d.MaxDuration
		}
		if v < 0 {
			v = 0
		}
		clampedDurations = append(clampedDurations, v)
	}

	// --- Noisy Sum ---
	// Sensitivity of sum = MaxRecordsPerUser * MaxDuration
	sumSensitivity := float64(d.MaxRecordsPerUser) * d.MaxDuration
	trueSum := 0.0
	for _, v := range clampedDurations {
		trueSum += v
	}
	noisySum := trueSum + laplaceSample(sumSensitivity/epsPerQuery)
	if noisySum < 0 {
		noisySum = 0
	}

	// --- Noisy Mean ---
	// Compute mean from noisy sum / noisy count to maintain consistency
	noisyMean := 0.0
	if noisyCount > 0 {
		noisyMean = noisySum / noisyCount
	}

	return NoisyAggregates{
		NoisyCount:        int64(math.Round(noisyCount)),
		NoisyMeanDuration: math.Round(noisyMean*100) / 100,
		NoisySumDuration:  math.Round(noisySum*100) / 100,
	}
}

// laplaceSample generates a sample from the Laplace distribution with
// location 0 and scale b. The Laplace mechanism adds noise calibrated
// to sensitivity/epsilon, achieving epsilon-differential privacy.
func laplaceSample(scale float64) float64 {
	u := rand.Float64() - 0.5
	if u == 0 {
		return 0
	}
	if u > 0 {
		return -scale * math.Log(1-2*u)
	}
	return scale * math.Log(1+2*u)
}
