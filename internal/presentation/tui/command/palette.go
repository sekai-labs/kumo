package command

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type PaletteCloseMsg struct{}

type Palette struct {
	registry *Registry
	input    textinput.Model
	filtered []Command
	cursor   int
	Width    int
	Height   int
}

func NewPalette(registry *Registry) *Palette {
	ti := textinput.New()
	ti.Prompt = ": "
	ti.Placeholder = "Type a command or filter..."
	ti.Focus()

	p := &Palette{
		registry: registry,
		input:    ti,
		cursor:   0,
		Width:    64,
		Height:   16,
	}
	p.refilter()
	return p
}

func (p *Palette) SetSize(width, height int) {
	p.Width = width
	p.Height = height
}

func (p *Palette) Init() tea.Cmd {
	return textinput.Blink
}

func (p *Palette) refilter() {
	query := p.input.Value()
	p.filtered = p.registry.Filter(query)
	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *Palette) Reset() {
	p.input.SetValue("")
	p.input.Focus()
	p.cursor = 0
	p.refilter()
}

func (p *Palette) Update(msg tea.Msg) (*Palette, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return p, func() tea.Msg {
				return PaletteCloseMsg{}
			}
		case "up", "ctrl+k":
			if p.cursor > 0 {
				p.cursor--
			}
			return p, nil
		case "down", "ctrl+j":
			if p.cursor < len(p.filtered)-1 {
				p.cursor++
			}
			return p, nil
		case "enter":
			if len(p.filtered) > 0 && p.cursor < len(p.filtered) {
				selected := p.filtered[p.cursor]
				return p, tea.Batch(
					func() tea.Msg {
						return PaletteCloseMsg{}
					},
					func() tea.Msg {
						if selected.Action != nil {
							return selected.Action()
						}
						return nil
					},
				)
			}
			return p, func() tea.Msg {
				return PaletteCloseMsg{}
			}
		}
	}

	var cmd tea.Cmd
	oldVal := p.input.Value()
	p.input, cmd = p.input.Update(msg)
	if p.input.Value() != oldVal {
		p.refilter()
	}
	return p, cmd
}

func (p *Palette) View() string {
	modalWidth := 64
	if p.Width > 0 && p.Width < 70 {
		modalWidth = p.Width - 4
		if modalWidth < 36 {
			modalWidth = 36
		}
	}
	contentWidth := modalWidth - 4

	inputBar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.ColorBorder).
		Width(contentWidth).
		Render(p.input.View())

	var items []string
	const maxVisible = 7

	if len(p.filtered) == 0 {
		items = append(items, theme.RenderMuted("  No matching commands"))
	} else {
		startIdx := 0
		if p.cursor >= maxVisible {
			startIdx = p.cursor - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(p.filtered) {
			endIdx = len(p.filtered)
		}

		for i := startIdx; i < endIdx; i++ {
			cmd := p.filtered[i]
			isCursor := i == p.cursor

			title := cmd.Title
			cat := string(cmd.Category)
			key := cmd.Shortcut

			titleStyled := lipgloss.NewStyle().Width(contentWidth - 28).Render(title)
			catStyled := lipgloss.NewStyle().Width(12).Foreground(theme.ColorCFBlue).Render(cat)
			keyStyled := lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Foreground(theme.ColorCFOrange).Bold(true).Render(key)

			var row string
			if isCursor {
				row = lipgloss.NewStyle().
					Background(theme.ColorSelection).
					Bold(true).
					Foreground(lipgloss.Color("#FFFFFF")).
					Width(contentWidth).
					Render(" > " + titleStyled + " " + catStyled + " " + keyStyled)
			} else {
				row = lipgloss.NewStyle().
					Foreground(theme.ColorForeground).
					Width(contentWidth).
					Render("   " + titleStyled + " " + catStyled + " " + keyStyled)
			}
			items = append(items, row)
		}
	}

	helpLine := theme.RenderMuted("[↑/↓] Navigate  [Enter] Execute  [Esc] Close")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCFOrange).Render("Command Palette"),
		inputBar,
		"",
		lipgloss.JoinVertical(lipgloss.Left, items...),
		"",
		helpLine,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCFOrange).
		Background(theme.ColorSurface).
		Padding(1, 2).
		Width(modalWidth)

	return boxStyle.Render(body)
}
