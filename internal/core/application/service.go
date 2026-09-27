package application

import (
	"context"
	"fmt"
	"time"

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

const (
	TTLAccount   = 5 * time.Minute
	TTLZone      = 5 * time.Minute
	TTLDNS       = 1 * time.Minute
	TTLTunnel    = 1 * time.Minute
	TTLWorker    = 2 * time.Minute
	TTLPages     = 2 * time.Minute
	TTLRuleset   = 2 * time.Minute
	TTLAnalytics = 1 * time.Minute
)

type Repositories struct {
	Account   ports.AccountRepository
	Zone      ports.ZoneRepository
	DNS       ports.DNSRepository
	Tunnel    ports.TunnelRepository
	Worker    ports.WorkerRepository
	Pages     ports.PagesRepository
	Ruleset   ports.RulesetRepository
	Analytics ports.AnalyticsRepository
}

type AppService struct {
	repos Repositories
	cache ports.Cache
}

func NewAppService(repos Repositories, cache ports.Cache) *AppService {
	return &AppService{
		repos: repos,
		cache: cache,
	}
}

func (s *AppService) VerifyToken(ctx context.Context, forceRefresh bool) (account.TokenInfo, error) {
	cacheKey := "token_info"
	if !forceRefresh && s.cache != nil {
		var cached account.TokenInfo
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	info, err := s.repos.Account.VerifyToken(ctx)
	if err != nil {
		return account.TokenInfo{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, info, TTLAccount)
	}
	return info, nil
}

func (s *AppService) ListAccounts(ctx context.Context, forceRefresh bool) ([]account.Account, error) {
	cacheKey := "accounts"
	if !forceRefresh && s.cache != nil {
		var cached []account.Account
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	accs, err := s.repos.Account.List(ctx)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, accs, TTLAccount)
	}
	return accs, nil
}

func (s *AppService) ListZones(ctx context.Context, accountID account.AccountID, forceRefresh bool) ([]zone.Zone, error) {
	cacheKey := fmt.Sprintf("zones:%s", accountID)
	if !forceRefresh && s.cache != nil {
		var cached []zone.Zone
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	zones, err := s.repos.Zone.List(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, zones, TTLZone)
	}
	return zones, nil
}

func (s *AppService) GetZone(ctx context.Context, zoneID zone.ZoneID, forceRefresh bool) (zone.Zone, error) {
	cacheKey := fmt.Sprintf("zone:%s", zoneID)
	if !forceRefresh && s.cache != nil {
		var cached zone.Zone
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	z, err := s.repos.Zone.Get(ctx, zoneID)
	if err != nil {
		return zone.Zone{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, z, TTLZone)
	}
	return z, nil
}

func (s *AppService) ListDNSRecords(ctx context.Context, zoneID zone.ZoneID, filter ports.DNSFilter, forceRefresh bool) ([]dns.DNSRecord, error) {
	cacheKey := fmt.Sprintf("dns:%s:%s:%s:%s", zoneID, filter.Type, filter.Name, filter.Content)
	if !forceRefresh && s.cache != nil {
		var cached []dns.DNSRecord
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	records, err := s.repos.DNS.List(ctx, zoneID, filter)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, records, TTLDNS)
	}
	return records, nil
}

func (s *AppService) CreateDNSRecord(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	rec, err := s.repos.DNS.Create(ctx, zoneID, record)
	if err != nil {
		return dns.DNSRecord{}, err
	}
	s.invalidateDNSCache(ctx, zoneID)
	return rec, nil
}

func (s *AppService) UpdateDNSRecord(ctx context.Context, zoneID zone.ZoneID, record dns.DNSRecord) (dns.DNSRecord, error) {
	rec, err := s.repos.DNS.Update(ctx, zoneID, record)
	if err != nil {
		return dns.DNSRecord{}, err
	}
	s.invalidateDNSCache(ctx, zoneID)
	return rec, nil
}

func (s *AppService) DeleteDNSRecord(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) error {
	if err := s.repos.DNS.Delete(ctx, zoneID, id); err != nil {
		return err
	}
	s.invalidateDNSCache(ctx, zoneID)
	return nil
}

func (s *AppService) ToggleDNSProxy(ctx context.Context, zoneID zone.ZoneID, id dns.RecordID) (dns.DNSRecord, error) {
	rec, err := s.repos.DNS.ToggleProxy(ctx, zoneID, id)
	if err != nil {
		return dns.DNSRecord{}, err
	}
	s.invalidateDNSCache(ctx, zoneID)
	return rec, nil
}

func (s *AppService) invalidateDNSCache(ctx context.Context, zoneID zone.ZoneID) {
	if s.cache == nil {
		return
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("dns:%s:::", zoneID))
}

func (s *AppService) ListTunnels(ctx context.Context, accountID account.AccountID, forceRefresh bool) ([]tunnel.Tunnel, error) {
	cacheKey := fmt.Sprintf("tunnels:%s", accountID)
	if !forceRefresh && s.cache != nil {
		var cached []tunnel.Tunnel
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	tunnels, err := s.repos.Tunnel.List(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, tunnels, TTLTunnel)
	}
	return tunnels, nil
}

func (s *AppService) GetTunnel(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID, forceRefresh bool) (tunnel.Tunnel, error) {
	cacheKey := fmt.Sprintf("tunnel:%s:%s", accountID, tunnelID)
	if !forceRefresh && s.cache != nil {
		var cached tunnel.Tunnel
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	tun, err := s.repos.Tunnel.Get(ctx, accountID, tunnelID)
	if err != nil {
		return tunnel.Tunnel{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, tun, TTLTunnel)
	}
	return tun, nil
}

func (s *AppService) GetTunnelConfiguration(ctx context.Context, accountID account.AccountID, tunnelID tunnel.TunnelID, forceRefresh bool) ([]tunnel.IngressRule, error) {
	cacheKey := fmt.Sprintf("tunnel_cfg:%s:%s", accountID, tunnelID)
	if !forceRefresh && s.cache != nil {
		var cached []tunnel.IngressRule
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	rules, err := s.repos.Tunnel.GetConfiguration(ctx, accountID, tunnelID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, rules, TTLTunnel)
	}
	return rules, nil
}

func (s *AppService) ListWorkers(ctx context.Context, accountID account.AccountID, forceRefresh bool) ([]worker.WorkerScript, error) {
	cacheKey := fmt.Sprintf("workers:%s", accountID)
	if !forceRefresh && s.cache != nil {
		var cached []worker.WorkerScript
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	workers, err := s.repos.Worker.List(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, workers, TTLWorker)
	}
	return workers, nil
}

func (s *AppService) ListPages(ctx context.Context, accountID account.AccountID, forceRefresh bool) ([]pages.PagesProject, error) {
	cacheKey := fmt.Sprintf("pages:%s", accountID)
	if !forceRefresh && s.cache != nil {
		var cached []pages.PagesProject
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	projects, err := s.repos.Pages.List(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, projects, TTLPages)
	}
	return projects, nil
}

func (s *AppService) ListZoneRulesets(ctx context.Context, zoneID zone.ZoneID, forceRefresh bool) ([]ruleset.Ruleset, error) {
	cacheKey := fmt.Sprintf("rulesets:%s", zoneID)
	if !forceRefresh && s.cache != nil {
		var cached []ruleset.Ruleset
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	rulesets, err := s.repos.Ruleset.ListZoneRulesets(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, rulesets, TTLRuleset)
	}
	return rulesets, nil
}

func (s *AppService) GetZoneTrafficSummary(ctx context.Context, zoneID zone.ZoneID, since time.Duration, forceRefresh bool) (analytics.TrafficSummary, error) {
	cacheKey := fmt.Sprintf("analytics:%s:%s", zoneID, since)
	if !forceRefresh && s.cache != nil {
		var cached analytics.TrafficSummary
		found, err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found {
			return cached, nil
		}
	}

	summary, err := s.repos.Analytics.GetZoneTrafficSummary(ctx, zoneID, since)
	if err != nil {
		return analytics.TrafficSummary{}, err
	}

	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, summary, TTLAnalytics)
	}
	return summary, nil
}

func (s *AppService) FlushCache(ctx context.Context) error {
	if s.cache != nil {
		return s.cache.Flush(ctx)
	}
	return nil
}
