package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type AnalyticsView struct {
	summary analytics.TrafficSummary
	width   int
	height  int
}

func NewAnalyticsView() *AnalyticsView {
	return &AnalyticsView{
		summary: analytics.TrafficSummary{
			StatusCodes: make(map[string]int64),
		},
		width:  100,
		height: 24,
	}
}

func (v *AnalyticsView) Init() tea.Cmd {
	return nil
}

func (v *AnalyticsView) SetDimensions(dims layout.Dimensions) {
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

func (v *AnalyticsView) Shortcuts() []Shortcut {
	return []Shortcut{
		{Key: "r", Desc: "refresh analytics"},
	}
}

func (v *AnalyticsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case AnalyticsLoadedMsg:
		v.summary = msg.Summary
		return v, nil
	}
	return v, nil
}

func (v *AnalyticsView) View() string {
	th := theme.Current()
	var b strings.Builder

	title := th.Title.Render("EDGE TRAFFIC ANALYTICS & STATUS CODES")
	subtitle := th.Muted.Render("Real-time aggregated metrics (Last 24 Hours)")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	cardWidth := (v.width - 10) / 4
	if cardWidth < 18 {
		cardWidth = 18
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorder).
		Background(theme.ColorSurface).
		Padding(0, 1).
		Width(cardWidth)

	c1 := cardStyle.Render(fmt.Sprintf("%s\n%s",
		th.Muted.Render("Total Requests"),
		th.Highlight.Render(fmt.Sprintf("%d", v.summary.TotalRequests)),
	))

	hitRatio := v.summary.CacheHitRatio()
	ratioColor := theme.ColorStatusHealthy
	if hitRatio < 50.0 {
		ratioColor = theme.ColorStatusDegraded
	}
	ratioStyle := lipgloss.NewStyle().Foreground(ratioColor).Bold(true)

	c2 := cardStyle.Render(fmt.Sprintf("%s\n%s (%d cached)",
		th.Muted.Render("Cache Hit Ratio"),
		ratioStyle.Render(fmt.Sprintf("%.1f%%", hitRatio)),
		v.summary.CachedRequests,
	))

	c3 := cardStyle.Render(fmt.Sprintf("%s\n%s",
		th.Muted.Render("Bandwidth Served"),
		th.Highlight.Render(formatBytes(v.summary.BandwidthBytes)),
	))

	threatColor := theme.ColorStatusHealthy
	if v.summary.ThreatCount > 0 {
		threatColor = theme.ColorStatusDown
	}
	threatStyle := lipgloss.NewStyle().Foreground(threatColor).Bold(true)

	c4 := cardStyle.Render(fmt.Sprintf("%s\n%s threats",
		th.Muted.Render("Threats Mitigated"),
		threatStyle.Render(fmt.Sprintf("%d", v.summary.ThreatCount)),
	))

	metricsRow := lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2, " ", c3, " ", c4)
	b.WriteString(metricsRow)
	b.WriteString("\n\n")

	b.WriteString(th.Header.Render("HTTP STATUS CODE DISTRIBUTION"))
	b.WriteString("\n\n")

	s2xx := v.summary.StatusGroup("2")
	s3xx := v.summary.StatusGroup("3")
	s4xx := v.summary.StatusGroup("4")
	s5xx := v.summary.StatusGroup("5")

	total := v.summary.TotalRequests
	if total == 0 {
		total = s2xx + s3xx + s4xx + s5xx
	}

	chartWidth := v.width - 30
	if chartWidth < 20 {
		chartWidth = 20
	}
	if chartWidth > 50 {
		chartWidth = 50
	}

	b.WriteString(renderBarRow("2xx Success", s2xx, total, chartWidth, theme.ColorStatusHealthy))
	b.WriteString(renderBarRow("3xx Redirect", s3xx, total, chartWidth, theme.ColorCFBlue))
	b.WriteString(renderBarRow("4xx Client Err", s4xx, total, chartWidth, theme.ColorStatusDegraded))
	b.WriteString(renderBarRow("5xx Server Err", s5xx, total, chartWidth, theme.ColorStatusDown))

	b.WriteString("\n")
	b.WriteString(th.Subtitle.Render("Detailed Status Codes Breakdown:"))
	b.WriteString("\n")
	if len(v.summary.StatusCodes) == 0 {
		b.WriteString(th.Muted.Render("  No status code traffic recorded."))
		b.WriteString("\n")
	} else {
		var codeItems []string
		for code, count := range v.summary.StatusCodes {
			codeColor := theme.ColorForeground
			if strings.HasPrefix(code, "2") {
				codeColor = theme.ColorStatusHealthy
			} else if strings.HasPrefix(code, "3") {
				codeColor = theme.ColorCFBlue
			} else if strings.HasPrefix(code, "4") {
				codeColor = theme.ColorStatusDegraded
			} else if strings.HasPrefix(code, "5") {
				codeColor = theme.ColorStatusDown
			}

			tag := lipgloss.NewStyle().Foreground(codeColor).Bold(true).Render(code)
			codeItems = append(codeItems, fmt.Sprintf("%s: %d", tag, count))
		}
		b.WriteString("  ")
		b.WriteString(strings.Join(codeItems, "   "))
		b.WriteString("\n")
	}

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

func renderBarRow(label string, count, total int64, maxBarLen int, color lipgloss.Color) string {
	th := theme.Current()

	var pct float64
	var barLen int
	if total > 0 {
		pct = (float64(count) / float64(total)) * 100.0
		barLen = int(float64(maxBarLen) * (float64(count) / float64(total)))
	}
	if barLen == 0 && count > 0 {
		barLen = 1
	}

	barStr := strings.Repeat("█", barLen)
	emptyStr := strings.Repeat("░", maxBarLen-barLen)

	barRendered := lipgloss.NewStyle().Foreground(color).Render(barStr) +
		lipgloss.NewStyle().Foreground(theme.ColorBorder).Render(emptyStr)

	labelStyle := th.StatusKey.Copy().Width(15)
	valStyle := th.Body.Copy().Width(12)
	pctStyle := th.Muted.Copy().Width(8)

	return fmt.Sprintf("%s %s %s (%s)\n",
		labelStyle.Render(label),
		barRendered,
		valStyle.Render(fmt.Sprintf("%d reqs", count)),
		pctStyle.Render(fmt.Sprintf("%.1f%%", pct)),
	)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
