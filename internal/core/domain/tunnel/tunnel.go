package tunnel

import (
	"errors"
	"strings"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
)

var (
	ErrEmptyTunnelID   = errors.New("tunnel ID cannot be empty")
	ErrEmptyTunnelName = errors.New("tunnel name cannot be empty")
)

type TunnelID string

func (id TunnelID) String() string {
	return string(id)
}

func (id TunnelID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyTunnelID
	}
	return nil
}

type TunnelStatus string

const (
	StatusHealthy  TunnelStatus = "healthy"
	StatusDegraded TunnelStatus = "degraded"
	StatusDown     TunnelStatus = "down"
	StatusInactive TunnelStatus = "inactive"
)

type Connector struct {
	ID            string     `json:"id"`
	ClientVersion string     `json:"client_version"`
	ColoName      string     `json:"colo_name"`
	OpenedAt      *time.Time `json:"opened_at,omitempty"`
	OriginIP      string     `json:"origin_ip"`
	Status        string     `json:"status"`
}

type IngressRule struct {
	Hostname string `json:"hostname"`
	Path     string `json:"path,omitempty"`
	Service  string `json:"service"`
}

type Tunnel struct {
	ID            TunnelID          `json:"id"`
	AccountID     account.AccountID `json:"account_id"`
	Name          string            `json:"name"`
	Status        TunnelStatus      `json:"status"`
	ConnsActiveAt *time.Time        `json:"conns_active_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	RemoteConfig  bool              `json:"remote_config"`
	Connectors    []Connector       `json:"connectors"`
	IngressRules  []IngressRule     `json:"ingress_rules"`
}

func NewTunnel(
	id string,
	accountID account.AccountID,
	name string,
	status TunnelStatus,
	connsActiveAt *time.Time,
	createdAt time.Time,
	remoteConfig bool,
	connectors []Connector,
	ingressRules []IngressRule,
) (Tunnel, error) {
	tID := TunnelID(strings.TrimSpace(id))
	if err := tID.Validate(); err != nil {
		return Tunnel{}, err
	}
	if err := accountID.Validate(); err != nil {
		return Tunnel{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Tunnel{}, ErrEmptyTunnelName
	}
	if status == "" {
		status = DeriveTunnelStatus(connectors)
	}

	return Tunnel{
		ID:            tID,
		AccountID:     accountID,
		Name:          name,
		Status:        status,
		ConnsActiveAt: connsActiveAt,
		CreatedAt:     createdAt,
		RemoteConfig:  remoteConfig,
		Connectors:    connectors,
		IngressRules:  ingressRules,
	}, nil
}

func DeriveTunnelStatus(connectors []Connector) TunnelStatus {
	if len(connectors) == 0 {
		return StatusDown
	}

	activeCount := 0
	for _, c := range connectors {
		st := strings.ToLower(strings.TrimSpace(c.Status))
		if st == "connected" || st == "healthy" || st == "active" {
			activeCount++
		}
	}

	if activeCount == len(connectors) && activeCount > 0 {
		return StatusHealthy
	}
	if activeCount > 0 {
		return StatusDegraded
	}
	return StatusDown
}

type HealthTally struct {
	Total    int
	Healthy  int
	Degraded int
	Down     int
	Inactive int
}

func CalculateHealthTally(tunnels []Tunnel) HealthTally {
	var tally HealthTally
	tally.Total = len(tunnels)
	for _, t := range tunnels {
		switch t.Status {
		case StatusHealthy:
			tally.Healthy++
		case StatusDegraded:
			tally.Degraded++
		case StatusDown:
			tally.Down++
		case StatusInactive:
			tally.Inactive++
		default:
			tally.Down++
		}
	}
	return tally
}
