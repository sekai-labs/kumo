package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type TunnelView struct {
	tunnels     []tunnel.Tunnel
	table       table.Model
	width       int
	height      int
	listWidth   int
	detailWidth int
	breakpoint  layout.Breakpoint
}

func NewTunnelView() *TunnelView {
	columns := []table.Column{
		{Title: "Status", Width: 12},
		{Title: "Name", Width: 24},
		{Title: "Connections", Width: 13},
		{Title: "Created", Width: 16},
		{Title: "Config", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.ColorCFBlue)
	s.Selected = s.Selected.
		Foreground(theme.ColorForeground).
		Background(theme.ColorSelection).
		Bold(true)
	t.SetStyles(s)

	return &TunnelView{
		tunnels:     make([]tunnel.Tunnel, 0),
		table:       t,
		width:       100,
		height:      24,
		listWidth:   60,
		detailWidth: 38,
		breakpoint:  layout.BreakpointWide,
	}
}

func (v *TunnelView) Init() tea.Cmd {
	return nil
}

func (v *TunnelView) SetDimensions(dims layout.Dimensions) {
	v.breakpoint = dims.Breakpoint
	v.width = dims.TotalWidth - dims.Nav.Width
	v.height = dims.List.Height

	if v.width <= 0 {
		v.width = 80
	}
	if v.height <= 0 {
		v.height = 20
	}

	if dims.Breakpoint == layout.BreakpointCompact {
		v.listWidth = v.width
		v.detailWidth = 0
	} else {
		v.listWidth = dims.List.Width
		v.detailWidth = dims.Detail.Width
		if v.detailWidth <= 0 && v.width > 60 {
			v.detailWidth = v.width * 40 / 100
			v.listWidth = v.width - v.detailWidth
		}
	}

	v.updateTableDimensions()
}

func (v *TunnelView) updateTableDimensions() {
	tableWidth := v.listWidth - 4
	if tableWidth < 40 {
		tableWidth = 40
	}

	statusW := 12
	connsW := 13
	createdW := 14
	configW := 9
	remain := tableWidth - statusW - connsW - createdW - configW - 6
	if remain < 15 {
		remain = 15
	}
	nameW := remain

	v.table.SetColumns([]table.Column{
		{Title: "Status", Width: statusW},
		{Title: "Name", Width: nameW},
		{Title: "Active Conns", Width: connsW},
		{Title: "Created", Width: createdW},
		{Title: "Config", Width: configW},
	})

	tableHeight := v.height - 4
	if tableHeight < 3 {
		tableHeight = 3
	}
	v.table.SetHeight(tableHeight)
}

func (v *TunnelView) Shortcuts() []Shortcut {
	return []Shortcut{
		{Key: "↑/↓", Desc: "select tunnel"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *TunnelView) SelectedTunnel() *tunnel.Tunnel {
	idx := v.table.Cursor()
	if idx >= 0 && idx < len(v.tunnels) {
		return &v.tunnels[idx]
	}
	return nil
}

func (v *TunnelView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case TunnelsLoadedMsg:
		v.tunnels = msg.Tunnels
		v.populateTable()
		return v, nil
	}

	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *TunnelView) populateTable() {
	rows := make([]table.Row, len(v.tunnels))
	for i, tun := range v.tunnels {
		var statusStr string
		switch tun.Status {
		case tunnel.StatusHealthy:
			statusStr = theme.SymbolHealthy + " Healthy"
		case tunnel.StatusDegraded:
			statusStr = theme.SymbolDegraded + " Degraded"
		case tunnel.StatusDown:
			statusStr = theme.SymbolDown + " Down"
		default:
			statusStr = theme.SymbolDown + " Inactive"
		}

		connsCount := len(tun.Connectors)
		connsStr := fmt.Sprintf("%d active", connsCount)

		createdStr := tun.CreatedAt.Format("2006-01-02")
		if tun.CreatedAt.IsZero() {
			createdStr = "—"
		}

		cfgStr := "Local"
		if tun.RemoteConfig {
			cfgStr = "Remote"
		}

		rows[i] = table.Row{
			statusStr,
			tun.Name,
			connsStr,
			createdStr,
			cfgStr,
		}
	}
	v.table.SetRows(rows)
}

func (v *TunnelView) View() string {
	th := theme.Current()

	var listBuilder strings.Builder
	titleBar := th.Title.Render(fmt.Sprintf("CLOUDFLARE TUNNELS (%d)", len(v.tunnels)))
	listBuilder.WriteString(titleBar + "\n\n")
	listBuilder.WriteString(v.table.View())

	listPaneWidth := v.listWidth - 2
	if listPaneWidth < 0 {
		listPaneWidth = 0
	}
	listPaneHeight := v.height - 2
	if listPaneHeight < 0 {
		listPaneHeight = 0
	}

	listPane := th.Pane.
		Width(listPaneWidth).
		Height(listPaneHeight).
		Render(listBuilder.String())

	if v.breakpoint == layout.BreakpointCompact || v.detailWidth <= 0 {
		return listPane
	}

	detailPane := v.renderDetailPane()
	return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

func (v *TunnelView) renderDetailPane() string {
	th := theme.Current()
	var b strings.Builder

	b.WriteString(th.Header.Render("TUNNEL DETAILS") + "\n\n")

	sel := v.SelectedTunnel()
	if sel == nil {
		b.WriteString(th.Muted.Render("No tunnel selected."))
	} else {
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Name"), th.Highlight.Render(sel.Name)))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("ID"), th.Muted.Render(string(sel.ID))))

		var statusBadge string
		switch sel.Status {
		case tunnel.StatusHealthy:
			statusBadge = theme.BadgeHealthy(string(sel.Status))
		case tunnel.StatusDegraded:
			statusBadge = theme.BadgeDegraded(string(sel.Status))
		default:
			statusBadge = theme.BadgeDown(string(sel.Status))
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Health"), statusBadge))

		cfgType := "Locally Managed"
		if sel.RemoteConfig {
			cfgType = "Remotely Managed (Cloudflare Dashboard/Zero Trust)"
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Config"), cfgType))

		if sel.ConnsActiveAt != nil && !sel.ConnsActiveAt.IsZero() {
			age := time.Since(*sel.ConnsActiveAt).Round(time.Second)
			b.WriteString(fmt.Sprintf("%s: %s ago\n", th.StatusKey.Render("Active Since"), age))
		}

		b.WriteString("\n" + th.Subtitle.Render(fmt.Sprintf("Connectors (%d):", len(sel.Connectors))) + "\n")
		if len(sel.Connectors) == 0 {
			b.WriteString(th.Muted.Render("  No active connectors connected.") + "\n")
		} else {
			for _, conn := range sel.Connectors {
				connBadge := theme.SymbolHealthy
				if strings.ToLower(conn.Status) != "connected" && strings.ToLower(conn.Status) != "healthy" {
					connBadge = theme.SymbolDown
				}
				b.WriteString(fmt.Sprintf("  %s %s (Colo: %s, Ver: %s, Origin: %s)\n",
					connBadge,
					truncate(conn.ID, 8),
					fallback(conn.ColoName, "N/A"),
					fallback(conn.ClientVersion, "N/A"),
					fallback(conn.OriginIP, "N/A"),
				))
			}
		}

		b.WriteString("\n" + th.Subtitle.Render(fmt.Sprintf("Ingress Rules (%d):", len(sel.IngressRules))) + "\n")
		if len(sel.IngressRules) == 0 {
			b.WriteString(th.Muted.Render("  No ingress rules configured.") + "\n")
		} else {
			for _, rule := range sel.IngressRules {
				host := rule.Hostname
				if host == "" {
					host = "*"
				}
				if rule.Path != "" {
					host += rule.Path
				}
				b.WriteString(fmt.Sprintf("  • %s → %s\n",
					th.Highlight.Render(host),
					th.Muted.Render(rule.Service),
				))
			}
		}
	}

	detailPaneWidth := v.detailWidth - 2
	if detailPaneWidth < 0 {
		detailPaneWidth = 0
	}
	detailPaneHeight := v.height - 2
	if detailPaneHeight < 0 {
		detailPaneHeight = 0
	}

	return th.Pane.
		Width(detailPaneWidth).
		Height(detailPaneHeight).
		Render(b.String())
}
