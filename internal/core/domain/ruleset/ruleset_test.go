package ruleset_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
)

func TestNewRuleset(t *testing.T) {
	now := time.Now()
	t.Run("valid ruleset", func(t *testing.T) {
		rs, err := ruleset.NewRuleset(
			"rs-1",
			"custom WAF rules",
			"http_request_firewall_custom",
			"zone",
			now,
			"Block suspicious traffic",
			5,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rs.ID.String() != "rs-1" {
			t.Errorf("expected ID rs-1, got %s", rs.ID)
		}
		if rs.RulesCount != 5 {
			t.Errorf("expected 5 rules, got %d", rs.RulesCount)
		}
		if rs.Description != "Block suspicious traffic" {
			t.Errorf("expected description, got %s", rs.Description)
		}
	})

	t.Run("default phase and kind", func(t *testing.T) {
		rs, err := ruleset.NewRuleset("rs-2", "rate limiting", "", "", now, "", 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rs.Phase != "http_request_firewall_custom" {
			t.Errorf("expected default phase, got %s", rs.Phase)
		}
		if rs.Kind != "zone" {
			t.Errorf("expected default kind zone, got %s", rs.Kind)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := ruleset.NewRuleset("  ", "ruleset", "phase", "zone", now, "", 1)
		if !errors.Is(err, ruleset.ErrEmptyRulesetID) {
			t.Errorf("expected ErrEmptyRulesetID, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := ruleset.NewRuleset("rs-1", "  ", "phase", "zone", now, "", 1)
		if !errors.Is(err, ruleset.ErrEmptyRulesetName) {
			t.Errorf("expected ErrEmptyRulesetName, got %v", err)
		}
	})
}
