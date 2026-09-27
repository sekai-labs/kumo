package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type RulesetsView struct {
	rulesets    []ruleset.Ruleset
	table       table.Model
	width       int
	height      int
	listWidth   int
	detailWidth int
	breakpoint  layout.Breakpoint
}

func NewRulesetsView() *RulesetsView {
	columns := []table.Column{
		{Title: "Name", Width: 26},
		{Title: "Phase", Width: 28},
		{Title: "Kind", Width: 10},
		{Title: "Rules", Width: 8},
		{Title: "Updated", Width: 16},
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

	return &RulesetsView{
		rulesets:    make([]ruleset.Ruleset, 0),
		table:       t,
		width:       100,
		height:      24,
		listWidth:   60,
		detailWidth: 38,
		breakpoint:  layout.BreakpointWide,
	}
}

func (v *RulesetsView) Init() tea.Cmd {
	return nil
}

func (v *RulesetsView) SetDimensions(dims layout.Dimensions) {
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

func (v *RulesetsView) updateTableDimensions() {
	contentWidth := v.listWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	kindW := 8
	rulesW := 8
	updateW := 14
	remain := contentWidth - kindW - rulesW - updateW - 6
	if remain < 20 {
		remain = 20
	}
	nameW := remain * 45 / 100
	phaseW := remain - nameW

	v.table.SetColumns([]table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Phase", Width: phaseW},
		{Title: "Kind", Width: kindW},
		{Title: "Rules", Width: rulesW},
		{Title: "Updated", Width: updateW},
	})

	tableHeight := v.height - 4
	if tableHeight < 3 {
		tableHeight = 3
	}
	v.table.SetHeight(tableHeight)
}

func (v *RulesetsView) Shortcuts() []Shortcut {
	return []Shortcut{
		{Key: "↑/↓", Desc: "select ruleset"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *RulesetsView) SelectedRuleset() *ruleset.Ruleset {
	idx := v.table.Cursor()
	if idx >= 0 && idx < len(v.rulesets) {
		return &v.rulesets[idx]
	}
	return nil
}

func (v *RulesetsView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case RulesetsLoadedMsg:
		v.rulesets = msg.Rulesets
		v.populateTable()
		return v, nil
	}

	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v *RulesetsView) populateTable() {
	rows := make([]table.Row, len(v.rulesets))
	for i, r := range v.rulesets {
		updateStr := r.LastUpdated.Format("2006-01-02")
		if r.LastUpdated.IsZero() {
			updateStr = "—"
		}

		rows[i] = table.Row{
			r.Name,
			r.Phase,
			r.Kind,
			fmt.Sprintf("%d", r.RulesCount),
			updateStr,
		}
	}
	v.table.SetRows(rows)
}

func (v *RulesetsView) View() string {
	th := theme.Current()

	var listBuilder strings.Builder
	titleBar := th.Title.Render(fmt.Sprintf("ZONE RULESETS (%d)", len(v.rulesets)))
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

func (v *RulesetsView) renderDetailPane() string {
	th := theme.Current()
	var b strings.Builder

	b.WriteString(th.Header.Render("RULESET DETAILS") + "\n\n")

	sel := v.SelectedRuleset()
	if sel == nil {
		b.WriteString(th.Muted.Render("No ruleset selected."))
	} else {
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Ruleset Name"), th.Highlight.Render(sel.Name)))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Ruleset ID"), th.Muted.Render(string(sel.ID))))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Phase"), sel.Phase))
		b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Kind"), sel.Kind))
		b.WriteString(fmt.Sprintf("%s: %d\n", th.StatusKey.Render("Active Rules"), sel.RulesCount))
		if !sel.LastUpdated.IsZero() {
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Last Updated"), sel.LastUpdated.Format("2006-01-02 15:04:05")))
		}
		if sel.Description != "" {
			b.WriteString(fmt.Sprintf("\n%s:\n%s\n", th.StatusKey.Render("Description"), sel.Description))
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
