package account

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyAccountID   = errors.New("account ID cannot be empty")
	ErrEmptyAccountName = errors.New("account name cannot be empty")
	ErrInvalidTokenInfo = errors.New("invalid token info")
)

type AccountID string

func (id AccountID) String() string {
	return string(id)
}

func (id AccountID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyAccountID
	}
	return nil
}

type Account struct {
	ID        AccountID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(id string, name string, accType string, createdAt time.Time) (Account, error) {
	accID := AccountID(strings.TrimSpace(id))
	if err := accID.Validate(); err != nil {
		return Account{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Account{}, ErrEmptyAccountName
	}
	if accType == "" {
		accType = "standard"
	}
	return Account{
		ID:        accID,
		Name:      name,
		Type:      accType,
		CreatedAt: createdAt,
	}, nil
}

type TokenStatus string

const (
	TokenStatusActive   TokenStatus = "active"
	TokenStatusDisabled TokenStatus = "disabled"
	TokenStatusExpired  TokenStatus = "expired"
)

type TokenInfo struct {
	ID        string      `json:"id"`
	Status    TokenStatus `json:"status"`
	ExpiresOn *time.Time  `json:"expires_on,omitempty"`
}

func (t TokenInfo) IsActive() bool {
	if t.Status != TokenStatusActive {
		return false
	}
	if t.ExpiresOn != nil && time.Now().After(*t.ExpiresOn) {
		return false
	}
	return true
}
