package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type WorkersPagesTab int

const (
	TabWorkers WorkersPagesTab = iota
	TabPages
)

type WorkersPagesView struct {
	activeTab    WorkersPagesTab
	workers      []worker.WorkerScript
	pages        []pages.PagesProject
	workersTable table.Model
	pagesTable   table.Model
	width        int
	height       int
	listWidth    int
	detailWidth  int
	breakpoint   layout.Breakpoint
}

func NewWorkersPagesView() *WorkersPagesView {
	wCols := []table.Column{
		{Title: "Script ID", Width: 26},
		{Title: "Usage Model", Width: 14},
		{Title: "Logpush", Width: 10},
		{Title: "Assets", Width: 10},
		{Title: "Modified", Width: 16},
	}

	pCols := []table.Column{
		{Title: "Project Name", Width: 24},
		{Title: "Subdomain", Width: 26},
		{Title: "Branch", Width: 12},
		{Title: "Domains", Width: 10},
		{Title: "Created", Width: 14},
	}

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

	_ = th

	wt := table.New(table.WithColumns(wCols), table.WithRows([]table.Row{}), table.WithFocused(true), table.WithHeight(10))
	wt.SetStyles(s)

	pt := table.New(table.WithColumns(pCols), table.WithRows([]table.Row{}), table.WithFocused(false), table.WithHeight(10))
	pt.SetStyles(s)

	return &WorkersPagesView{
		activeTab:    TabWorkers,
		workers:      make([]worker.WorkerScript, 0),
		pages:        make([]pages.PagesProject, 0),
		workersTable: wt,
		pagesTable:   pt,
		width:        100,
		height:       24,
		listWidth:    60,
		detailWidth:  38,
		breakpoint:   layout.BreakpointWide,
	}
}

func (v *WorkersPagesView) Init() tea.Cmd {
	return nil
}

func (v *WorkersPagesView) SetDimensions(dims layout.Dimensions) {
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

func (v *WorkersPagesView) updateTableDimensions() {
	tableHeight := v.height - 6
	if tableHeight < 3 {
		tableHeight = 3
	}
	v.workersTable.SetHeight(tableHeight)
	v.pagesTable.SetHeight(tableHeight)

	contentWidth := v.listWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	scriptW := contentWidth * 40 / 100
	modelW := 14
	logW := 10
	assetW := 10
	modW := contentWidth - scriptW - modelW - logW - assetW
	if modW < 12 {
		modW = 12
	}
	v.workersTable.SetColumns([]table.Column{
		{Title: "Script ID", Width: scriptW},
		{Title: "Usage Model", Width: modelW},
		{Title: "Logpush", Width: logW},
		{Title: "Assets", Width: assetW},
		{Title: "Modified", Width: modW},
	})

	nameW := contentWidth * 32 / 100
	subW := contentWidth * 32 / 100
	branchW := 12
	domW := 10
	createW := contentWidth - nameW - subW - branchW - domW
	if createW < 12 {
		createW = 12
	}
	v.pagesTable.SetColumns([]table.Column{
		{Title: "Project Name", Width: nameW},
		{Title: "Subdomain", Width: subW},
		{Title: "Branch", Width: branchW},
		{Title: "Domains", Width: domW},
		{Title: "Created", Width: createW},
	})
}

func (v *WorkersPagesView) Shortcuts() []Shortcut {
	return []Shortcut{
		{Key: "Tab / [ / ]", Desc: "switch tab"},
		{Key: "↑/↓", Desc: "navigate items"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *WorkersPagesView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case WorkersLoadedMsg:
		v.workers = msg.Workers
		v.populateWorkersTable()
		return v, nil

	case PagesLoadedMsg:
		v.pages = msg.Pages
		v.populatePagesTable()
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "[", "]":
			if v.activeTab == TabWorkers {
				v.activeTab = TabPages
				v.workersTable.Blur()
				v.pagesTable.Focus()
			} else {
				v.activeTab = TabWorkers
				v.pagesTable.Blur()
				v.workersTable.Focus()
			}
			return v, nil
		}
	}

	if v.activeTab == TabWorkers {
		v.workersTable, cmd = v.workersTable.Update(msg)
	} else {
		v.pagesTable, cmd = v.pagesTable.Update(msg)
	}

	return v, cmd
}

func (v *WorkersPagesView) populateWorkersTable() {
	rows := make([]table.Row, len(v.workers))
	for i, w := range v.workers {
		logStr := "No"
		if w.Logpush {
			logStr = "Yes"
		}
		assetStr := "No"
		if w.HasAssets {
			assetStr = "Yes"
		}
		modStr := w.ModifiedOn.Format("2006-01-02")
		if w.ModifiedOn.IsZero() {
			modStr = "—"
		}

		rows[i] = table.Row{
			string(w.ID),
			w.UsageModel,
			logStr,
			assetStr,
			modStr,
		}
	}
	v.workersTable.SetRows(rows)
}

func (v *WorkersPagesView) populatePagesTable() {
	rows := make([]table.Row, len(v.pages))
	for i, p := range v.pages {
		domCount := fmt.Sprintf("%d", len(p.Domains))
		createStr := p.CreatedOn.Format("2006-01-02")
		if p.CreatedOn.IsZero() {
			createStr = "—"
		}

		rows[i] = table.Row{
			p.Name,
			p.Subdomain,
			p.ProductionBranch,
			domCount,
			createStr,
		}
	}
	v.pagesTable.SetRows(rows)
}

func (v *WorkersPagesView) SelectedWorker() *worker.WorkerScript {
	idx := v.workersTable.Cursor()
	if idx >= 0 && idx < len(v.workers) {
		return &v.workers[idx]
	}
	return nil
}

func (v *WorkersPagesView) SelectedPage() *pages.PagesProject {
	idx := v.pagesTable.Cursor()
	if idx >= 0 && idx < len(v.pages) {
		return &v.pages[idx]
	}
	return nil
}

func (v *WorkersPagesView) View() string {
	th := theme.Current()

	var tabWorkersStyle, tabPagesStyle lipgloss.Style
	if v.activeTab == TabWorkers {
		tabWorkersStyle = th.SelectedRow.Copy().Padding(0, 2)
		tabPagesStyle = th.InactiveItem.Copy().Padding(0, 2)
	} else {
		tabWorkersStyle = th.InactiveItem.Copy().Padding(0, 2)
		tabPagesStyle = th.SelectedRow.Copy().Padding(0, 2)
	}

	tabBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tabWorkersStyle.Render(fmt.Sprintf("Workers Scripts (%d)", len(v.workers))),
		"  ",
		tabPagesStyle.Render(fmt.Sprintf("Pages Projects (%d)", len(v.pages))),
	)

	var listBuilder strings.Builder
	listBuilder.WriteString(tabBar)
	listBuilder.WriteString("\n\n")

	if v.activeTab == TabWorkers {
		listBuilder.WriteString(v.workersTable.View())
	} else {
		listBuilder.WriteString(v.pagesTable.View())
	}

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

func (v *WorkersPagesView) renderDetailPane() string {
	th := theme.Current()
	var b strings.Builder

	if v.activeTab == TabWorkers {
		b.WriteString(th.Header.Render("WORKER DETAILS") + "\n\n")
		w := v.SelectedWorker()
		if w == nil {
			b.WriteString(th.Muted.Render("No worker script selected."))
		} else {
			fmt.Fprintf(&b, "%s: %s\n", th.StatusKey.Render("Script ID"), th.Highlight.Render(string(w.ID)))
			fmt.Fprintf(&b, "%s: %s\n", th.StatusKey.Render("Usage Model"), w.UsageModel)
			fmt.Fprintf(&b, "%s: %t\n", th.StatusKey.Render("Logpush"), w.Logpush)
			fmt.Fprintf(&b, "%s: %t\n", th.StatusKey.Render("Static Assets"), w.HasAssets)
			if !w.CreatedOn.IsZero() {
				b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Created"), w.CreatedOn.Format("2006-01-02 15:04:05")))
			}
			if !w.ModifiedOn.IsZero() {
				b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Modified"), w.ModifiedOn.Format("2006-01-02 15:04:05")))
			}
		}
	} else {
		b.WriteString(th.Header.Render("PAGES PROJECT DETAILS") + "\n\n")
		p := v.SelectedPage()
		if p == nil {
			b.WriteString(th.Muted.Render("No pages project selected."))
		} else {
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Project Name"), th.Highlight.Render(p.Name)))
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Project ID"), th.Muted.Render(string(p.ID))))
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Subdomain"), p.Subdomain))
			b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Prod Branch"), p.ProductionBranch))
			if !p.CreatedOn.IsZero() {
				b.WriteString(fmt.Sprintf("%s: %s\n", th.StatusKey.Render("Created"), p.CreatedOn.Format("2006-01-02 15:04:05")))
			}
			b.WriteString("\n" + th.Subtitle.Render(fmt.Sprintf("Custom Domains (%d):", len(p.Domains))) + "\n")
			if len(p.Domains) == 0 {
				b.WriteString(th.Muted.Render("  No custom domains attached.") + "\n")
			} else {
				for _, dom := range p.Domains {
					b.WriteString(fmt.Sprintf("  • %s\n", dom))
				}
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
