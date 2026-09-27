package analytics_test

import (
	"math"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
)

func TestTrafficSummary_CacheHitRatio(t *testing.T) {
	tests := []struct {
		name     string
		summary  analytics.TrafficSummary
		expected float64
	}{
		{
			name: "50% cached",
			summary: analytics.TrafficSummary{
				TotalRequests:  1000,
				CachedRequests: 500,
			},
			expected: 50.0,
		},
		{
			name: "100% cached",
			summary: analytics.TrafficSummary{
				TotalRequests:  250,
				CachedRequests: 250,
			},
			expected: 100.0,
		},
		{
			name: "0% cached",
			summary: analytics.TrafficSummary{
				TotalRequests:  500,
				CachedRequests: 0,
			},
			expected: 0.0,
		},
		{
			name: "zero total requests",
			summary: analytics.TrafficSummary{
				TotalRequests:  0,
				CachedRequests: 0,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := tt.summary.CacheHitRatio()
			if math.Abs(ratio-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, ratio)
			}
		})
	}
}

func TestTrafficSummary_UncachedAndStatusGroup(t *testing.T) {
	ts := analytics.TrafficSummary{
		TotalRequests:  1000,
		CachedRequests: 600,
		BandwidthBytes: 52428800,
		StatusCodes: map[string]int64{
			"200": 800,
			"204": 50,
			"301": 20,
			"404": 100,
			"500": 30,
		},
		ThreatCount: 5,
		Timestamp:   time.Now(),
	}

	if ts.UncachedRequests() != 400 {
		t.Errorf("expected 400 uncached, got %d", ts.UncachedRequests())
	}

	if ts.StatusGroup("2xx") != 850 {
		t.Errorf("expected 850 2xx, got %d", ts.StatusGroup("2xx"))
	}
	if ts.StatusGroup("3xx") != 20 {
		t.Errorf("expected 20 3xx, got %d", ts.StatusGroup("3xx"))
	}
	if ts.StatusGroup("4xx") != 100 {
		t.Errorf("expected 100 4xx, got %d", ts.StatusGroup("4xx"))
	}
	if ts.StatusGroup("5xx") != 30 {
		t.Errorf("expected 30 5xx, got %d", ts.StatusGroup("5xx"))
	}
}
