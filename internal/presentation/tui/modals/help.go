package modals

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type CloseHelpModalMsg struct{}

type HelpShortcut struct {
	Key  string
	Desc string
}

type HelpCategory struct {
	Title     string
	Shortcuts []HelpShortcut
}

type HelpModal struct {
	Categories []HelpCategory
	Width      int
	Height     int
}

func NewHelpModal() *HelpModal {
	return &HelpModal{
		Width: 70,
		Categories: []HelpCategory{
			{
				Title: "Navigation",
				Shortcuts: []HelpShortcut{
					{Key: "j / ↓", Desc: "Move selection down"},
					{Key: "k / ↑", Desc: "Move selection up"},
					{Key: "h / l", Desc: "Switch pane / tabs"},
					{Key: "Tab / Shift+Tab", Desc: "Cycle through panes"},
					{Key: "Enter", Desc: "Select / Open item"},
					{Key: "Esc", Desc: "Close modal / Clear filter"},
					{Key: "/", Desc: "Fuzzy search / Filter list"},
				},
			},
			{
				Title: "Views",
				Shortcuts: []HelpShortcut{
					{Key: "1", Desc: "System Overview"},
					{Key: "2", Desc: "DNS Records"},
					{Key: "3", Desc: "Cloudflare Tunnels"},
					{Key: "4", Desc: "Workers & Pages"},
					{Key: "5", Desc: "Rulesets"},
					{Key: "6", Desc: "Traffic Analytics"},
					{Key: "Z", Desc: "Switch Active Zone / Account"},
				},
			},
			{
				Title: "Actions",
				Shortcuts: []HelpShortcut{
					{Key: "p", Desc: "Toggle Proxy (Orange/Grey Cloud)"},
					{Key: "n", Desc: "New DNS Record"},
					{Key: "e", Desc: "Edit selected DNS Record"},
					{Key: "d", Desc: "Delete selected DNS Record"},
					{Key: "y", Desc: "Yank / Copy record content"},
				},
			},
			{
				Title: "Global",
				Shortcuts: []HelpShortcut{
					{Key: ":", Desc: "Open Command Palette"},
					{Key: "r", Desc: "Force refresh & bypass cache"},
					{Key: "?", Desc: "Toggle Help Overlay"},
					{Key: "q / Ctrl+C", Desc: "Quit Kumo"},
				},
			},
		},
	}
}

func (m *HelpModal) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *HelpModal) Init() tea.Cmd {
	return nil
}

func (m *HelpModal) Update(msg tea.Msg) (*HelpModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			return m, func() tea.Msg {
				return CloseHelpModalMsg{}
			}
		}
	}
	return m, nil
}

func (m *HelpModal) View() string {
	modalWidth := 72
	if m.Width > 0 && m.Width < 76 {
		modalWidth = m.Width - 4
		if modalWidth < 40 {
			modalWidth = 40
		}
	}
	contentWidth := modalWidth - 4

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFOrange).
		MarginBottom(1)

	catTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFBlue).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFOrange).
		Width(18)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.ColorForeground)

	var sections []string
	sections = append(sections, titleStyle.Render("Kumo Keyboard Shortcuts"))

	for _, cat := range m.Categories {
		sections = append(sections, catTitleStyle.Render(cat.Title))
		for _, sc := range cat.Shortcuts {
			line := lipgloss.JoinHorizontal(
				lipgloss.Left,
				keyStyle.Render(sc.Key),
				descStyle.Render(sc.Desc),
			)
			sections = append(sections, line)
		}
	}

	sections = append(sections, "")
	sections = append(sections, theme.RenderMuted("[Esc/q/?] Close Help"))

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCFBlue).
		Background(theme.ColorSurface).
		Padding(1, 2).
		Width(modalWidth)

	_ = contentWidth
	return boxStyle.Render(strings.TrimSpace(body))
}
