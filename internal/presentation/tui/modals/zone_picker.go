package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type ZoneSelectedMsg struct {
	Zone    zone.Zone
	Account account.Account
}

type AccountSelectedMsg struct {
	Account account.Account
}

type ZonePickerCloseMsg struct{}

type PickerTab int

const (
	TabZones PickerTab = iota
	TabAccounts
)

type ZonePickerModal struct {
	ActiveTab PickerTab
	Accounts  []account.Account
	Zones     []zone.Zone
	ActiveAcc account.Account
	ActiveZn  zone.Zone

	filterInput textinput.Model
	zoneCursor  int
	accCursor   int

	filteredZones []zone.Zone
	filteredAccs  []account.Account

	Width  int
	Height int
}

func NewZonePickerModal(accounts []account.Account, zones []zone.Zone, activeAcc account.Account, activeZn zone.Zone) *ZonePickerModal {
	ti := textinput.New()
	ti.Placeholder = "Type to filter zones or accounts..."
	ti.Prompt = " / "
	ti.Focus()

	m := &ZonePickerModal{
		ActiveTab:   TabZones,
		Accounts:    accounts,
		Zones:       zones,
		ActiveAcc:   activeAcc,
		ActiveZn:    activeZn,
		filterInput: ti,
		Width:       68,
		Height:      20,
	}

	m.refilter()
	return m
}

func (m *ZonePickerModal) SetData(accounts []account.Account, zones []zone.Zone, activeAcc account.Account, activeZn zone.Zone) {
	m.Accounts = accounts
	m.Zones = zones
	m.ActiveAcc = activeAcc
	m.ActiveZn = activeZn
	m.refilter()
}

func (m *ZonePickerModal) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *ZonePickerModal) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ZonePickerModal) refilter() {
	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))

	m.filteredZones = nil
	for _, z := range m.Zones {
		if query == "" || strings.Contains(strings.ToLower(z.Name), query) || strings.Contains(strings.ToLower(string(z.ID)), query) {
			m.filteredZones = append(m.filteredZones, z)
		}
	}
	if m.zoneCursor >= len(m.filteredZones) {
		m.zoneCursor = len(m.filteredZones) - 1
	}
	if m.zoneCursor < 0 {
		m.zoneCursor = 0
	}

	m.filteredAccs = nil
	for _, a := range m.Accounts {
		if query == "" || strings.Contains(strings.ToLower(a.Name), query) || strings.Contains(strings.ToLower(string(a.ID)), query) {
			m.filteredAccs = append(m.filteredAccs, a)
		}
	}
	if m.accCursor >= len(m.filteredAccs) {
		m.accCursor = len(m.filteredAccs) - 1
	}
	if m.accCursor < 0 {
		m.accCursor = 0
	}
}

func (m *ZonePickerModal) Update(msg tea.Msg) (*ZonePickerModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return ZonePickerCloseMsg{}
			}
		case "tab", "right":
			if msg.String() == "tab" || (msg.String() == "right" && m.filterInput.Value() == "") {
				if m.ActiveTab == TabZones {
					m.ActiveTab = TabAccounts
				} else {
					m.ActiveTab = TabZones
				}
				return m, nil
			}
		case "shift+tab", "left":
			if msg.String() == "shift+tab" || (msg.String() == "left" && m.filterInput.Value() == "") {
				if m.ActiveTab == TabZones {
					m.ActiveTab = TabAccounts
				} else {
					m.ActiveTab = TabZones
				}
				return m, nil
			}
		case "up", "ctrl+k":
			if m.ActiveTab == TabZones {
				if m.zoneCursor > 0 {
					m.zoneCursor--
				}
			} else {
				if m.accCursor > 0 {
					m.accCursor--
				}
			}
			return m, nil
		case "down", "ctrl+j":
			if m.ActiveTab == TabZones {
				if m.zoneCursor < len(m.filteredZones)-1 {
					m.zoneCursor++
				}
			} else {
				if m.accCursor < len(m.filteredAccs)-1 {
					m.accCursor++
				}
			}
			return m, nil
		case "enter":
			if m.ActiveTab == TabZones && len(m.filteredZones) > 0 && m.zoneCursor < len(m.filteredZones) {
				selectedZone := m.filteredZones[m.zoneCursor]

				var acc account.Account
				for _, a := range m.Accounts {
					if a.ID == selectedZone.AccountID {
						acc = a
						break
					}
				}
				return m, func() tea.Msg {
					return ZoneSelectedMsg{
						Zone:    selectedZone,
						Account: acc,
					}
				}
			} else if m.ActiveTab == TabAccounts && len(m.filteredAccs) > 0 && m.accCursor < len(m.filteredAccs) {
				selectedAcc := m.filteredAccs[m.accCursor]
				return m, func() tea.Msg {
					return AccountSelectedMsg{
						Account: selectedAcc,
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	oldVal := m.filterInput.Value()
	m.filterInput, cmd = m.filterInput.Update(msg)
	if m.filterInput.Value() != oldVal {
		m.refilter()
	}
	return m, cmd
}

func (m *ZonePickerModal) View() string {
	modalWidth := 68
	if m.Width > 0 && m.Width < 74 {
		modalWidth = m.Width - 4
		if modalWidth < 40 {
			modalWidth = 40
		}
	}
	contentWidth := modalWidth - 4

	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(theme.ColorCFBlue).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(theme.ColorTextDim).
		Background(theme.ColorSurfaceAlt).
		Padding(0, 2)

	var tabsHeader string
	if m.ActiveTab == TabZones {
		tabsHeader = lipgloss.JoinHorizontal(
			lipgloss.Center,
			activeTabStyle.Render(fmt.Sprintf("Zones (%d)", len(m.filteredZones))),
			inactiveTabStyle.Render(fmt.Sprintf("Accounts (%d)", len(m.filteredAccs))),
		)
	} else {
		tabsHeader = lipgloss.JoinHorizontal(
			lipgloss.Center,
			inactiveTabStyle.Render(fmt.Sprintf("Zones (%d)", len(m.filteredZones))),
			activeTabStyle.Render(fmt.Sprintf("Accounts (%d)", len(m.filteredAccs))),
		)
	}

	filterBar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.ColorBorder).
		Width(contentWidth).
		Render(m.filterInput.View())

	var listItems []string
	const maxVisible = 8

	if m.ActiveTab == TabZones {
		if len(m.filteredZones) == 0 {
			listItems = append(listItems, theme.RenderMuted("  No matching zones found"))
		} else {
			startIdx := 0
			if m.zoneCursor >= maxVisible {
				startIdx = m.zoneCursor - maxVisible + 1
			}
			endIdx := startIdx + maxVisible
			if endIdx > len(m.filteredZones) {
				endIdx = len(m.filteredZones)
			}

			for i := startIdx; i < endIdx; i++ {
				z := m.filteredZones[i]
				isActive := z.ID == m.ActiveZn.ID
				isCursor := i == m.zoneCursor

				statusBadge := theme.BadgeActive()
				if z.Paused {
					statusBadge = theme.BadgeDegraded("Paused")
				}

				name := z.Name
				if isActive {
					name = name + " [active]"
				}

				var row string
				nameStyled := lipgloss.NewStyle().Width(contentWidth - 24).Render(name)
				planStyled := lipgloss.NewStyle().Width(10).Foreground(theme.ColorCFOrange).Render(z.Plan)

				if isCursor {
					row = lipgloss.NewStyle().
						Background(theme.ColorSelection).
						Bold(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Width(contentWidth).
						Render(fmt.Sprintf(" > %s %s %s", nameStyled, planStyled, statusBadge))
				} else {
					row = lipgloss.NewStyle().
						Foreground(theme.ColorForeground).
						Width(contentWidth).
						Render(fmt.Sprintf("   %s %s %s", nameStyled, planStyled, statusBadge))
				}
				listItems = append(listItems, row)
			}
		}
	} else {
		if len(m.filteredAccs) == 0 {
			listItems = append(listItems, theme.RenderMuted("  No matching accounts found"))
		} else {
			startIdx := 0
			if m.accCursor >= maxVisible {
				startIdx = m.accCursor - maxVisible + 1
			}
			endIdx := startIdx + maxVisible
			if endIdx > len(m.filteredAccs) {
				endIdx = len(m.filteredAccs)
			}

			for i := startIdx; i < endIdx; i++ {
				a := m.filteredAccs[i]
				isActive := a.ID == m.ActiveAcc.ID
				isCursor := i == m.accCursor

				name := a.Name
				if isActive {
					name = name + " [active]"
				}

				nameStyled := lipgloss.NewStyle().Width(contentWidth - 22).Render(name)
				idStyled := lipgloss.NewStyle().Foreground(theme.ColorTextDim).Render(string(a.ID))

				var row string
				if isCursor {
					row = lipgloss.NewStyle().
						Background(theme.ColorSelection).
						Bold(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Width(contentWidth).
						Render(fmt.Sprintf(" > %s %s", nameStyled, idStyled))
				} else {
					row = lipgloss.NewStyle().
						Foreground(theme.ColorForeground).
						Width(contentWidth).
						Render(fmt.Sprintf("   %s %s", nameStyled, idStyled))
				}
				listItems = append(listItems, row)
			}
		}
	}

	helpLine := theme.RenderMuted("[↑/↓] Navigate  [Tab] Switch Tab  [Enter] Select  [Esc] Cancel")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		tabsHeader,
		filterBar,
		"",
		lipgloss.JoinVertical(lipgloss.Left, listItems...),
		"",
		helpLine,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorCFBlue).
		Background(theme.ColorSurface).
		Padding(1, 2).
		Width(modalWidth)

	return boxStyle.Render(body)
}
