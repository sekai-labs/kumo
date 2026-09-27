package theme_test

import (
	"strings"
	"testing"

	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

func TestThemeColorsAndSymbols(t *testing.T) {
	th := theme.DefaultTheme()
	if th == nil {
		t.Fatal("expected non-nil default theme")
	}

	healthy := theme.BadgeHealthy("")
	if !strings.Contains(healthy, theme.SymbolHealthy) || !strings.Contains(healthy, "Healthy") {
		t.Errorf("unexpected healthy badge: %q", healthy)
	}

	proxied := theme.BadgeProxied()
	if !strings.Contains(proxied, theme.SymbolHealthy) || !strings.Contains(proxied, "Proxied") {
		t.Errorf("unexpected proxied badge: %q", proxied)
	}

	degraded := theme.BadgeDegraded("")
	if !strings.Contains(degraded, theme.SymbolDegraded) || !strings.Contains(degraded, "Degraded") {
		t.Errorf("unexpected degraded badge: %q", degraded)
	}

	dnsOnly := theme.BadgeDNSOnly()
	if !strings.Contains(dnsOnly, theme.SymbolDegraded) || !strings.Contains(dnsOnly, "DNS Only") {
		t.Errorf("unexpected dnsOnly badge: %q", dnsOnly)
	}

	down := theme.BadgeDown("")
	if !strings.Contains(down, theme.SymbolDown) || !strings.Contains(down, "Down") {
		t.Errorf("unexpected down badge: %q", down)
	}

	errBadge := theme.BadgeError("")
	if !strings.Contains(errBadge, theme.SymbolError) || !strings.Contains(errBadge, "Error") {
		t.Errorf("unexpected error badge: %q", errBadge)
	}

	autoBadge := theme.BadgeAuto("")
	if !strings.Contains(autoBadge, theme.SymbolAuto) || !strings.Contains(autoBadge, "Auto") {
		t.Errorf("unexpected auto badge: %q", autoBadge)
	}

	keyBadge := theme.FormatKeyBadge("Tab")
	if !strings.Contains(keyBadge, "[Tab]") {
		t.Errorf("unexpected key badge: %q", keyBadge)
	}
}
