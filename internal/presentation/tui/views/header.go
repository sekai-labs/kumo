package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type Header struct {
	Profile     string
	AccountName string
	AccountID   string
	ZoneName    string
	TokenValid  bool
	TokenStatus string
	Breadcrumb  string
	Loading     bool
	Spinner     spinner.Model
	Width       int
	Height      int
}

func NewHeader() *Header {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorCFOrange)

	return &Header{
		Profile:     "default",
		AccountName: "Cloudflare",
		AccountID:   "",
		ZoneName:    "—",
		TokenValid:  true,
		TokenStatus: "Active",
		Breadcrumb:  "Overview",
		Loading:     false,
		Spinner:     s,
		Width:       80,
		Height:      2,
	}
}

func (h *Header) SetDimensions(width, height int) {
	h.Width = width
	h.Height = height
}

func (h *Header) SetLoading(loading bool) {
	h.Loading = loading
}

func (h *Header) SetBreadcrumb(bc string) {
	h.Breadcrumb = bc
}

func (h *Header) SetProfile(p string) {
	if p != "" {
		h.Profile = p
	}
}

func (h *Header) SetAccount(name, id string) {
	if name != "" {
		h.AccountName = name
	}
	h.AccountID = id
}

func (h *Header) SetZone(name string) {
	if name != "" {
		h.ZoneName = name
	} else {
		h.ZoneName = "—"
	}
}

func (h *Header) SetTokenStatus(valid bool, desc string) {
	h.TokenValid = valid
	h.TokenStatus = desc
}

func (h *Header) Init() tea.Cmd {
	return h.Spinner.Tick
}

func (h *Header) Update(msg tea.Msg) (*Header, tea.Cmd) {
	var cmd tea.Cmd
	if h.Loading {
		h.Spinner, cmd = h.Spinner.Update(msg)
	}
	return h, cmd
}

func (h *Header) View() string {
	th := theme.Current()

	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFOrange).
		Render("☁  KUMO")

	separator := th.Muted.Render(" › ")
	bc := th.Breadcrumb.Render(h.Breadcrumb)

	var spinStr string
	if h.Loading {
		spinStr = " " + h.Spinner.View()
	}

	left := fmt.Sprintf("%s%s%s%s", logo, separator, bc, spinStr)

	profileBadge := th.TagBadge.Render("profile:" + h.Profile)

	accStr := h.AccountName
	if h.AccountID != "" {
		accStr = fmt.Sprintf("%s (%s)", h.AccountName, truncate(h.AccountID, 8))
	}
	accountBadge := th.TagBadge.Render("acc:" + accStr)

	zoneBadge := th.TagBadge.Render("zone:" + h.ZoneName)

	var tokenBadge string
	if h.TokenValid {
		tokenBadge = th.BadgeHealthy.Render(theme.SymbolHealthy + " " + h.TokenStatus)
	} else {
		tokenBadge = th.BadgeError.Render(theme.SymbolError + " " + h.TokenStatus)
	}

	right := fmt.Sprintf("%s  %s  %s  %s", profileBadge, accountBadge, zoneBadge, tokenBadge)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	space := h.Width - leftWidth - rightWidth - 2

	var line string
	if space > 0 {
		line = left + strings.Repeat(" ", space) + right
	} else {

		line = left
	}

	headerStyle := lipgloss.NewStyle().
		Width(h.Width).
		Height(h.Height).
		Background(theme.ColorSurface).
		Foreground(theme.ColorForeground).
		Padding(0, 1)

	return headerStyle.Render(line)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
