package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type DNSView struct {
	records         []dns.DNSRecord
	filteredRecords []dns.DNSRecord
	table           table.Model
	filterInput     textinput.Model
	filtering       bool
	activeZone      zone.Zone
	width           int
	height          int
	listWidth       int
	detailWidth     int
	breakpoint      layout.Breakpoint
}

func NewDNSView() *DNSView {
	columns := []table.Column{
		{Title: "Type", Width: 7},
		{Title: "Name", Width: 26},
		{Title: "Content", Width: 28},
		{Title: "Proxy Status", Width: 14},
		{Title: "TTL", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	th := theme.Current()
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

	ti := textinput.New()
	ti.Placeholder = "Filter records (type:, name:, proxied:true/false, or search text)..."
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.ColorCFOrange).Bold(true)

	_ = th

	return &DNSView{
		records:         make([]dns.DNSRecord, 0),
		filteredRecords: make([]dns.DNSRecord, 0),
		table:           t,
		filterInput:     ti,
		filtering:       false,
		width:           100,
		height:          24,
		listWidth:       60,
		detailWidth:     38,
		breakpoint:      layout.BreakpointWide,
	}
}

func (v *DNSView) Init() tea.Cmd {
	return nil
}

func (v *DNSView) SetDimensions(dims layout.Dimensions) {
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
			v.detailWidth = v.width * 38 / 100
			v.listWidth = v.width - v.detailWidth
		}
	}

	v.updateTableDimensions()
}

func (v *DNSView) updateTableDimensions() {
	tableWidth := v.listWidth - 4
	if tableWidth < 40 {
		tableWidth = 40
	}

	typeW := 7
	proxyW := 14
	ttlW := 8
	remain := tableWidth - typeW - proxyW - ttlW - 6
	if remain < 20 {
		remain = 20
	}
	nameW := remain * 45 / 100
	contentW := remain - nameW

	v.table.SetColumns([]table.Column{
		{Title: "Type", Width: typeW},
		{Title: "Name", Width: nameW},
		{Title: "Content", Width: contentW},
		{Title: "Proxy Status", Width: proxyW},
		{Title: "TTL", Width: ttlW},
	})

	tableHeight := v.height - 6
	if tableHeight < 3 {
		tableHeight = 3
	}
	v.table.SetHeight(tableHeight)
}

func (v *DNSView) Shortcuts() []Shortcut {
	if v.filtering {
		return []Shortcut{
			{Key: "Enter", Desc: "apply filter"},
			{Key: "Esc", Desc: "cancel filter"},
		}
	}
	return []Shortcut{
		{Key: "/", Desc: "filter"},
		{Key: "p", Desc: "toggle proxy"},
		{Key: "n", Desc: "new"},
		{Key: "e", Desc: "edit"},
		{Key: "d", Desc: "delete"},
		{Key: "y", Desc: "yank content"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *DNSView) SelectedRecord() *dns.DNSRecord {
	idx := v.table.Cursor()
	if idx >= 0 && idx < len(v.filteredRecords) {
		return &v.filteredRecords[idx]
	}
	return nil
}

func (v *DNSView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case DNSLoadedMsg:
		v.records = msg.Records
		v.applyFilter()
		return v, nil

	case ActiveZoneChangedMsg:
		v.activeZone = msg.Zone
		return v, nil

	case tea.KeyMsg:
		if v.filtering {
			switch msg.String() {
			case "enter":
				v.filtering = false
				v.table.Focus()
				return v, nil
			case "esc":
				v.filtering = false
				v.filterInput.SetValue("")
				v.applyFilter()
				v.table.Focus()
				return v, nil
			default:
				v.filterInput, cmd = v.filterInput.Update(msg)
				v.applyFilter()
				return v, cmd
			}
		}

		switch msg.String() {
		case "/":
			v.filtering = true
			v.filterInput.Focus()
			return v, textinput.Blink

		case "p":
			sel := v.SelectedRecord()
			if sel != nil && sel.Proxiable {
				return v, func() tea.Msg {
					return ToggleDNSProxyReqMsg{
						ZoneID:   v.activeZone.ID,
						RecordID: sel.ID,
					}
				}
			}
			return v, nil

		case "n":
			return v, func() tea.Msg {
				return OpenNewRecordModalMsg{ZoneID: v.activeZone.ID}
			}

		case "e":
			sel := v.SelectedRecord()
			if sel != nil {
				return v, func() tea.Msg {
					return OpenEditRecordModalMsg{Record: *sel}
				}
			}
			return v, nil

		case "d":
			sel := v.SelectedRecord()
			if sel != nil {
				return v, func() tea.Msg {
					return OpenDeleteRecordModalMsg{Record: *sel}
				}
			}
			return v, nil

		case "y":
			sel := v.SelectedRecord()
			if sel != nil {
				return v, func() tea.Msg {
					return RecordCopiedMsg{Content: sel.Content}
				}
			}
			return v, nil
		}
	}

	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *DNSView) applyFilter() {
	query := strings.TrimSpace(strings.ToLower(v.filterInput.Value()))
	if query == "" {
		v.filteredRecords = v.records
	} else {
		var filtered []dns.DNSRecord
		for _, r := range v.records {
			match := false

			if strings.HasPrefix(query, "type:") {
				t := strings.TrimPrefix(query, "type:")
				if strings.EqualFold(string(r.Type), t) {
					match = true
				}
			} else if strings.HasPrefix(query, "name:") {
				n := strings.TrimPrefix(query, "name:")
				if strings.Contains(strings.ToLower(r.Name), n) {
					match = true
				}
			} else if strings.HasPrefix(query, "proxied:") {
				p := strings.TrimPrefix(query, "proxied:")
				if p == "true" && r.Proxied {
					match = true
				} else if p == "false" && !r.Proxied {
					match = true
				}
			} else {

				if strings.Contains(strings.ToLower(r.Name), query) ||
					strings.Contains(strings.ToLower(string(r.Type)), query) ||
					strings.Contains(strings.ToLower(r.Content), query) ||
					strings.Contains(strings.ToLower(r.Comment), query) {
					match = true
				}
			}

			if match {
				filtered = append(filtered, r)
			}
		}
		v.filteredRecords = filtered
	}

	rows := make([]table.Row, len(v.filteredRecords))
	for i, r := range v.filteredRecords {
		var proxyBadge string
		if r.Proxied {
			proxyBadge = theme.SymbolHealthy + " Proxied"
		} else {
			proxyBadge = theme.SymbolDegraded + " DNS Only"
		}

		ttlStr := strconv.Itoa(int(r.TTL))
		if r.TTL == dns.TTLAuto {
			ttlStr = "Auto"
		}

		rows[i] = table.Row{
			string(r.Type),
			r.Name,
			r.Content,
			proxyBadge,
			ttlStr,
		}
	}
	v.table.SetRows(rows)
}

func (v *DNSView) View() string {
	th := theme.Current()

	var listBuilder strings.Builder
	titleBar := th.Title.Render(fmt.Sprintf("DNS RECORDS — %s (%d)",
		fallback(v.activeZone.Name, "All"), len(v.filteredRecords)))
	listBuilder.WriteString(titleBar)
	listBuilder.WriteString("\n")

	if v.filtering {
		listBuilder.WriteString(v.filterInput.View())
		listBuilder.WriteString("\n")
	} else if v.filterInput.Value() != "" {
		filterTag := th.TagBadge.Render("filter: " + v.filterInput.Value())
		listBuilder.WriteString(filterTag)
		listBuilder.WriteString("\n")
	} else {
		listBuilder.WriteString(th.Muted.Render("Type '/' to filter records..."))
		listBuilder.WriteString("\n")
	}

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

func (v *DNSView) renderDetailPane() string {
	th := theme.Current()
	var b strings.Builder

	b.WriteString(th.Header.Render("RECORD DETAILS"))
	b.WriteString("\n\n")

	sel := v.SelectedRecord()
	if sel == nil {
		b.WriteString(th.Muted.Render("No record selected."))
	} else {
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Type"), th.Highlight.Render(string(sel.Type))))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Name"), sel.Name))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Content"), sel.Content))

		var proxiedBadge string
		if sel.Proxied {
			proxiedBadge = theme.BadgeProxied()
		} else {
			proxiedBadge = theme.BadgeDNSOnly()
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Proxy"), proxiedBadge))

		ttlStr := strconv.Itoa(int(sel.TTL))
		if sel.TTL == dns.TTLAuto {
			ttlStr = "Auto (1)"
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("TTL"), ttlStr))

		proxiableStr := "No"
		if sel.Proxiable {
			proxiableStr = "Yes"
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Proxiable"), proxiableStr))

		if sel.Priority != nil {
			b.WriteString(fmt.Sprintf("%s: %d\n", th.StatusKey.Render("Priority"), *sel.Priority))
		}

		if sel.Comment != "" {
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Comment"), sel.Comment))
		}

		if !sel.ModifiedOn.IsZero() {
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Modified"), sel.ModifiedOn.Format("2006-01-02 15:04:05")))
		}

		b.WriteString(fmt.Sprintf("\n%s: %s\n", th.StatusKey.Render("Record ID"), th.Muted.Render(string(sel.ID))))

		b.WriteString("\n")
		b.WriteString(th.Subtitle.Render("Available Actions:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s %s\n", th.KeyBadge.Render("[p]"), th.Muted.Render("Toggle Proxy")))
		b.WriteString(fmt.Sprintf("%s %s\n", th.KeyBadge.Render("[e]"), th.Muted.Render("Edit Record")))
		b.WriteString(fmt.Sprintf("%s %s\n", th.KeyBadge.Render("[d]"), th.Muted.Render("Delete Record")))
		b.WriteString(fmt.Sprintf("%s %s\n", th.KeyBadge.Render("[y]"), th.Muted.Render("Copy Content")))
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
