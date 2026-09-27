package main

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/ports"
	"github.com/sekai-labs/kumo/internal/presentation/tui/app"
	"github.com/sekai-labs/kumo/internal/presentation/tui/app/testhelpers"
)

func TestSmokeE2E_Lifecycle(t *testing.T) {
	appService := testhelpers.NewTestAppService()

	appConfig := app.Config{
		Service:        appService,
		Profile:        "test-profile",
		InitialAccount: "acc-1",
		InitialZone:    "zone-1",
	}

	model := app.New(appConfig)

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected Init to return initial command")
	}

	model.Update(tea.WindowSizeMsg{Width: 140, Height: 45})

	time.Sleep(50 * time.Millisecond)

	viewStr := model.View()
	if !strings.Contains(viewStr, "Overview") && !strings.Contains(viewStr, "kumo") {
		t.Errorf("expected view to render overview or header, got: %s", viewStr)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	viewStr = model.View()
	if !strings.Contains(viewStr, "DNS") {
		t.Errorf("expected view to contain 'DNS', got: %s", viewStr)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	viewStr = model.View()
	if !strings.Contains(viewStr, "Tunnel") {
		t.Errorf("expected view to contain 'Tunnel', got: %s", viewStr)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	viewStr = model.View()
	if !strings.Contains(viewStr, "Help") && !strings.Contains(viewStr, "Shortcuts") {
		t.Errorf("expected help modal overlay, got: %s", viewStr)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	model.Update(tea.WindowSizeMsg{Width: 95, Height: 30})
	viewStr = model.View()
	if len(viewStr) == 0 {
		t.Fatal("expected non-empty view at medium size")
	}

	model.Update(tea.WindowSizeMsg{Width: 70, Height: 24})
	viewStr = model.View()
	if len(viewStr) == 0 {
		t.Fatal("expected non-empty view at compact size")
	}

	_, quitCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if quitCmd == nil {
		t.Errorf("expected quitCmd on 'q', got nil")
	}
}

func TestSmokeE2E_AppService(t *testing.T) {
	ctx := context.Background()
	appService := testhelpers.NewTestAppService()

	tok, err := appService.VerifyToken(ctx, false)
	if err != nil || !tok.IsActive() {
		t.Fatalf("token verification failed: %v", err)
	}

	zones, err := appService.ListZones(ctx, "acc-1", false)
	if err != nil || len(zones) == 0 {
		t.Fatalf("list zones failed: %v", err)
	}

	dnsRecs, err := appService.ListDNSRecords(ctx, zones[0].ID, ports.DNSFilter{}, false)
	if err != nil || len(dnsRecs) == 0 {
		t.Fatalf("list DNS records failed: %v", err)
	}

	tunnels, err := appService.ListTunnels(ctx, "acc-1", false)
	if err != nil || len(tunnels) == 0 {
		t.Fatalf("list tunnels failed: %v", err)
	}
}
