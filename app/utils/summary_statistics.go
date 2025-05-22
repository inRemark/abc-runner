package utils

import (
	"sort"
	"time"
)

type ClientMetrics struct {
	RPS int32
	Avg time.Duration
	Min time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
	Max time.Duration
}

func CalculateStatistics(durations []time.Duration) (metrics ClientMetrics) {
	n := len(durations)
	metrics = ClientMetrics{
		RPS: 0,
		Avg: 0,
		Min: 0,
		P90: 0,
		P95: 0,
		P99: 0,
		Max: 0,
	}
	if n == 0 {
		return metrics
	}

	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	metrics.Avg = sum / time.Duration(n)

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	getPercentile := func(p float64) time.Duration {
		pos := int(float64(n)*p + 0.5)
		if pos > n-1 {
			pos = n - 1
		}
		return durations[pos]
	}

	metrics.P90 = getPercentile(0.90)
	metrics.P95 = getPercentile(0.95)
	metrics.P99 = getPercentile(0.99)
	metrics.Min = durations[0]
	metrics.Max = durations[n-1]

	return metrics
}
