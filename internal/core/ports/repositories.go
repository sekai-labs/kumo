package ports

import (
	"context"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
)

type DNSFilter struct {
	Type    dns.RecordType
	Name    string
	Content string
}

type AccountRepository interface {
	List(ctx context.Context) ([]account.Account, error)
	VerifyToken(ctx context.Context) (account.TokenInfo, error)
}

type ZoneRepository interface {
	List(ctx context.Context, accountID account.AccountID) ([]zone.Zone, error)
	Get(ctx context.Context, zoneID zone.ZoneID) (zone.Zone, error)
}

type DNSRepository interface {
	List(ctx context.Context, zoneID zone.ZoneID, filter DNSFilter) ([]dns.DNSRecord, error)
	Get(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error)
	Create(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error)
	Update(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error)
	Delete(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) error
	ToggleProxy(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error)
}

type TunnelRepository interface {
	List(ctx context.Context, accountID account.AccountID) ([]tunnel.Tunnel, error)
	Get(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) (tunnel.Tunnel, error)
	GetConfiguration(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) ([]tunnel.IngressRule, error)
}

type WorkerRepository interface {
	List(ctx context.Context, accountID account.AccountID) ([]worker.WorkerScript, error)
}

type PagesRepository interface {
	List(ctx context.Context, accountID account.AccountID) ([]pages.PagesProject, error)
}

type RulesetRepository interface {
	ListZoneRulesets(ctx context.Context, zoneID zone.ZoneID) ([]ruleset.Ruleset, error)
}

type AnalyticsRepository interface {
	GetZoneTrafficSummary(ctx context.Context, zoneID zone.ZoneID, since time.Duration) (analytics.TrafficSummary, error)
}
