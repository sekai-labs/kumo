package zone

import (
	"errors"
	"strings"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
)

var (
	ErrEmptyZoneID   = errors.New("zone ID cannot be empty")
	ErrEmptyZoneName = errors.New("zone name cannot be empty")
)

type ZoneID string

func (id ZoneID) String() string {
	return string(id)
}

func (id ZoneID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyZoneID
	}
	return nil
}

type ZoneStatus string

const (
	ZoneStatusActive       ZoneStatus = "active"
	ZoneStatusPending      ZoneStatus = "pending"
	ZoneStatusInitializing ZoneStatus = "initializing"
	ZoneStatusMoved        ZoneStatus = "moved"
	ZoneStatusDeleted      ZoneStatus = "deleted"
	ZoneStatusDeactivated  ZoneStatus = "deactivated"
)

type Zone struct {
	ID          ZoneID            `json:"id"`
	Name        string            `json:"name"`
	Status      ZoneStatus        `json:"status"`
	Plan        string            `json:"plan"`
	AccountID   account.AccountID `json:"account_id"`
	AccountName string            `json:"account_name"`
	NameServers []string          `json:"name_servers"`
	Paused      bool              `json:"paused"`
}

func NewZone(id, name, status, plan string, accID account.AccountID, accName string, nameServers []string, paused bool) (Zone, error) {
	zID := ZoneID(strings.TrimSpace(id))
	if err := zID.Validate(); err != nil {
		return Zone{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Zone{}, ErrEmptyZoneName
	}
	if status == "" {
		status = string(ZoneStatusActive)
	}
	if plan == "" {
		plan = "Free"
	}
	return Zone{
		ID:          zID,
		Name:        name,
		Status:      ZoneStatus(status),
		Plan:        plan,
		AccountID:   accID,
		AccountName: accName,
		NameServers: nameServers,
		Paused:      paused,
	}, nil
}
