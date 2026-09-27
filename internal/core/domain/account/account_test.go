package account

import (
	"testing"
	"time"
)

func TestNewAccount(t *testing.T) {
	t.Run("valid account", func(t *testing.T) {
		acc, err := NewAccount("acc-123", "Test Account", "standard", time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc.ID.String() != "acc-123" {
			t.Errorf("expected acc-123, got %s", acc.ID)
		}
		if acc.Name != "Test Account" {
			t.Errorf("expected 'Test Account', got %s", acc.Name)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := NewAccount("", "Test Account", "standard", time.Now())
		if err != ErrEmptyAccountID {
			t.Errorf("expected ErrEmptyAccountID, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewAccount("acc-123", "   ", "standard", time.Now())
		if err != ErrEmptyAccountName {
			t.Errorf("expected ErrEmptyAccountName, got %v", err)
		}
	})
}

func TestTokenInfo_IsActive(t *testing.T) {
	activeToken := TokenInfo{ID: "tok-1", Status: TokenStatusActive}
	if !activeToken.IsActive() {
		t.Errorf("expected active token to be active")
	}

	disabledToken := TokenInfo{ID: "tok-2", Status: TokenStatusDisabled}
	if disabledToken.IsActive() {
		t.Errorf("expected disabled token not to be active")
	}

	past := time.Now().Add(-1 * time.Hour)
	expiredToken := TokenInfo{ID: "tok-3", Status: TokenStatusActive, ExpiresOn: &past}
	if expiredToken.IsActive() {
		t.Errorf("expected expired token not to be active")
	}

	future := time.Now().Add(1 * time.Hour)
	validExpToken := TokenInfo{ID: "tok-4", Status: TokenStatusActive, ExpiresOn: &future}
	if !validExpToken.IsActive() {
		t.Errorf("expected future expiring token to be active")
	}
}
