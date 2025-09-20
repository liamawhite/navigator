// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"math"
	"sort"
	"time"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/prometheus/prometheus/promql"
	"google.golang.org/protobuf/types/known/durationpb"
)

// CalculateQuantile calculates a quantile from a histogram distribution using Prometheus's algorithm
func CalculateQuantile(q float64, dist *typesv1alpha1.LatencyDistribution) (float64, error) {
	if dist == nil || len(dist.Buckets) == 0 {
		return math.NaN(), fmt.Errorf("empty distribution")
	}

	// Convert our protobuf histogram to Prometheus bucket format
	buckets, err := convertToPrometheusBuckets(dist)
	if err != nil {
		return math.NaN(), fmt.Errorf("failed to convert histogram: %w", err)
	}

	// Use Prometheus's bucket quantile calculation
	result, _, _ := promql.BucketQuantile(q, buckets)

	return result, nil
}

// AggregateHistograms merges multiple histogram distributions into a single distribution
func AggregateHistograms(distributions []*typesv1alpha1.LatencyDistribution) (*typesv1alpha1.LatencyDistribution, error) {
	if len(distributions) == 0 {
		return nil, fmt.Errorf("no distributions to aggregate")
	}

	if len(distributions) == 1 {
		return distributions[0], nil
	}

	// Collect all unique bucket boundaries (le values)
	bucketMap := make(map[float64]float64) // le -> count
	var totalCount float64
	var totalSum float64

	for _, dist := range distributions {
		if dist == nil {
			continue
		}

		totalCount += dist.TotalCount
		totalSum += dist.Sum

		// Aggregate bucket counts
		for _, bucket := range dist.Buckets {
			bucketMap[bucket.Le] += bucket.Count
		}
	}

	// Convert back to sorted buckets
	var buckets []*typesv1alpha1.HistogramBucket
	var les []float64
	for le := range bucketMap {
		les = append(les, le)
	}
	sort.Float64s(les)

	for _, le := range les {
		buckets = append(buckets, &typesv1alpha1.HistogramBucket{
			Le:    le,
			Count: bucketMap[le],
		})
	}

	return &typesv1alpha1.LatencyDistribution{
		Buckets:    buckets,
		TotalCount: totalCount,
		Sum:        totalSum,
	}, nil
}

// convertToPrometheusBuckets converts our protobuf histogram to Prometheus bucket format
func convertToPrometheusBuckets(dist *typesv1alpha1.LatencyDistribution) (promql.Buckets, error) {
	if len(dist.Buckets) == 0 {
		return nil, fmt.Errorf("no buckets in distribution")
	}

	// Sort buckets by upper bound to ensure proper order
	sortedBuckets := make([]*typesv1alpha1.HistogramBucket, len(dist.Buckets))
	copy(sortedBuckets, dist.Buckets)
	sort.Slice(sortedBuckets, func(i, j int) bool {
		return sortedBuckets[i].Le < sortedBuckets[j].Le
	})

	// Convert to Prometheus bucket format
	var buckets promql.Buckets

	for _, bucket := range sortedBuckets {
		buckets = append(buckets, promql.Bucket{
			UpperBound: bucket.Le,
			Count:      bucket.Count,
		})
	}

	// Ensure the last bucket has infinite upper bound as required by Prometheus
	if len(buckets) > 0 && !math.IsInf(buckets[len(buckets)-1].UpperBound, 1) {
		// Add +Inf bucket with same count as last bucket (cumulative nature)
		buckets = append(buckets, promql.Bucket{
			UpperBound: math.Inf(1),
			Count:      buckets[len(buckets)-1].Count,
		})
	}

	return buckets, nil
}

// CalculateP99 is a convenience function to calculate the 99th percentile
func CalculateP99(dist *typesv1alpha1.LatencyDistribution) (float64, error) {
	return CalculateQuantile(0.99, dist)
}

// CalculateP95 is a convenience function to calculate the 95th percentile
func CalculateP95(dist *typesv1alpha1.LatencyDistribution) (float64, error) {
	return CalculateQuantile(0.95, dist)
}

// CalculateP50 is a convenience function to calculate the 50th percentile (median)
func CalculateP50(dist *typesv1alpha1.LatencyDistribution) (float64, error) {
	return CalculateQuantile(0.50, dist)
}

// CalculateP99AsDuration calculates P99 and returns it as a protobuf Duration
func CalculateP99AsDuration(dist *typesv1alpha1.LatencyDistribution) (*durationpb.Duration, error) {
	p99Ms, err := CalculateP99(dist)
	if err != nil || math.IsNaN(p99Ms) {
		return nil, err
	}

	// Convert milliseconds to duration
	return durationpb.New(time.Duration(p99Ms * float64(time.Millisecond))), nil
}

// CalculateQuantileAsDuration calculates any quantile and returns it as a protobuf Duration
func CalculateQuantileAsDuration(q float64, dist *typesv1alpha1.LatencyDistribution) (*durationpb.Duration, error) {
	quantileMs, err := CalculateQuantile(q, dist)
	if err != nil || math.IsNaN(quantileMs) {
		return nil, err
	}

	// Convert milliseconds to duration
	return durationpb.New(time.Duration(quantileMs * float64(time.Millisecond))), nil
}
