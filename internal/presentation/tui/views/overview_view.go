package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type OverviewView struct {
	zones      []zone.Zone
	tunnels    []tunnel.Tunnel
	dnsRecords []dns.DNSRecord
	workers    []worker.WorkerScript
	pages      []pages.PagesProject
	activeZone zone.Zone
	activeAcc  account.Account
	tokenInfo  account.TokenInfo
	tally      tunnel.HealthTally
	width      int
	height     int
	focused    bool
}

func NewOverviewView() *OverviewView {
	return &OverviewView{
		zones:      make([]zone.Zone, 0),
		tunnels:    make([]tunnel.Tunnel, 0),
		dnsRecords: make([]dns.DNSRecord, 0),
		workers:    make([]worker.WorkerScript, 0),
		pages:      make([]pages.PagesProject, 0),
		width:      80,
		height:     24,
	}
}

func (v *OverviewView) Init() tea.Cmd {
	return nil
}

func (v *OverviewView) SetDimensions(dims layout.Dimensions) {

	contentWidth := dims.List.Width + dims.Detail.Width
	if contentWidth <= 0 {
		contentWidth = dims.TotalWidth - dims.Nav.Width
	}
	if contentWidth <= 0 {
		contentWidth = 80
	}
	v.width = contentWidth
	v.height = dims.List.Height
	if v.height <= 0 {
		v.height = 20
	}
}

func (v *OverviewView) Shortcuts() []Shortcut {
	return []Shortcut{
		{Key: "r", Desc: "refresh"},
		{Key: "2", Desc: "view dns"},
		{Key: "3", Desc: "view tunnels"},
	}
}

func (v *OverviewView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case ZonesLoadedMsg:
		v.zones = msg.Zones
	case TunnelsLoadedMsg:
		v.tunnels = msg.Tunnels
		v.tally = tunnel.CalculateHealthTally(v.tunnels)
	case DNSLoadedMsg:
		v.dnsRecords = msg.Records
	case WorkersLoadedMsg:
		v.workers = msg.Workers
	case PagesLoadedMsg:
		v.pages = msg.Pages
	case ActiveZoneChangedMsg:
		v.activeZone = msg.Zone
	case ActiveAccountChangedMsg:
		v.activeAcc = msg.Account
	case TokenVerifiedMsg:
		v.tokenInfo = msg.TokenInfo
	}
	return v, nil
}

func (v *OverviewView) View() string {
	th := theme.Current()

	var b strings.Builder

	title := th.Title.Render("SYSTEM OVERVIEW DASHBOARD")
	subtitle := th.Muted.Render("Real-time summary of Cloudflare edge assets")
	b.WriteString(title + "\n" + subtitle + "\n\n")

	cardGap := 2
	availWidth := v.width - 4
	if availWidth < 30 {
		availWidth = 30
	}

	cardWidth := (availWidth - cardGap) / 2
	if cardWidth < 28 {
		cardWidth = availWidth
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Background(theme.ColorSurface).
		Padding(1, 2).
		Width(cardWidth)

	card1Content := fmt.Sprintf("%s\n\n%s: %s\n%s: %d\n%s: %s",
		th.Header.Render("Active Zone & Accounts"),
		th.StatusKey.Render("Zone"), th.StatusValue.Render(fallback(v.activeZone.Name, "None selected")),
		th.StatusKey.Render("Total Zones"), len(v.zones),
		th.StatusKey.Render("Plan"), th.Highlight.Render(fallback(v.activeZone.Plan, "Free")),
	)
	card1 := cardStyle.Render(card1Content)

	card2Content := fmt.Sprintf("%s\n\n%s: %d\n%s: %d\n%s: %d\n%s: %d",
		th.Header.Render("Cloudflare Tunnels"),
		theme.BadgeHealthy("Healthy"), v.tally.Healthy,
		theme.BadgeDegraded("Degraded"), v.tally.Degraded,
		theme.BadgeDown("Down"), v.tally.Down,
		th.StatusKey.Render("Total Tunnels"), v.tally.Total,
	)
	card2 := cardStyle.Render(card2Content)

	card3Content := fmt.Sprintf("%s\n\n%s: %d\n%s: %d\n%s: %d",
		th.Header.Render("DNS Records"),
		th.StatusKey.Render("Total Records"), len(v.dnsRecords),
		th.StatusKey.Render("Proxied"), countProxied(v.dnsRecords),
		th.StatusKey.Render("DNS-Only"), len(v.dnsRecords)-countProxied(v.dnsRecords),
	)
	card3 := cardStyle.Render(card3Content)

	card4Content := fmt.Sprintf("%s\n\n%s: %d\n%s: %d\n%s: %d",
		th.Header.Render("Compute & Hosting"),
		th.StatusKey.Render("Workers Scripts"), len(v.workers),
		th.StatusKey.Render("Pages Projects"), len(v.pages),
		th.StatusKey.Render("Total Deployments"), len(v.workers)+len(v.pages),
	)
	card4 := cardStyle.Render(card4Content)

	if cardWidth == availWidth {
		b.WriteString(card1 + "\n\n" + card2 + "\n\n" + card3 + "\n\n" + card4 + "\n")
	} else {
		row1 := lipgloss.JoinHorizontal(lipgloss.Top, card1, strings.Repeat(" ", cardGap), card2)
		row2 := lipgloss.JoinHorizontal(lipgloss.Top, card3, strings.Repeat(" ", cardGap), card4)
		b.WriteString(row1 + "\n\n" + row2 + "\n")
	}

	b.WriteString("\n" + th.Subtitle.Render("Quick Actions & Navigation") + "\n")
	b.WriteString(th.Muted.Render("Press [2] for DNS Manager, [3] for Tunnels, [4] for Workers & Pages, [:] for Command Palette") + "\n")

	paneWidth := v.width - 2
	if paneWidth < 0 {
		paneWidth = 0
	}
	paneHeight := v.height - 2
	if paneHeight < 0 {
		paneHeight = 0
	}

	return th.Pane.
		Width(paneWidth).
		Height(paneHeight).
		Render(b.String())
}

func countProxied(records []dns.DNSRecord) int {
	c := 0
	for _, r := range records {
		if r.Proxied {
			c++
		}
	}
	return c
}

func fallback(val, def string) string {
	if strings.TrimSpace(val) == "" {
		return def
	}
	return val
}
