package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/accounts"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/rulesets"
	"github.com/cloudflare/cloudflare-go/v4/shared"
	"github.com/cloudflare/cloudflare-go/v4/workers"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/cloudflare/cloudflare-go/v4/zones"

	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	domainDNS "github.com/sekai-labs/kumo/internal/core/domain/dns"
	domainPages "github.com/sekai-labs/kumo/internal/core/domain/pages"
	domainRuleset "github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	domainTunnel "github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	domainWorker "github.com/sekai-labs/kumo/internal/core/domain/worker"
	domainZone "github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/core/ports"
)

type ClientAdapter struct {
	client *cf.Client
}

var (
	_ ports.AccountRepository   = (*ClientAdapter)(nil)
	_ ports.ZoneRepository      = (*ZoneAdapter)(nil)
	_ ports.DNSRepository       = (*DNSAdapter)(nil)
	_ ports.TunnelRepository    = (*TunnelAdapter)(nil)
	_ ports.WorkerRepository    = (*WorkerAdapter)(nil)
	_ ports.PagesRepository     = (*PagesAdapter)(nil)
	_ ports.RulesetRepository   = (*ClientAdapter)(nil)
	_ ports.AnalyticsRepository = (*ClientAdapter)(nil)
)

type ZoneAdapter struct {
	adapter *ClientAdapter
}

func (z *ZoneAdapter) List(ctx context.Context, accountID account.AccountID) ([]domainZone.Zone, error) {
	return z.adapter.listZones(ctx, accountID)
}

func (z *ZoneAdapter) Get(ctx context.Context, zoneID domainZone.ZoneID) (domainZone.Zone, error) {
	return z.adapter.Get(ctx, zoneID)
}

type DNSAdapter struct {
	adapter *ClientAdapter
}

func (d *DNSAdapter) List(ctx context.Context, zoneID domainZone.ZoneID, filter ports.DNSFilter) ([]domainDNS.DNSRecord, error) {
	return d.adapter.ListRecords(ctx, zoneID, filter)
}

func (d *DNSAdapter) Get(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) (domainDNS.DNSRecord, error) {
	return d.adapter.GetRecord(ctx, zoneID, id)
}

func (d *DNSAdapter) Create(ctx context.Context, zoneID domainZone.ZoneID, record domainDNS.DNSRecord) (domainDNS.DNSRecord, error) {
	return d.adapter.Create(ctx, zoneID, record)
}

func (d *DNSAdapter) Update(ctx context.Context, zoneID domainZone.ZoneID, record domainDNS.DNSRecord) (domainDNS.DNSRecord, error) {
	return d.adapter.Update(ctx, zoneID, record)
}

func (d *DNSAdapter) Delete(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) error {
	return d.adapter.Delete(ctx, zoneID, id)
}

func (d *DNSAdapter) ToggleProxy(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) (domainDNS.DNSRecord, error) {
	return d.adapter.ToggleProxy(ctx, zoneID, id)
}

type TunnelAdapter struct {
	adapter *ClientAdapter
}

func (t *TunnelAdapter) List(ctx context.Context, accountID account.AccountID) ([]domainTunnel.Tunnel, error) {
	return t.adapter.listTunnels(ctx, accountID)
}

func (t *TunnelAdapter) Get(ctx context.Context, accountID account.AccountID, tunnelID domainTunnel.TunnelID) (domainTunnel.Tunnel, error) {
	return t.adapter.getTunnel(ctx, accountID, tunnelID)
}

func (t *TunnelAdapter) GetConfiguration(ctx context.Context, accountID account.AccountID, tunnelID domainTunnel.TunnelID) ([]domainTunnel.IngressRule, error) {
	return t.adapter.GetConfiguration(ctx, accountID, tunnelID)
}

type WorkerAdapter struct {
	adapter *ClientAdapter
}

func (w *WorkerAdapter) List(ctx context.Context, accountID account.AccountID) ([]domainWorker.WorkerScript, error) {
	return w.adapter.listWorkers(ctx, accountID)
}

type PagesAdapter struct {
	adapter *ClientAdapter
}

func (p *PagesAdapter) List(ctx context.Context, accountID account.AccountID) ([]domainPages.PagesProject, error) {
	return p.adapter.listPages(ctx, accountID)
}

func (a *ClientAdapter) Zones() *ZoneAdapter {
	return &ZoneAdapter{adapter: a}
}

func (a *ClientAdapter) DNS() *DNSAdapter {
	return &DNSAdapter{adapter: a}
}

func (a *ClientAdapter) Tunnels() *TunnelAdapter {
	return &TunnelAdapter{adapter: a}
}

func (a *ClientAdapter) Workers() *WorkerAdapter {
	return &WorkerAdapter{adapter: a}
}

func (a *ClientAdapter) Pages() *PagesAdapter {
	return &PagesAdapter{adapter: a}
}

func NewClient(token string, opts ...option.RequestOption) *ClientAdapter {
	defaultOpts := []option.RequestOption{
		option.WithAPIToken(token),
	}
	combinedOpts := append(defaultOpts, opts...)
	return &ClientAdapter{
		client: cf.NewClient(combinedOpts...),
	}
}

func (a *ClientAdapter) List(ctx context.Context) ([]account.Account, error) {
	pager := a.client.Accounts.ListAutoPaging(ctx, accounts.AccountListParams{})
	var result []account.Account
	for pager.Next() {
		acc := pager.Current()
		domainAcc, err := account.NewAccount(acc.ID, acc.Name, "standard", acc.CreatedOn)
		if err != nil {
			continue
		}
		result = append(result, domainAcc)
	}
	if err := pager.Err(); err != nil {
		return nil, MapError(err)
	}
	return result, nil
}

func (a *ClientAdapter) VerifyToken(ctx context.Context) (account.TokenInfo, error) {
	resp, err := a.client.User.Tokens.Verify(ctx)
	if err != nil {
		return account.TokenInfo{}, MapError(err)
	}

	var expiresOn *time.Time
	if !resp.ExpiresOn.IsZero() {
		exp := resp.ExpiresOn
		expiresOn = &exp
	}

	status := account.TokenStatusActive
	switch resp.Status {
	case "disabled":
		status = account.TokenStatusDisabled
	case "expired":
		status = account.TokenStatusExpired
	}

	return account.TokenInfo{
		ID:        resp.ID,
		Status:    status,
		ExpiresOn: expiresOn,
	}, nil
}

func (a *ClientAdapter) ListZones(ctx context.Context, accountID account.AccountID) ([]domainZone.Zone, error) {
	return a.listZones(ctx, accountID)
}

func (a *ClientAdapter) listZones(ctx context.Context, accountID account.AccountID) ([]domainZone.Zone, error) {
	params := zones.ZoneListParams{}
	if accountID != "" {
		params.Account = cf.F(zones.ZoneListParamsAccount{
			ID: cf.F(string(accountID)),
		})
	}

	pager := a.client.Zones.ListAutoPaging(ctx, params)
	var result []domainZone.Zone
	for pager.Next() {
		z := pager.Current()
		planName := z.Plan.Name
		if planName == "" {
			planName = "Free"
		}

		domZone, err := domainZone.NewZone(
			z.ID,
			z.Name,
			string(z.Status),
			planName,
			account.AccountID(z.Account.ID),
			z.Account.Name,
			z.NameServers,
			z.Paused,
		)
		if err != nil {
			continue
		}
		result = append(result, domZone)
	}
	if err := pager.Err(); err != nil {
		return nil, MapError(err)
	}
	return result, nil
}

func (a *ClientAdapter) Get(ctx context.Context, zoneID domainZone.ZoneID) (domainZone.Zone, error) {
	z, err := a.client.Zones.Get(ctx, zones.ZoneGetParams{
		ZoneID: cf.F(string(zoneID)),
	})
	if err != nil {
		return domainZone.Zone{}, MapError(err)
	}

	planName := z.Plan.Name
	if planName == "" {
		planName = "Free"
	}

	return domainZone.NewZone(
		z.ID,
		z.Name,
		string(z.Status),
		planName,
		account.AccountID(z.Account.ID),
		z.Account.Name,
		z.NameServers,
		z.Paused,
	)
}

func (a *ClientAdapter) ListRecords(ctx context.Context, zoneID domainZone.ZoneID, filter ports.DNSFilter) ([]domainDNS.DNSRecord, error) {
	params := dns.RecordListParams{
		ZoneID: cf.F(string(zoneID)),
	}
	if filter.Type != "" {
		params.Type = cf.F(dns.RecordListParamsType(filter.Type))
	}
	if filter.Name != "" {
		params.Name = cf.F(dns.RecordListParamsName{
			Contains: cf.F(filter.Name),
		})
	}
	if filter.Content != "" {
		params.Content = cf.F(dns.RecordListParamsContent{
			Contains: cf.F(filter.Content),
		})
	}

	pager := a.client.DNS.Records.ListAutoPaging(ctx, params)
	var records []domainDNS.DNSRecord
	for pager.Next() {
		rec := pager.Current()
		domRec := mapDNSRecordResponse(&rec, zoneID)
		records = append(records, domRec)
	}
	if err := pager.Err(); err != nil {
		return nil, MapError(err)
	}
	return records, nil
}

func (a *ClientAdapter) GetRecord(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) (domainDNS.DNSRecord, error) {
	rec, err := a.client.DNS.Records.Get(ctx, string(id), dns.RecordGetParams{
		ZoneID: cf.F(string(zoneID)),
	})
	if err != nil {
		return domainDNS.DNSRecord{}, MapError(err)
	}
	return mapDNSRecordResponse(rec, zoneID), nil
}

func (a *ClientAdapter) Create(ctx context.Context, zoneID domainZone.ZoneID, record domainDNS.DNSRecord) (domainDNS.DNSRecord, error) {
	body := dns.RecordNewParamsBody{
		Name:    cf.F(record.Name),
		Type:    cf.F(dns.RecordNewParamsBodyType(record.Type)),
		Content: cf.F(record.Content),
		TTL:     cf.F(dns.TTL(record.TTL)),
		Proxied: cf.F(record.Proxied),
		Comment: cf.F(record.Comment),
	}
	if record.Priority != nil {
		body.Priority = cf.F(float64(*record.Priority))
	}

	params := dns.RecordNewParams{
		ZoneID: cf.F(string(zoneID)),
		Body:   body,
	}

	res, err := a.client.DNS.Records.New(ctx, params)
	if err != nil {
		return domainDNS.DNSRecord{}, MapError(err)
	}
	return mapDNSRecordResponse(res, zoneID), nil
}

func (a *ClientAdapter) Update(ctx context.Context, zoneID domainZone.ZoneID, record domainDNS.DNSRecord) (domainDNS.DNSRecord, error) {
	body := dns.RecordUpdateParamsBody{
		Name:    cf.F(record.Name),
		Type:    cf.F(dns.RecordUpdateParamsBodyType(record.Type)),
		Content: cf.F(record.Content),
		TTL:     cf.F(dns.TTL(record.TTL)),
		Proxied: cf.F(record.Proxied),
		Comment: cf.F(record.Comment),
	}
	if record.Priority != nil {
		body.Priority = cf.F(float64(*record.Priority))
	}

	params := dns.RecordUpdateParams{
		ZoneID: cf.F(string(zoneID)),
		Body:   body,
	}

	res, err := a.client.DNS.Records.Update(ctx, string(record.ID), params)
	if err != nil {
		return domainDNS.DNSRecord{}, MapError(err)
	}
	return mapDNSRecordResponse(res, zoneID), nil
}

func (a *ClientAdapter) Delete(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) error {
	_, err := a.client.DNS.Records.Delete(ctx, string(id), dns.RecordDeleteParams{
		ZoneID: cf.F(string(zoneID)),
	})
	if err != nil {
		return MapError(err)
	}
	return nil
}

func (a *ClientAdapter) ToggleProxy(ctx context.Context, zoneID domainZone.ZoneID, id domainDNS.RecordID) (domainDNS.DNSRecord, error) {
	rec, err := a.GetRecord(ctx, zoneID, id)
	if err != nil {
		return domainDNS.DNSRecord{}, err
	}

	if !rec.Proxiable {
		return domainDNS.DNSRecord{}, domainDNS.ErrCannotProxy
	}

	newProxied := !rec.Proxied
	body := dns.RecordEditParamsBody{
		Proxied: cf.F(newProxied),
	}
	params := dns.RecordEditParams{
		ZoneID: cf.F(string(zoneID)),
		Body:   body,
	}

	res, err := a.client.DNS.Records.Edit(ctx, string(id), params)
	if err != nil {
		return domainDNS.DNSRecord{}, MapError(err)
	}
	return mapDNSRecordResponse(res, zoneID), nil
}

func mapDNSRecordResponse(rec *dns.RecordResponse, zoneID domainZone.ZoneID) domainDNS.DNSRecord {
	recType := domainDNS.RecordType(strings.ToUpper(string(rec.Type)))
	proxiable := domainDNS.ProxiableTypes[recType]

	var prio *uint16
	if rec.Priority > 0 {
		p := uint16(rec.Priority)
		prio = &p
	}

	return domainDNS.DNSRecord{
		ID:         domainDNS.RecordID(rec.ID),
		ZoneID:     zoneID,
		Type:       recType,
		Name:       rec.Name,
		Content:    rec.Content,
		Proxied:    rec.Proxied,
		TTL:        domainDNS.TTL(rec.TTL),
		Proxiable:  proxiable,
		Comment:    rec.Comment,
		Priority:   prio,
		ModifiedOn: rec.ModifiedOn,
	}
}

func (a *ClientAdapter) listTunnels(ctx context.Context, accountID account.AccountID) ([]domainTunnel.Tunnel, error) {
	pager := a.client.ZeroTrust.Tunnels.Cloudflared.ListAutoPaging(ctx, zero_trust.TunnelCloudflaredListParams{
		AccountID: cf.F(string(accountID)),
	})

	var result []domainTunnel.Tunnel
	for pager.Next() {
		t := pager.Current()
		var connsActiveAt *time.Time
		if !t.ConnsActiveAt.IsZero() {
			at := t.ConnsActiveAt
			connsActiveAt = &at
		}

		var connectors []domainTunnel.Connector
		if t.Connections != nil {
			if rawJSON, err := json.Marshal(t.Connections); err == nil {
				var cfConns []shared.CloudflareTunnelConnection
				if err := json.Unmarshal(rawJSON, &cfConns); err == nil {
					for _, c := range cfConns {
						var openedAt *time.Time
						if !c.OpenedAt.IsZero() {
							o := c.OpenedAt
							openedAt = &o
						}
						statusStr := "active"
						if c.IsPendingReconnect {
							statusStr = "pending_reconnect"
						}
						connectors = append(connectors, domainTunnel.Connector{
							ID:            c.ID,
							ClientVersion: c.ClientVersion,
							ColoName:      c.ColoName,
							OpenedAt:      openedAt,
							OriginIP:      c.OriginIP,
							Status:        statusStr,
						})
					}
				}
			}
		}

		status := domainTunnel.DeriveTunnelStatus(connectors)
		if len(connectors) == 0 && t.Status != "" {
			switch t.Status {
			case "healthy":
				status = domainTunnel.StatusHealthy
			case "degraded":
				status = domainTunnel.StatusDegraded
			case "down":
				status = domainTunnel.StatusDown
			case "inactive":
				status = domainTunnel.StatusInactive
			}
		}

		domTunnel, err := domainTunnel.NewTunnel(
			t.ID,
			accountID,
			t.Name,
			status,
			connsActiveAt,
			t.CreatedAt,
			t.RemoteConfig,
			connectors,
			nil,
		)
		if err != nil {
			continue
		}
		result = append(result, domTunnel)
	}

	if err := pager.Err(); err != nil {
		return nil, MapError(err)
	}
	return result, nil
}

func (a *ClientAdapter) getTunnel(ctx context.Context, accountID account.AccountID, tunnelID domainTunnel.TunnelID) (domainTunnel.Tunnel, error) {
	t, err := a.client.ZeroTrust.Tunnels.Cloudflared.Get(ctx, string(tunnelID), zero_trust.TunnelCloudflaredGetParams{
		AccountID: cf.F(string(accountID)),
	})
	if err != nil {
		return domainTunnel.Tunnel{}, MapError(err)
	}

	var connsActiveAt *time.Time
	if !t.ConnsActiveAt.IsZero() {
		at := t.ConnsActiveAt
		connsActiveAt = &at
	}

	var connectors []domainTunnel.Connector
	if t.Connections != nil {
		if rawJSON, err := json.Marshal(t.Connections); err == nil {
			var cfConns []shared.CloudflareTunnelConnection
			if err := json.Unmarshal(rawJSON, &cfConns); err == nil {
				for _, c := range cfConns {
					var openedAt *time.Time
					if !c.OpenedAt.IsZero() {
						o := c.OpenedAt
						openedAt = &o
					}
					statusStr := "active"
					if c.IsPendingReconnect {
						statusStr = "pending_reconnect"
					}
					connectors = append(connectors, domainTunnel.Connector{
						ID:            c.ID,
						ClientVersion: c.ClientVersion,
						ColoName:      c.ColoName,
						OpenedAt:      openedAt,
						OriginIP:      c.OriginIP,
						Status:        statusStr,
					})
				}
			}
		}
	}

	ingressRules, _ := a.GetConfiguration(ctx, accountID, tunnelID)

	status := domainTunnel.DeriveTunnelStatus(connectors)
	if len(connectors) == 0 && t.Status != "" {
		switch t.Status {
		case "healthy":
			status = domainTunnel.StatusHealthy
		case "degraded":
			status = domainTunnel.StatusDegraded
		case "down":
			status = domainTunnel.StatusDown
		case "inactive":
			status = domainTunnel.StatusInactive
		}
	}

	return domainTunnel.NewTunnel(
		t.ID,
		accountID,
		t.Name,
		status,
		connsActiveAt,
		t.CreatedAt,
		t.RemoteConfig,
		connectors,
		ingressRules,
	)
}

func (a *ClientAdapter) GetConfiguration(ctx context.Context, accountID account.AccountID, tunnelID domainTunnel.TunnelID) ([]domainTunnel.IngressRule, error) {
	cfg, err := a.client.ZeroTrust.Tunnels.Cloudflared.Configurations.Get(ctx, string(tunnelID), zero_trust.TunnelCloudflaredConfigurationGetParams{
		AccountID: cf.F(string(accountID)),
	})
	if err != nil {
		return nil, MapError(err)
	}

	var rules []domainTunnel.IngressRule
	for _, ing := range cfg.Config.Ingress {
		rules = append(rules, domainTunnel.IngressRule{
			Hostname: ing.Hostname,
			Path:     ing.Path,
			Service:  ing.Service,
		})
	}
	return rules, nil
}

func (a *ClientAdapter) listWorkers(ctx context.Context, accountID account.AccountID) ([]domainWorker.WorkerScript, error) {
	res, err := a.client.Workers.Scripts.List(ctx, workers.ScriptListParams{
		AccountID: cf.F(string(accountID)),
	})
	if err != nil {
		return nil, MapError(err)
	}

	var scripts []domainWorker.WorkerScript
	for _, s := range res.Result {
		domScript, err := domainWorker.NewWorkerScript(
			s.ID,
			s.CreatedOn,
			s.ModifiedOn,
			string(s.UsageModel),
			s.Logpush,
			s.HasAssets,
		)
		if err != nil {
			continue
		}
		scripts = append(scripts, domScript)
	}
	return scripts, nil
}

func (a *ClientAdapter) listPages(ctx context.Context, accountID account.AccountID) ([]domainPages.PagesProject, error) {
	var resp struct {
		Result []struct {
			ID               string    `json:"id"`
			Name             string    `json:"name"`
			Subdomain        string    `json:"subdomain"`
			ProductionBranch string    `json:"production_branch"`
			Domains          []string  `json:"domains"`
			CreatedOn        time.Time `json:"created_on"`
		} `json:"result"`
	}

	path := fmt.Sprintf("accounts/%s/pages/projects", accountID)
	err := a.client.Get(ctx, path, nil, &resp)
	if err != nil {
		return nil, MapError(err)
	}

	var projects []domainPages.PagesProject
	for _, p := range resp.Result {
		domProj, err := domainPages.NewPagesProject(
			p.ID,
			p.Name,
			p.Subdomain,
			p.ProductionBranch,
			p.Domains,
			p.CreatedOn,
		)
		if err != nil {
			continue
		}
		projects = append(projects, domProj)
	}
	return projects, nil
}

func (a *ClientAdapter) ListZoneRulesets(ctx context.Context, zoneID domainZone.ZoneID) ([]domainRuleset.Ruleset, error) {
	pager := a.client.Rulesets.ListAutoPaging(ctx, rulesets.RulesetListParams{
		ZoneID: cf.F(string(zoneID)),
	})

	var result []domainRuleset.Ruleset
	for pager.Next() {
		rs := pager.Current()
		detail, err := a.client.Rulesets.Get(ctx, rs.ID, rulesets.RulesetGetParams{
			ZoneID: cf.F(string(zoneID)),
		})
		rulesCount := 0
		if err == nil && detail != nil {
			rulesCount = len(detail.Rules)
		}

		domRs, err := domainRuleset.NewRuleset(
			rs.ID,
			rs.Name,
			string(rs.Phase),
			string(rs.Kind),
			rs.LastUpdated,
			rs.Description,
			rulesCount,
		)
		if err != nil {
			continue
		}
		result = append(result, domRs)
	}

	if err := pager.Err(); err != nil {
		return nil, MapError(err)
	}
	return result, nil
}

func (a *ClientAdapter) GetZoneTrafficSummary(ctx context.Context, zoneID domainZone.ZoneID, since time.Duration) (analytics.TrafficSummary, error) {
	sinceMinutes := -int(since.Minutes())
	if sinceMinutes == 0 {
		sinceMinutes = -1440
	}

	var resp struct {
		Result struct {
			Totals struct {
				Requests struct {
					All        int64            `json:"all"`
					Cached     int64            `json:"cached"`
					HTTPStatus map[string]int64 `json:"http_status"`
				} `json:"requests"`
				Bandwidth struct {
					All int64 `json:"all"`
				} `json:"bandwidth"`
				Threats struct {
					All int64 `json:"all"`
				} `json:"threats"`
			} `json:"totals"`
		} `json:"result"`
	}

	path := fmt.Sprintf("zones/%s/analytics/dashboard?since=%d", zoneID, sinceMinutes)
	err := a.client.Get(ctx, path, nil, &resp)
	if err != nil {
		return analytics.TrafficSummary{}, MapError(err)
	}

	statusCodes := resp.Result.Totals.Requests.HTTPStatus
	if statusCodes == nil {
		statusCodes = make(map[string]int64)
	}

	return analytics.TrafficSummary{
		TotalRequests:  resp.Result.Totals.Requests.All,
		CachedRequests: resp.Result.Totals.Requests.Cached,
		BandwidthBytes: resp.Result.Totals.Bandwidth.All,
		StatusCodes:    statusCodes,
		ThreatCount:    resp.Result.Totals.Threats.All,
		Timestamp:      time.Now(),
	}, nil
}
