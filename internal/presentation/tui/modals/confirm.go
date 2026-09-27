package modals

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type ConfirmResultMsg struct {
	ID        string
	Confirmed bool
	Data      any
}

type ConfirmModal struct {
	ID      string
	Title   string
	Message string
	Data    any
	Width   int
	Height  int
	focused bool
	th      *theme.Theme
}

func NewConfirmModal(id, title, message string, data any) *ConfirmModal {
	return &ConfirmModal{
		ID:      id,
		Title:   title,
		Message: message,
		Data:    data,
		Width:   56,
		focused: true,
		th:      theme.Current(),
	}
}

func (m *ConfirmModal) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *ConfirmModal) Init() tea.Cmd {
	return nil
}

func (m *ConfirmModal) Update(msg tea.Msg) (*ConfirmModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg {
				return ConfirmResultMsg{
					ID:        m.ID,
					Confirmed: false,
					Data:      m.Data,
				}
			}
		case "tab", "shift+tab", "left", "right", "h", "l":
			m.focused = !m.focused
			return m, nil
		case "enter":
			return m, func() tea.Msg {
				return ConfirmResultMsg{
					ID:        m.ID,
					Confirmed: m.focused,
					Data:      m.Data,
				}
			}
		case "y", "Y":
			return m, func() tea.Msg {
				return ConfirmResultMsg{
					ID:        m.ID,
					Confirmed: true,
					Data:      m.Data,
				}
			}
		case "n", "N":
			return m, func() tea.Msg {
				return ConfirmResultMsg{
					ID:        m.ID,
					Confirmed: false,
					Data:      m.Data,
				}
			}
		}
	}
	return m, nil
}

func (m *ConfirmModal) View() string {
	modalWidth := 54
	if m.Width > 0 && m.Width < 60 {
		modalWidth = m.Width - 4
		if modalWidth < 30 {
			modalWidth = 30
		}
	}

	contentWidth := modalWidth - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFOrange).
		MarginBottom(1)

	msgStyle := lipgloss.NewStyle().
		Foreground(theme.ColorForeground).
		Width(contentWidth).
		MarginBottom(2)

	btnStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(2)

	btnConfirmActive := btnStyle.
		Background(theme.ColorStatusDown).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render("[Enter] Confirm")

	btnConfirmInactive := btnStyle.
		Background(theme.ColorSurfaceAlt).
		Foreground(theme.ColorTextDim).
		Render(" Confirm ")

	btnCancelActive := btnStyle.
		Background(theme.ColorCFBlue).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render("[Esc] Cancel")

	btnCancelInactive := btnStyle.
		Background(theme.ColorSurfaceAlt).
		Foreground(theme.ColorTextDim).
		Render(" Cancel ")

	var buttons string
	if m.focused {
		buttons = lipgloss.JoinHorizontal(lipgloss.Center, btnConfirmActive, btnCancelInactive)
	} else {
		buttons = lipgloss.JoinHorizontal(lipgloss.Center, btnConfirmInactive, btnCancelActive)
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(m.Title),
		msgStyle.Render(m.Message),
		"",
		lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Right).Render(buttons),
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorStatusDown).
		Background(theme.ColorSurface).
		Padding(1, 2).
		Width(modalWidth)

	return boxStyle.Render(strings.TrimSpace(body))
}
