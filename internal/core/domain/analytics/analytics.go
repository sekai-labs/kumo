package analytics

import (
	"time"
)

type TrafficSummary struct {
	TotalRequests  int64            `json:"total_requests"`
	CachedRequests int64            `json:"cached_requests"`
	BandwidthBytes int64            `json:"bandwidth_bytes"`
	StatusCodes    map[string]int64 `json:"status_codes"`
	ThreatCount    int64            `json:"threat_count"`
	Timestamp      time.Time        `json:"timestamp"`
}

func (ts TrafficSummary) CacheHitRatio() float64 {
	if ts.TotalRequests <= 0 {
		return 0.0
	}
	ratio := (float64(ts.CachedRequests) / float64(ts.TotalRequests)) * 100.0
	if ratio > 100.0 {
		return 100.0
	}
	if ratio < 0.0 {
		return 0.0
	}
	return ratio
}

func (ts TrafficSummary) UncachedRequests() int64 {
	diff := ts.TotalRequests - ts.CachedRequests
	if diff < 0 {
		return 0
	}
	return diff
}

func (ts TrafficSummary) StatusGroup(prefix string) int64 {
	var count int64
	for code, val := range ts.StatusCodes {
		if len(code) > 0 && code[0] == prefix[0] {
			count += val
		}
	}
	return count
}
