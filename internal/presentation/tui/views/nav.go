package views

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
	"strings"
)

type NavItem struct {
	ID       string
	Title    string
	Shortcut string
	Badge    string
}

type Nav struct {
	Items     []NavItem
	ActiveIdx int
	Width     int
	Height    int
	Focused   bool
}

func NewNav() *Nav {
	return &Nav{
		Items: []NavItem{
			{ID: ViewOverview, Title: "Overview", Shortcut: "1"},
			{ID: ViewDNS, Title: "DNS Records", Shortcut: "2"},
			{ID: ViewTunnels, Title: "Tunnels", Shortcut: "3"},
			{ID: ViewWorkers, Title: "Workers & Pages", Shortcut: "4"},
			{ID: ViewRulesets, Title: "Rulesets", Shortcut: "5"},
			{ID: ViewAnalytics, Title: "Analytics", Shortcut: "6"},
		},
		ActiveIdx: 0,
		Width:     24,
		Height:    20,
		Focused:   false,
	}
}

func (n *Nav) SetDimensions(width, height int) {
	n.Width = width
	n.Height = height
}

func (n *Nav) SetFocused(focused bool) {
	n.Focused = focused
}

func (n *Nav) SetBadge(id string, badge string) {
	for i := range n.Items {
		if n.Items[i].ID == id {
			n.Items[i].Badge = badge
			return
		}
	}
}

func (n *Nav) ActiveItem() NavItem {
	if n.ActiveIdx >= 0 && n.ActiveIdx < len(n.Items) {
		return n.Items[n.ActiveIdx]
	}
	return n.Items[0]
}

func (n *Nav) SelectByID(id string) {
	for i, item := range n.Items {
		if item.ID == id {
			n.ActiveIdx = i
			return
		}
	}
}

func (n *Nav) Next() {
	if len(n.Items) == 0 {
		return
	}
	n.ActiveIdx = (n.ActiveIdx + 1) % len(n.Items)
}

func (n *Nav) Prev() {
	if len(n.Items) == 0 {
		return
	}
	n.ActiveIdx = (n.ActiveIdx - 1 + len(n.Items)) % len(n.Items)
}

func (n *Nav) Update(msg tea.Msg) (*Nav, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			n.Prev()
			return n, func() tea.Msg {
				return SelectNavMsg{Index: n.ActiveIdx, ID: n.ActiveItem().ID}
			}
		case "down", "j":
			n.Next()
			return n, func() tea.Msg {
				return SelectNavMsg{Index: n.ActiveIdx, ID: n.ActiveItem().ID}
			}
		case "1", "2", "3", "4", "5", "6":
			idx := int(msg.Runes[0] - '1')
			if idx >= 0 && idx < len(n.Items) {
				n.ActiveIdx = idx
				return n, func() tea.Msg {
					return SelectNavMsg{Index: n.ActiveIdx, ID: n.ActiveItem().ID}
				}
			}
		}
	}
	return n, nil
}

func (n *Nav) View() string {
	th := theme.Current()
	var b strings.Builder

	titleStyle := th.Title.Copy().Padding(0, 1)
	b.WriteString(titleStyle.Render("KUMO CONTROL"))
	b.WriteString("\n\n")

	contentWidth := n.Width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	for i, item := range n.Items {
		isSelected := i == n.ActiveIdx

		keyBadge := th.KeyBadge.Render(item.Shortcut)
		title := item.Title

		var badgeStr string
		if item.Badge != "" {
			badgeStr = " " + th.TagBadge.Render(item.Badge)
		}

		lineContent := fmt.Sprintf("%s %s%s", keyBadge, title, badgeStr)

		var renderedLine string
		if isSelected {
			activeIndicator := "▸ "
			lineStyle := th.SelectedRow.Copy().
				Width(contentWidth).
				PaddingLeft(1)
			renderedLine = lineStyle.Render(activeIndicator + lineContent)
		} else {
			lineStyle := th.InactiveItem.Copy().
				Width(contentWidth).
				PaddingLeft(3)
			renderedLine = lineStyle.Render(lineContent)
		}

		b.WriteString(renderedLine)
		b.WriteString("\n")
	}

	renderedContent := b.String()
	lines := strings.Count(renderedContent, "\n")
	if n.Height > lines+2 {
		b.WriteString(strings.Repeat("\n", n.Height-lines-2))
	}

	paneStyle := th.Pane
	if n.Focused {
		paneStyle = th.PaneFocus
	}

	innerWidth := n.Width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := n.Height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	return paneStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(b.String())
}
