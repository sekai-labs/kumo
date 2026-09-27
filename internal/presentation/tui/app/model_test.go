package app_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/application"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/app"
	"github.com/sekai-labs/kumo/internal/presentation/tui/app/testhelpers"
	"github.com/sekai-labs/kumo/internal/presentation/tui/command"
	"github.com/sekai-labs/kumo/internal/presentation/tui/modals"
	"github.com/sekai-labs/kumo/internal/presentation/tui/views"
)

func createTestService(t *testing.T) *application.AppService {
	return testhelpers.NewTestAppService()
}

func TestAppModelLifecycle(t *testing.T) {
	svc := createTestService(t)
	model := app.New(app.Config{
		Service: svc,
		Profile: "test-profile",
	})

	initCmd := model.Init()
	if initCmd == nil {
		t.Fatal("expected Init command batch")
	}

	m, _ := model.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	model = m.(*app.AppModel)

	rendered := model.View()
	if rendered == "" {
		t.Fatal("expected non-empty rendered view")
	}

	m, _ = model.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	model = m.(*app.AppModel)
	if model.View() == "" {
		t.Fatal("medium layout render failed")
	}

	m, _ = model.Update(tea.WindowSizeMsg{Width: 70, Height: 24})
	model = m.(*app.AppModel)
	if model.View() == "" {
		t.Fatal("compact layout render failed")
	}

	m, _ = model.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	model = m.(*app.AppModel)
}

func TestAppModelViewSwitching(t *testing.T) {
	svc := createTestService(t)
	model := app.New(app.Config{
		Service: svc,
	})
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model = m.(*app.AppModel)

	m, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = m.(*app.AppModel)
	_ = cmd

	rendered := model.View()
	if rendered == "" {
		t.Fatal("expected non-empty render after switching to DNS view")
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = m.(*app.AppModel)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = m.(*app.AppModel)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	model = m.(*app.AppModel)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	model = m.(*app.AppModel)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model = m.(*app.AppModel)
}

func TestAppModelModalInteractions(t *testing.T) {
	svc := createTestService(t)
	model := app.New(app.Config{
		Service: svc,
	})
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model = m.(*app.AppModel)

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = m.(*app.AppModel)
	helpRender := model.View()
	if helpRender == "" {
		t.Fatal("expected help modal render")
	}

	m, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(*app.AppModel)
	if cmd != nil {
		msg := cmd()
		m, _ = model.Update(msg)
		model = m.(*app.AppModel)
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	model = m.(*app.AppModel)
	paletteRender := model.View()
	if paletteRender == "" {
		t.Fatal("expected palette modal render")
	}

	m, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(*app.AppModel)
	if cmd != nil {
		msg := cmd()
		m, _ = model.Update(msg)
		model = m.(*app.AppModel)
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})
	model = m.(*app.AppModel)
	zonePickerRender := model.View()
	if zonePickerRender == "" {
		t.Fatal("expected zone picker render")
	}

	m, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(*app.AppModel)
	if cmd != nil {
		msg := cmd()
		m, _ = model.Update(msg)
		model = m.(*app.AppModel)
	}

	m, _ = model.Update(views.OpenNewRecordModalMsg{ZoneID: "zone-123"})
	model = m.(*app.AppModel)
	newDnsRender := model.View()
	if newDnsRender == "" {
		t.Fatal("expected new dns modal render")
	}

	m, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(*app.AppModel)
	if cmd != nil {
		msg := cmd()
		m, _ = model.Update(msg)
		model = m.(*app.AppModel)
	}

	rec, err := dns.NewDNSRecord("rec-1", "zone-1", dns.TypeA, "test.example.com", "1.1.1.1", true, 300, "", nil, time.Now())
	if err != nil {
		t.Fatalf("failed to create DNS record: %v", err)
	}
	m, _ = model.Update(views.OpenDeleteRecordModalMsg{Record: rec})
	model = m.(*app.AppModel)
	delRender := model.View()
	if delRender == "" {
		t.Fatal("expected delete record modal render")
	}

	m, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = m.(*app.AppModel)
	if cmd != nil {
		msg := cmd()
		m, _ = model.Update(msg)
		model = m.(*app.AppModel)
	}
}

func TestAppModelAsyncDataFlow(t *testing.T) {
	svc := createTestService(t)
	model := app.New(app.Config{
		Service: svc,
	})
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model = m.(*app.AppModel)

	exp := time.Now().Add(24 * time.Hour)
	token := account.TokenInfo{
		ID:        "tok-1",
		Status:    "active",
		ExpiresOn: &exp,
	}
	m, _ = model.Update(views.TokenVerifiedMsg{TokenInfo: token})
	model = m.(*app.AppModel)

	accs := []account.Account{
		{ID: "acc-1", Name: "Engineering"},
		{ID: "acc-2", Name: "Personal"},
	}
	m, _ = model.Update(views.AccountsLoadedMsg{Accounts: accs})
	model = m.(*app.AppModel)

	zones := []zone.Zone{
		{ID: "zone-1", Name: "example.com", AccountID: "acc-1", Plan: "Enterprise"},
	}
	m, _ = model.Update(views.ZonesLoadedMsg{Zones: zones})
	model = m.(*app.AppModel)

	rec, _ := dns.NewDNSRecord("rec-1", "zone-1", dns.TypeA, "example.com", "93.184.216.34", true, 1, "", nil, time.Now())
	m, _ = model.Update(views.DNSLoadedMsg{Records: []dns.DNSRecord{rec}})
	model = m.(*app.AppModel)

	activeAt := time.Now().Add(-10 * time.Minute)
	tun, _ := tunnel.NewTunnel("tun-1", "acc-1", "prod-tunnel", tunnel.StatusHealthy, &activeAt, time.Now().Add(-24*time.Hour), true, nil, nil)
	m, _ = model.Update(views.TunnelsLoadedMsg{Tunnels: []tunnel.Tunnel{tun}})
	model = m.(*app.AppModel)

	wrk, _ := worker.NewWorkerScript("auth-worker", time.Now(), time.Now(), "bundled", false, false)
	m, _ = model.Update(views.WorkersLoadedMsg{Workers: []worker.WorkerScript{wrk}})
	model = m.(*app.AppModel)

	pgs, _ := pages.NewPagesProject("pg-1", "docs", "docs.sekai.dev", "main", []string{"docs.sekai.dev"}, time.Now())
	m, _ = model.Update(views.PagesLoadedMsg{Pages: []pages.PagesProject{pgs}})
	model = m.(*app.AppModel)

	rules := []ruleset.Ruleset{
		{ID: "rs-1", Name: "default", Phase: "http_request_firewall_custom", Kind: "zone", Description: "WAF rules"},
	}
	m, _ = model.Update(views.RulesetsLoadedMsg{Rulesets: rules})
	model = m.(*app.AppModel)

	summary := analytics.TrafficSummary{
		TotalRequests:  10000,
		CachedRequests: 7500,
		BandwidthBytes: 104857600,
		ThreatCount:    12,
		StatusCodes:    map[string]int64{"200": 8000, "304": 1500, "404": 400, "500": 100},
	}
	m, _ = model.Update(views.AnalyticsLoadedMsg{Summary: summary})
	model = m.(*app.AppModel)

	m, _ = model.Update(views.StatusMsg{Message: "Operation completed", IsError: false})
	model = m.(*app.AppModel)

	m, _ = model.Update(views.RecordCopiedMsg{Content: "93.184.216.34"})
	model = m.(*app.AppModel)

	rendered := model.View()
	if rendered == "" {
		t.Fatal("expected non-empty render after populating all domain data")
	}
}

func TestAppModelDirectCommands(t *testing.T) {
	svc := createTestService(t)
	model := app.New(app.Config{
		Service: svc,
	})
	m, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model = m.(*app.AppModel)

	m, refreshCmd := model.Update(command.RefreshMsg{})
	model = m.(*app.AppModel)
	if refreshCmd == nil {
		t.Fatal("expected cmd on RefreshMsg")
	}

	m, _ = model.Update(command.SwitchViewMsg{View: views.ViewDNS})
	model = m.(*app.AppModel)

	m, toggleCmd := model.Update(views.ToggleDNSProxyReqMsg{
		ZoneID:   "zone-1",
		RecordID: "rec-1",
	})
	model = m.(*app.AppModel)
	if toggleCmd == nil {
		t.Fatal("expected cmd on ToggleDNSProxyReqMsg")
	}

	newRec, _ := dns.NewDNSRecord("new-1", "zone-1", dns.TypeA, "test.example.com", "1.1.1.1", true, 300, "", nil, time.Now())
	m, submitCmd := model.Update(modals.DNSFormSubmitMsg{
		Mode:   modals.DNSFormCreate,
		Record: newRec,
	})
	model = m.(*app.AppModel)
	if submitCmd == nil {
		t.Fatal("expected cmd on DNSFormSubmitMsg")
	}

	m, quitCmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quitCmd == nil {
		t.Fatal("expected quit cmd on ctrl+c")
	}
}
