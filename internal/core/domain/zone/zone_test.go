package zone

import (
	"testing"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
)

func TestNewZone(t *testing.T) {
	accID := account.AccountID("acc-123")

	t.Run("valid zone", func(t *testing.T) {
		z, err := NewZone("z-1", "example.com", "active", "Pro", accID, "Main Acc", []string{"ns1.cloudflare.com"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if z.ID.String() != "z-1" {
			t.Errorf("expected z-1, got %s", z.ID)
		}
		if z.Name != "example.com" {
			t.Errorf("expected example.com, got %s", z.Name)
		}
		if z.Status != ZoneStatusActive {
			t.Errorf("expected active, got %s", z.Status)
		}
	})

	t.Run("empty zone id", func(t *testing.T) {
		_, err := NewZone("", "example.com", "active", "Pro", accID, "Main Acc", nil, false)
		if err != ErrEmptyZoneID {
			t.Errorf("expected ErrEmptyZoneID, got %v", err)
		}
	})

	t.Run("empty zone name", func(t *testing.T) {
		_, err := NewZone("z-1", "  ", "active", "Pro", accID, "Main Acc", nil, false)
		if err != ErrEmptyZoneName {
			t.Errorf("expected ErrEmptyZoneName, got %v", err)
		}
	})
}
