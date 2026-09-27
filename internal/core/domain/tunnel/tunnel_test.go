package tunnel_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
)

func TestNewTunnel_Valid(t *testing.T) {
	now := time.Now()
	connectors := []tunnel.Connector{
		{
			ID:            "conn-1",
			ClientVersion: "2024.1.0",
			ColoName:      "SJC",
			OpenedAt:      &now,
			OriginIP:      "192.0.2.10",
			Status:        "connected",
		},
	}
	rules := []tunnel.IngressRule{
		{
			Hostname: "app.example.com",
			Path:     "/api",
			Service:  "http://localhost:8080",
		},
	}

	tun, err := tunnel.NewTunnel(
		"tun-123",
		account.AccountID("acc-1"),
		"prod-tunnel",
		tunnel.StatusHealthy,
		&now,
		now,
		true,
		connectors,
		rules,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tun.ID.String() != "tun-123" {
		t.Errorf("expected ID tun-123, got %s", tun.ID)
	}
	if tun.Status != tunnel.StatusHealthy {
		t.Errorf("expected healthy, got %s", tun.Status)
	}
	if len(tun.Connectors) != 1 {
		t.Errorf("expected 1 connector, got %d", len(tun.Connectors))
	}
	if len(tun.IngressRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(tun.IngressRules))
	}
}

func TestNewTunnel_Invalid(t *testing.T) {
	now := time.Now()

	t.Run("empty id", func(t *testing.T) {
		_, err := tunnel.NewTunnel("", "acc-1", "name", tunnel.StatusHealthy, nil, now, false, nil, nil)
		if !errors.Is(err, tunnel.ErrEmptyTunnelID) {
			t.Errorf("expected ErrEmptyTunnelID, got %v", err)
		}
	})

	t.Run("empty account id", func(t *testing.T) {
		_, err := tunnel.NewTunnel("tun-1", "", "name", tunnel.StatusHealthy, nil, now, false, nil, nil)
		if !errors.Is(err, account.ErrEmptyAccountID) {
			t.Errorf("expected ErrEmptyAccountID, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := tunnel.NewTunnel("tun-1", "acc-1", "  ", tunnel.StatusHealthy, nil, now, false, nil, nil)
		if !errors.Is(err, tunnel.ErrEmptyTunnelName) {
			t.Errorf("expected ErrEmptyTunnelName, got %v", err)
		}
	})
}

func TestDeriveTunnelStatus(t *testing.T) {
	tests := []struct {
		name       string
		connectors []tunnel.Connector
		expected   tunnel.TunnelStatus
	}{
		{
			name:       "empty connectors",
			connectors: nil,
			expected:   tunnel.StatusDown,
		},
		{
			name: "all connected",
			connectors: []tunnel.Connector{
				{Status: "connected"},
				{Status: "healthy"},
			},
			expected: tunnel.StatusHealthy,
		},
		{
			name: "partially connected",
			connectors: []tunnel.Connector{
				{Status: "connected"},
				{Status: "disconnected"},
			},
			expected: tunnel.StatusDegraded,
		},
		{
			name: "all down",
			connectors: []tunnel.Connector{
				{Status: "disconnected"},
				{Status: "error"},
			},
			expected: tunnel.StatusDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tunnel.DeriveTunnelStatus(tt.connectors)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestCalculateHealthTally(t *testing.T) {
	tunnels := []tunnel.Tunnel{
		{Status: tunnel.StatusHealthy},
		{Status: tunnel.StatusHealthy},
		{Status: tunnel.StatusDegraded},
		{Status: tunnel.StatusDown},
		{Status: tunnel.StatusInactive},
	}

	tally := tunnel.CalculateHealthTally(tunnels)
	if tally.Total != 5 {
		t.Errorf("expected Total 5, got %d", tally.Total)
	}
	if tally.Healthy != 2 {
		t.Errorf("expected Healthy 2, got %d", tally.Healthy)
	}
	if tally.Degraded != 1 {
		t.Errorf("expected Degraded 1, got %d", tally.Degraded)
	}
	if tally.Down != 1 {
		t.Errorf("expected Down 1, got %d", tally.Down)
	}
	if tally.Inactive != 1 {
		t.Errorf("expected Inactive 1, got %d", tally.Inactive)
	}
}
