package testhelpers

import (
	"context"
	"time"

	"github.com/sekai-labs/kumo/internal/core/application"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/core/ports"
	"github.com/sekai-labs/kumo/internal/infrastructure/cache"
)

type StubAdapter struct{}

func (s *StubAdapter) List(ctx context.Context) ([]account.Account, error) {
	return []account.Account{
		{
			ID:        "acc-1",
			Name:      "Test Account",
			Type:      "standard",
			CreatedAt: time.Now(),
		},
	}, nil
}

func (s *StubAdapter) VerifyToken(ctx context.Context) (account.TokenInfo, error) {
	return account.TokenInfo{
		ID:     "tok-1",
		Status: account.TokenStatusActive,
	}, nil
}

type StubZoneAdapter struct{}

func (s *StubZoneAdapter) List(ctx context.Context, accountID account.AccountID) ([]zone.Zone, error) {
	return []zone.Zone{
		{
			ID:          "zone-1",
			Name:        "example.com",
			Status:      zone.ZoneStatusActive,
			Plan:        "Free",
			AccountID:   accountID,
			AccountName: "Test Account",
			NameServers: []string{"ns1.cloudflare.com"},
		},
	}, nil
}

func (s *StubZoneAdapter) Get(ctx context.Context, zoneID zone.ZoneID) (zone.Zone, error) {
	return zone.Zone{
		ID:          zoneID,
		Name:        "example.com",
		Status:      zone.ZoneStatusActive,
		Plan:        "Free",
		AccountID:   "acc-1",
		AccountName: "Test Account",
	}, nil
}

type StubDNSAdapter struct{}

func (s *StubDNSAdapter) List(ctx context.Context, zoneID zone.ZoneID, filter ports.DNSFilter) ([]dns.DNSRecord, error) {
	return []dns.DNSRecord{
		{
			ID:         "rec-1",
			ZoneID:     zoneID,
			Type:       dns.TypeA,
			Name:       "example.com",
			Content:    "1.2.3.4",
			Proxied:    true,
			TTL:        1,
			Proxiable:  true,
			ModifiedOn: time.Now(),
		},
	}, nil
}

func (s *StubDNSAdapter) Get(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error) {
	return dns.DNSRecord{
		ID:         id,
		ZoneID:     zoneID,
		Type:       dns.TypeA,
		Name:       "example.com",
		Content:    "1.2.3.4",
		Proxied:    true,
		TTL:        1,
		Proxiable:  true,
		ModifiedOn: time.Now(),
	}, nil
}

func (s *StubDNSAdapter) Create(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	return record, nil
}

func (s *StubDNSAdapter) Update(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	return record, nil
}

func (s *StubDNSAdapter) Delete(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) error {
	return nil
}

func (s *StubDNSAdapter) ToggleProxy(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error) {
	return dns.DNSRecord{
		ID:         id,
		ZoneID:     zoneID,
		Type:       dns.TypeA,
		Name:       "example.com",
		Content:    "1.2.3.4",
		Proxied:    false,
		TTL:        1,
		Proxiable:  true,
		ModifiedOn: time.Now(),
	}, nil
}

type StubTunnelAdapter struct{}

func (s *StubTunnelAdapter) List(ctx context.Context, accountID account.AccountID) ([]tunnel.Tunnel, error) {
	return []tunnel.Tunnel{
		{
			ID:        "tun-1",
			AccountID: accountID,
			Name:      "test-tunnel",
			Status:    tunnel.StatusHealthy,
			CreatedAt: time.Now(),
		},
	}, nil
}

func (s *StubTunnelAdapter) Get(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) (tunnel.Tunnel, error) {
	return tunnel.Tunnel{
		ID:        tunnelID,
		AccountID: accountID,
		Name:      "test-tunnel",
		Status:    tunnel.StatusHealthy,
		CreatedAt: time.Now(),
	}, nil
}

func (s *StubTunnelAdapter) GetConfiguration(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) ([]tunnel.IngressRule, error) {
	return []tunnel.IngressRule{
		{
			Hostname: "app.example.com",
			Service:  "http://localhost:8080",
		},
	}, nil
}

type StubWorkerAdapter struct{}

func (s *StubWorkerAdapter) List(ctx context.Context, accountID account.AccountID) ([]worker.WorkerScript, error) {
	return []worker.WorkerScript{
		{
			ID:         "worker-1",
			CreatedOn:  time.Now(),
			ModifiedOn: time.Now(),
			UsageModel: "bundled",
		},
	}, nil
}

type StubPagesAdapter struct{}

func (s *StubPagesAdapter) List(ctx context.Context, accountID account.AccountID) ([]pages.PagesProject, error) {
	return []pages.PagesProject{
		{
			ID:               "page-1",
			Name:             "test-pages",
			Subdomain:        "test-pages.pages.dev",
			ProductionBranch: "main",
			CreatedOn:        time.Now(),
		},
	}, nil
}

func (s *StubAdapter) ListZoneRulesets(ctx context.Context, zoneID zone.ZoneID) ([]ruleset.Ruleset, error) {
	return []ruleset.Ruleset{
		{
			ID:          "ruleset-1",
			Name:        "default ruleset",
			Phase:       "http_request_firewall_custom",
			Kind:        "zone",
			LastUpdated: time.Now(),
		},
	}, nil
}

func (s *StubAdapter) GetZoneTrafficSummary(ctx context.Context, zoneID zone.ZoneID, since time.Duration) (analytics.TrafficSummary, error) {
	return analytics.TrafficSummary{
		TotalRequests:  100,
		CachedRequests: 80,
		BandwidthBytes: 1024 * 1024,
		StatusCodes:    map[string]int64{"200": 100},
		Timestamp:      time.Now(),
	}, nil
}

func NewTestAppService() *application.AppService {
	stub := &StubAdapter{}
	repos := application.Repositories{
		Account:   stub,
		Zone:      &StubZoneAdapter{},
		DNS:       &StubDNSAdapter{},
		Tunnel:    &StubTunnelAdapter{},
		Worker:    &StubWorkerAdapter{},
		Pages:     &StubPagesAdapter{},
		Ruleset:   stub,
		Analytics: stub,
	}
	return application.NewAppService(repos, cache.NewMemoryCache())
}
