package views_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/views"
)

func sampleDimensions() layout.Dimensions {
	return layout.Calculate(120, 30, layout.PaneList)
}

func TestNav(t *testing.T) {
	nav := views.NewNav()
	nav.SetDimensions(24, 20)
	nav.SetFocused(true)
	nav.SetBadge(views.ViewDNS, "12")

	if nav.ActiveItem().ID != views.ViewOverview {
		t.Fatalf("expected active item to be %s, got %s", views.ViewOverview, nav.ActiveItem().ID)
	}

	nav.Next()
	if nav.ActiveItem().ID != views.ViewDNS {
		t.Fatalf("expected active item to be %s, got %s", views.ViewDNS, nav.ActiveItem().ID)
	}

	nav.Prev()
	if nav.ActiveItem().ID != views.ViewOverview {
		t.Fatalf("expected active item to be %s, got %s", views.ViewOverview, nav.ActiveItem().ID)
	}

	nav.SelectByID(views.ViewTunnels)
	if nav.ActiveItem().ID != views.ViewTunnels {
		t.Fatalf("expected active item to be %s, got %s", views.ViewTunnels, nav.ActiveItem().ID)
	}

	_, cmd := nav.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if nav.ActiveItem().ID != views.ViewDNS {
		t.Fatalf("expected active item to be %s after pressing '2', got %s", views.ViewDNS, nav.ActiveItem().ID)
	}
	if cmd != nil {
		msg := cmd()
		if selMsg, ok := msg.(views.SelectNavMsg); !ok || selMsg.ID != views.ViewDNS {
			t.Fatalf("unexpected select nav msg: %v", msg)
		}
	}

	output := nav.View()
	if len(output) == 0 {
		t.Fatal("expected non-empty nav view")
	}
}

func TestHeader(t *testing.T) {
	hdr := views.NewHeader()
	hdr.SetDimensions(120, 2)
	hdr.SetBreadcrumb("DNS Records")
	hdr.SetProfile("prod")
	hdr.SetAccount("Acme Corp", "acc-123456789")
	hdr.SetZone("example.com")
	hdr.SetTokenStatus(true, "Active")
	hdr.SetLoading(true)

	if err := hdr.Init(); err == nil {

	}

	hdr.Update(nil)
	output := hdr.View()
	if len(output) == 0 {
		t.Fatal("expected non-empty header view")
	}
}

func TestFooter(t *testing.T) {
	ftr := views.NewFooter()
	ftr.SetDimensions(120, 1)
	ftr.SetShortcuts([]views.Shortcut{
		{Key: "p", Desc: "toggle proxy"},
		{Key: "n", Desc: "new record"},
	})

	ftr.SetStatus("Record updated successfully", false, time.Second)
	output := ftr.View()
	if len(output) == 0 {
		t.Fatal("expected non-empty footer view")
	}

	ftr.ClearStatus()
	output2 := ftr.View()
	if len(output2) == 0 {
		t.Fatal("expected non-empty footer view after clear")
	}
}

func TestOverviewView(t *testing.T) {
	v := views.NewOverviewView()
	v.SetDimensions(sampleDimensions())

	z, _ := zone.NewZone("zone-1", "example.com", "active", "Enterprise", "acc-1", "Acme", []string{"ns1.cloudflare.com"}, false)
	tun, _ := tunnel.NewTunnel("tun-1", "acc-1", "k8s-ingress", tunnel.StatusHealthy, nil, time.Now(), true, nil, nil)
	rec, _ := dns.NewDNSRecord("rec-1", "zone-1", dns.TypeA, "api.example.com", "192.0.2.1", true, 300, "API endpoint", nil, time.Now())
	w, _ := worker.NewWorkerScript("auth-worker", time.Now(), time.Now(), "bundled", true, false)
	p, _ := pages.NewPagesProject("proj-1", "frontend", "frontend.pages.dev", "main", []string{"app.example.com"}, time.Now())

	v.Update(views.ZonesLoadedMsg{Zones: []zone.Zone{z}})
	v.Update(views.TunnelsLoadedMsg{Tunnels: []tunnel.Tunnel{tun}})
	v.Update(views.DNSLoadedMsg{Records: []dns.DNSRecord{rec}})
	v.Update(views.WorkersLoadedMsg{Workers: []worker.WorkerScript{w}})
	v.Update(views.PagesLoadedMsg{Pages: []pages.PagesProject{p}})
	v.Update(views.ActiveZoneChangedMsg{Zone: z})
	v.Update(views.ActiveAccountChangedMsg{Account: account.Account{ID: "acc-1", Name: "Acme"}})

	if len(v.Shortcuts()) == 0 {
		t.Fatal("expected non-empty shortcuts")
	}

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty overview render")
	}
}

func TestDNSView(t *testing.T) {
	v := views.NewDNSView()
	v.SetDimensions(sampleDimensions())

	z, _ := zone.NewZone("zone-1", "example.com", "active", "Pro", "acc-1", "Acme", nil, false)
	rec1, _ := dns.NewDNSRecord("rec-1", "zone-1", dns.TypeA, "api.example.com", "1.2.3.4", true, 300, "API", nil, time.Now())
	rec2, _ := dns.NewDNSRecord("rec-2", "zone-1", dns.TypeCNAME, "mail.example.com", "mail.google.com", false, 1, "", nil, time.Now())

	v.Update(views.ActiveZoneChangedMsg{Zone: z})
	v.Update(views.DNSLoadedMsg{Records: []dns.DNSRecord{rec1, rec2}})

	if v.SelectedRecord() == nil {
		t.Fatal("expected selected record not to be nil")
	}
	if v.SelectedRecord().Name != "api.example.com" {
		t.Fatalf("expected first record api.example.com, got %s", v.SelectedRecord().Name)
	}

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'p' key")
	}
	msg := cmd()
	if toggleMsg, ok := msg.(views.ToggleDNSProxyReqMsg); !ok || toggleMsg.RecordID != "rec-1" {
		t.Fatalf("expected ToggleDNSProxyReqMsg with rec-1, got %v", msg)
	}

	_, cmd = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'n' key")
	}
	if _, ok := cmd().(views.OpenNewRecordModalMsg); !ok {
		t.Fatalf("expected OpenNewRecordModalMsg, got %v", cmd())
	}

	_, cmd = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'e' key")
	}
	if _, ok := cmd().(views.OpenEditRecordModalMsg); !ok {
		t.Fatalf("expected OpenEditRecordModalMsg, got %v", cmd())
	}

	_, cmd = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'd' key")
	}
	if _, ok := cmd().(views.OpenDeleteRecordModalMsg); !ok {
		t.Fatalf("expected OpenDeleteRecordModalMsg, got %v", cmd())
	}

	_, cmd = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'y' key")
	}
	if yankMsg, ok := cmd().(views.RecordCopiedMsg); !ok || yankMsg.Content != "1.2.3.4" {
		t.Fatalf("expected RecordCopiedMsg with 1.2.3.4, got %v", cmd())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	for _, r := range "type:cname" {
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.SelectedRecord() == nil || v.SelectedRecord().Type != dns.TypeCNAME {
		t.Fatalf("expected filtered record to be CNAME, got %v", v.SelectedRecord())
	}

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty DNS view render")
	}
}

func TestTunnelView(t *testing.T) {
	v := views.NewTunnelView()
	v.SetDimensions(sampleDimensions())

	now := time.Now()
	tun1, _ := tunnel.NewTunnel(
		"tun-1",
		"acc-1",
		"prod-tunnel",
		tunnel.StatusHealthy,
		&now,
		now.Add(-24*time.Hour),
		true,
		[]tunnel.Connector{
			{
				ID:            "conn-1",
				ClientVersion: "2024.1.0",
				ColoName:      "SFO",
				OriginIP:      "192.168.1.10",
				Status:        "connected",
			},
		},
		[]tunnel.IngressRule{
			{
				Hostname: "api.example.com",
				Path:     "/v1",
				Service:  "http://localhost:8080",
			},
		},
	)

	v.Update(views.TunnelsLoadedMsg{Tunnels: []tunnel.Tunnel{tun1}})

	if v.SelectedTunnel() == nil || v.SelectedTunnel().Name != "prod-tunnel" {
		t.Fatalf("expected selected tunnel prod-tunnel, got %v", v.SelectedTunnel())
	}

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty tunnel view render")
	}
}

func TestWorkersPagesView(t *testing.T) {
	v := views.NewWorkersPagesView()
	v.SetDimensions(sampleDimensions())

	w1, _ := worker.NewWorkerScript("edge-router", time.Now(), time.Now(), "standard", true, true)
	p1, _ := pages.NewPagesProject("blog", "my-blog", "blog.pages.dev", "main", []string{"blog.example.com"}, time.Now())

	v.Update(views.WorkersLoadedMsg{Workers: []worker.WorkerScript{w1}})
	v.Update(views.PagesLoadedMsg{Pages: []pages.PagesProject{p1}})

	if v.SelectedWorker() == nil || string(v.SelectedWorker().ID) != "edge-router" {
		t.Fatalf("expected selected worker edge-router, got %v", v.SelectedWorker())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.SelectedPage() == nil || v.SelectedPage().Name != "my-blog" {
		t.Fatalf("expected selected page my-blog, got %v", v.SelectedPage())
	}

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty workers pages render")
	}
}

func TestRulesetsView(t *testing.T) {
	v := views.NewRulesetsView()
	v.SetDimensions(sampleDimensions())

	r1, _ := ruleset.NewRuleset("rs-1", "Default WAF", "http_request_firewall_custom", "zone", time.Now(), "Protects web apps", 5)
	v.Update(views.RulesetsLoadedMsg{Rulesets: []ruleset.Ruleset{r1}})

	if v.SelectedRuleset() == nil || v.SelectedRuleset().Name != "Default WAF" {
		t.Fatalf("expected selected ruleset Default WAF, got %v", v.SelectedRuleset())
	}

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty rulesets render")
	}
}

func TestAnalyticsView(t *testing.T) {
	v := views.NewAnalyticsView()
	v.SetDimensions(sampleDimensions())

	summary := analytics.TrafficSummary{
		TotalRequests:  150000,
		CachedRequests: 120000,
		BandwidthBytes: 1024 * 1024 * 512,
		StatusCodes: map[string]int64{
			"200": 130000,
			"301": 5000,
			"404": 10000,
			"500": 5000,
		},
		ThreatCount: 42,
		Timestamp:   time.Now(),
	}

	v.Update(views.AnalyticsLoadedMsg{Summary: summary})

	rendered := v.View()
	if len(rendered) == 0 {
		t.Fatal("expected non-empty analytics render")
	}
}
