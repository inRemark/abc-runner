package utils

import (
	"testing"
	"time"
)

func TestCalculateStatistics(t *testing.T) {
	// Test with sample data
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		150 * time.Millisecond,
		300 * time.Millisecond,
		250 * time.Millisecond,
	}

	metrics := CalculateStatistics(durations)

	// Check that all metrics are calculated
	if metrics.Avg == 0 {
		t.Error("Average should not be zero")
	}
	if metrics.Min == 0 {
		t.Error("Min should not be zero")
	}
	if metrics.Max == 0 {
		t.Error("Max should not be zero")
	}
	if metrics.P90 == 0 {
		t.Error("P90 should not be zero")
	}
	if metrics.P95 == 0 {
		t.Error("P95 should not be zero")
	}
	if metrics.P99 == 0 {
		t.Error("P99 should not be zero")
	}

	// Check specific values
	expectedMin := 100 * time.Millisecond
	if metrics.Min != expectedMin {
		t.Errorf("Min = %v, expected %v", metrics.Min, expectedMin)
	}

	expectedMax := 300 * time.Millisecond
	if metrics.Max != expectedMax {
		t.Errorf("Max = %v, expected %v", metrics.Max, expectedMax)
	}

	// Test with empty slice
	emptyMetrics := CalculateStatistics([]time.Duration{})
	if emptyMetrics.Avg != 0 || emptyMetrics.Min != 0 || emptyMetrics.Max != 0 {
		t.Error("Metrics should be zero for empty slice")
	}
}

func TestCalculateStatisticsSingleValue(t *testing.T) {
	// Test with single value
	durations := []time.Duration{100 * time.Millisecond}
	metrics := CalculateStatistics(durations)

	// All percentiles should be the same for single value
	if metrics.Min != metrics.Max || metrics.Max != metrics.P90 ||
		metrics.P90 != metrics.P95 || metrics.P95 != metrics.P99 {
		t.Error("All metrics should be equal for single value")
	}

	expected := 100 * time.Millisecond
	if metrics.Min != expected {
		t.Errorf("Min = %v, expected %v", metrics.Min, expected)
	}
}

func TestCalculateStatisticsSortedOrder(t *testing.T) {
	// Test that the function correctly sorts the input
	durations := []time.Duration{
		300 * time.Millisecond, // Max
		100 * time.Millisecond, // Min
		200 * time.Millisecond, // Middle
	}

	metrics := CalculateStatistics(durations)

	expectedMin := 100 * time.Millisecond
	expectedMax := 300 * time.Millisecond

	if metrics.Min != expectedMin {
		t.Errorf("Min = %v, expected %v", metrics.Min, expectedMin)
	}

	if metrics.Max != expectedMax {
		t.Errorf("Max = %v, expected %v", metrics.Max, expectedMax)
	}
}