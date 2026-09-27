package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
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
)

type mockCache struct {
	store map[string]any
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]any)}
}

func (m *mockCache) Get(ctx context.Context, key string, target any) (bool, error) {
	val, ok := m.store[key]
	if !ok {
		return false, nil
	}
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() == reflect.Pointer && !targetVal.IsNil() {
		targetVal.Elem().Set(reflect.ValueOf(val))
		return true, nil
	}
	return false, errors.New("target must be non-nil pointer")
}

func (m *mockCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	m.store[key] = val
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.store, key)
	return nil
}

func (m *mockCache) Flush(ctx context.Context) error {
	m.store = make(map[string]any)
	return nil
}

type mockAccountRepo struct {
	accounts []account.Account
	token    account.TokenInfo
	err      error
	callList int
}

func (m *mockAccountRepo) List(ctx context.Context) ([]account.Account, error) {
	m.callList++
	return m.accounts, m.err
}

func (m *mockAccountRepo) VerifyToken(ctx context.Context) (account.TokenInfo, error) {
	return m.token, m.err
}

type mockZoneRepo struct {
	zones    []zone.Zone
	err      error
	callList int
}

func (m *mockZoneRepo) List(ctx context.Context, accountID account.AccountID) ([]zone.Zone, error) {
	m.callList++
	return m.zones, m.err
}

func (m *mockZoneRepo) Get(ctx context.Context, zoneID zone.ZoneID) (zone.Zone, error) {
	for _, z := range m.zones {
		if z.ID == zoneID {
			return z, nil
		}
	}
	return zone.Zone{}, ports.ErrNotFound
}

type mockDNSRepo struct {
	records    []dns.DNSRecord
	callList   int
	callCreate int
	callUpdate int
	callDelete int
	callToggle int
}

func (m *mockDNSRepo) List(ctx context.Context, zoneID zone.ZoneID, filter ports.DNSFilter) ([]dns.DNSRecord, error) {
	m.callList++
	return m.records, nil
}

func (m *mockDNSRepo) Get(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error) {
	for _, r := range m.records {
		if r.ID == id {
			return r, nil
		}
	}
	return dns.DNSRecord{}, ports.ErrNotFound
}

func (m *mockDNSRepo) Create(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	m.callCreate++
	m.records = append(m.records, record)
	return record, nil
}

func (m *mockDNSRepo) Update(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	m.callUpdate++
	for i, r := range m.records {
		if r.ID == record.ID {
			m.records[i] = record
			return record, nil
		}
	}
	return dns.DNSRecord{}, ports.ErrNotFound
}

func (m *mockDNSRepo) Delete(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) error {
	m.callDelete++
	for i, r := range m.records {
		if r.ID == id {
			m.records = append(m.records[:i], m.records[i+1:]...)
			return nil
		}
	}
	return ports.ErrNotFound
}

func (m *mockDNSRepo) ToggleProxy(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error) {
	m.callToggle++
	for i, r := range m.records {
		if r.ID == id {
			m.records[i].Proxied = !m.records[i].Proxied
			return m.records[i], nil
		}
	}
	return dns.DNSRecord{}, ports.ErrNotFound
}

type mockTunnelRepo struct {
	tunnels  []tunnel.Tunnel
	callList int
}

func (m *mockTunnelRepo) List(ctx context.Context, accountID account.AccountID) ([]tunnel.Tunnel, error) {
	m.callList++
	return m.tunnels, nil
}

func (m *mockTunnelRepo) Get(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) (tunnel.Tunnel, error) {
	for _, t := range m.tunnels {
		if t.ID == tunnelID {
			return t, nil
		}
	}
	return tunnel.Tunnel{}, ports.ErrNotFound
}

func (m *mockTunnelRepo) GetConfiguration(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID) ([]tunnel.IngressRule, error) {
	for _, t := range m.tunnels {
		if t.ID == tunnelID {
			return t.IngressRules, nil
		}
	}
	return nil, ports.ErrNotFound
}

type mockWorkerRepo struct {
	scripts  []worker.WorkerScript
	callList int
}

func (m *mockWorkerRepo) List(ctx context.Context, accountID account.AccountID) ([]worker.WorkerScript, error) {
	m.callList++
	return m.scripts, nil
}

type mockPagesRepo struct {
	projects []pages.PagesProject
	callList int
}

func (m *mockPagesRepo) List(ctx context.Context, accountID account.AccountID) ([]pages.PagesProject, error) {
	m.callList++
	return m.projects, nil
}

type mockRulesetRepo struct {
	rulesets []ruleset.Ruleset
	callList int
}

func (m *mockRulesetRepo) ListZoneRulesets(ctx context.Context, zoneID zone.ZoneID) ([]ruleset.Ruleset, error) {
	m.callList++
	return m.rulesets, nil
}

type mockAnalyticsRepo struct {
	summary analytics.TrafficSummary
}

func (m *mockAnalyticsRepo) GetZoneTrafficSummary(ctx context.Context, zoneID zone.ZoneID, since time.Duration) (analytics.TrafficSummary, error) {
	return m.summary, nil
}

func TestAppService_Accounts_CacheAndForceRefresh(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	accRepo := &mockAccountRepo{
		accounts: []account.Account{
			{ID: "acc-1", Name: "Primary", Type: "standard"},
		},
	}
	svc := application.NewAppService(application.Repositories{Account: accRepo}, cache)

	accs, err := svc.ListAccounts(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accs) != 1 || accRepo.callList != 1 {
		t.Fatalf("expected 1 account and 1 repo call, got %d calls", accRepo.callList)
	}

	accs2, err := svc.ListAccounts(ctx, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accs2) != 1 || accRepo.callList != 1 {
		t.Fatalf("expected cached result without calling repo, call count: %d", accRepo.callList)
	}

	_, err = svc.ListAccounts(ctx, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accRepo.callList != 2 {
		t.Fatalf("expected 2 repo calls after forceRefresh, got %d", accRepo.callList)
	}
}

func TestAppService_DNS_OperationsAndCacheInvalidation(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	zoneID := zone.ZoneID("zone-123")
	recID := dns.RecordID("rec-1")

	dnsRepo := &mockDNSRepo{
		records: []dns.DNSRecord{
			{
				ID:        recID,
				ZoneID:    zoneID,
				Type:      dns.TypeA,
				Name:      "api.example.com",
				Content:   "1.1.1.1",
				Proxied:   true,
				TTL:       1,
				Proxiable: true,
			},
		},
	}

	svc := application.NewAppService(application.Repositories{DNS: dnsRepo}, cache)

	filter := ports.DNSFilter{}
	recs, err := svc.ListDNSRecords(ctx, zoneID, filter, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 || dnsRepo.callList != 1 {
		t.Fatalf("expected 1 record and 1 list call")
	}

	toggled, err := svc.ToggleDNSProxy(ctx, zoneID, recID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if toggled.Proxied {
		t.Fatalf("expected proxied to be false after toggle")
	}

	recsAfter, err := svc.ListDNSRecords(ctx, zoneID, filter, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dnsRepo.callList != 2 {
		t.Fatalf("expected 2 list calls after invalidation, got %d", dnsRepo.callList)
	}
	if recsAfter[0].Proxied != false {
		t.Fatalf("expected updated record in list")
	}

	err = svc.DeleteDNSRecord(ctx, zoneID, recID)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if dnsRepo.callDelete != 1 {
		t.Fatalf("expected 1 delete call")
	}
}

func TestAppService_Tunnels_ListAndConfig(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	accID := account.AccountID("acc-1")
	tunID := tunnel.TunnelID("tun-1")

	tunRepo := &mockTunnelRepo{
		tunnels: []tunnel.Tunnel{
			{
				ID:        tunID,
				AccountID: accID,
				Name:      "k8s-ingress",
				Status:    tunnel.StatusHealthy,
				IngressRules: []tunnel.IngressRule{
					{Hostname: "api.domain.com", Service: "http://localhost:3000"},
				},
			},
		},
	}

	svc := application.NewAppService(application.Repositories{Tunnel: tunRepo}, cache)

	tuns, err := svc.ListTunnels(ctx, accID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tuns) != 1 || tunRepo.callList != 1 {
		t.Fatalf("expected 1 tunnel and 1 repo call")
	}

	cfg, err := svc.GetTunnelConfiguration(ctx, accID, tunID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg) != 1 || cfg[0].Hostname != "api.domain.com" {
		t.Fatalf("unexpected ingress rules: %v", cfg)
	}
}

func TestAppService_WorkersAndPages(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	accID := account.AccountID("acc-1")

	workerRepo := &mockWorkerRepo{
		scripts: []worker.WorkerScript{
			{ID: "worker-1", UsageModel: "bundled"},
		},
	}
	pagesRepo := &mockPagesRepo{
		projects: []pages.PagesProject{
			{ID: "page-1", Name: "landing"},
		},
	}

	svc := application.NewAppService(application.Repositories{
		Worker: workerRepo,
		Pages:  pagesRepo,
	}, cache)

	workers, err := svc.ListWorkers(ctx, accID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workers) != 1 || workerRepo.callList != 1 {
		t.Fatalf("expected 1 worker")
	}

	projs, err := svc.ListPages(ctx, accID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projs) != 1 || pagesRepo.callList != 1 {
		t.Fatalf("expected 1 page project")
	}
}

func TestAppService_RulesetsAndAnalytics(t *testing.T) {
	ctx := context.Background()
	cache := newMockCache()
	zoneID := zone.ZoneID("zone-1")

	rulesetRepo := &mockRulesetRepo{
		rulesets: []ruleset.Ruleset{
			{ID: "rs-1", Name: "Managed Rules"},
		},
	}
	analyticsRepo := &mockAnalyticsRepo{
		summary: analytics.TrafficSummary{
			TotalRequests:  1000,
			CachedRequests: 400,
		},
	}

	svc := application.NewAppService(application.Repositories{
		Ruleset:   rulesetRepo,
		Analytics: analyticsRepo,
	}, cache)

	rs, err := svc.ListZoneRulesets(ctx, zoneID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs) != 1 || rulesetRepo.callList != 1 {
		t.Fatalf("expected 1 ruleset")
	}

	analyticsSummary, err := svc.GetZoneTrafficSummary(ctx, zoneID, 24*time.Hour, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analyticsSummary.TotalRequests != 1000 || analyticsSummary.CacheHitRatio() != 40.0 {
		t.Fatalf("unexpected analytics summary: %+v", analyticsSummary)
	}
}
