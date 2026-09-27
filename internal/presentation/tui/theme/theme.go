package theme

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	ColorForeground   = lipgloss.Color("#E0E0E0")
	ColorBackground   = lipgloss.Color("#121212")
	ColorSurface      = lipgloss.Color("#1E1E1E")
	ColorSurfaceAlt   = lipgloss.Color("#252525")
	ColorBorder       = lipgloss.Color("#333333")
	ColorBorderActive = lipgloss.Color("#444444")
	ColorBorderFocus  = lipgloss.Color("#00A3E0")

	ColorCFBlue   = lipgloss.Color("#00A3E0")
	ColorCFOrange = lipgloss.Color("#F38020")
	ColorCFAmber  = lipgloss.Color("#FF8800")

	ColorStatusHealthy  = lipgloss.Color("#00D26A")
	ColorStatusDegraded = lipgloss.Color("#FFB020")
	ColorStatusDown     = lipgloss.Color("#F83B3B")
	ColorStatusMuted    = lipgloss.Color("#888888")

	ColorText       = lipgloss.Color("#E0E0E0")
	ColorTextDim    = lipgloss.Color("#888888")
	ColorTextMuted  = lipgloss.Color("#6E6E6E")
	ColorTextSubtle = lipgloss.Color("#555555")
	ColorSelection  = lipgloss.Color("#2A2A2A")
)

const (
	SymbolHealthy  = "●"
	SymbolDegraded = "◐"
	SymbolDown     = "○"
	SymbolError    = "!"
	SymbolAuto     = "ℹ"
)

type Theme struct {
	Base      lipgloss.Style
	App       lipgloss.Style
	Surface   lipgloss.Style
	Pane      lipgloss.Style
	PaneFocus lipgloss.Style

	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Header      lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Highlight   lipgloss.Style
	Accent      lipgloss.Style
	ErrorText   lipgloss.Style
	SuccessText lipgloss.Style
	WarningText lipgloss.Style

	KeyBadge     lipgloss.Style
	ActionBadge  lipgloss.Style
	TagBadge     lipgloss.Style
	BadgeHealthy lipgloss.Style
	BadgeWarning lipgloss.Style
	BadgeError   lipgloss.Style
	BadgeMuted   lipgloss.Style

	TableHeader  lipgloss.Style
	SelectedRow  lipgloss.Style
	ActiveItem   lipgloss.Style
	InactiveItem lipgloss.Style

	StatusBar   lipgloss.Style
	StatusKey   lipgloss.Style
	StatusValue lipgloss.Style
	Breadcrumb  lipgloss.Style
}

func DefaultTheme() *Theme {
	t := &Theme{}

	t.Base = lipgloss.NewStyle().
		Foreground(ColorForeground).
		Background(ColorBackground)

	t.App = lipgloss.NewStyle().
		Background(ColorBackground).
		Foreground(ColorForeground)

	t.Surface = lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorForeground)

	t.Pane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Background(ColorBackground)

	t.PaneFocus = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCFBlue).
		Background(ColorBackground)

	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorForeground)

	t.Subtitle = lipgloss.NewStyle().
		Bold(false).
		Foreground(ColorCFBlue)

	t.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCFOrange)

	t.Body = lipgloss.NewStyle().
		Foreground(ColorForeground)

	t.Muted = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	t.Highlight = lipgloss.NewStyle().
		Foreground(ColorCFOrange).
		Bold(true)

	t.Accent = lipgloss.NewStyle().
		Foreground(ColorCFBlue).
		Bold(true)

	t.ErrorText = lipgloss.NewStyle().
		Foreground(ColorStatusDown).
		Bold(true)

	t.SuccessText = lipgloss.NewStyle().
		Foreground(ColorStatusHealthy).
		Bold(true)

	t.WarningText = lipgloss.NewStyle().
		Foreground(ColorStatusDegraded).
		Bold(true)

	t.KeyBadge = lipgloss.NewStyle().
		Foreground(ColorCFBlue).
		Bold(true)

	t.ActionBadge = lipgloss.NewStyle().
		Background(ColorSurfaceAlt).
		Foreground(ColorForeground).
		Padding(0, 1)

	t.TagBadge = lipgloss.NewStyle().
		Background(ColorBorder).
		Foreground(ColorForeground).
		Padding(0, 1)

	t.BadgeHealthy = lipgloss.NewStyle().
		Foreground(ColorStatusHealthy).
		Bold(true)

	t.BadgeWarning = lipgloss.NewStyle().
		Foreground(ColorStatusDegraded).
		Bold(true)

	t.BadgeError = lipgloss.NewStyle().
		Foreground(ColorStatusDown).
		Bold(true)

	t.BadgeMuted = lipgloss.NewStyle().
		Foreground(ColorStatusMuted)

	t.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorCFBlue).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ColorBorder)

	t.SelectedRow = lipgloss.NewStyle().
		Background(ColorSelection).
		Foreground(ColorForeground).
		Bold(true)

	t.ActiveItem = lipgloss.NewStyle().
		Foreground(ColorCFOrange).
		Bold(true)

	t.InactiveItem = lipgloss.NewStyle().
		Foreground(ColorTextDim)

	t.StatusBar = lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorTextDim).
		Padding(0, 1)

	t.StatusKey = lipgloss.NewStyle().
		Foreground(ColorCFBlue).
		Bold(true)

	t.StatusValue = lipgloss.NewStyle().
		Foreground(ColorForeground)

	t.Breadcrumb = lipgloss.NewStyle().
		Foreground(ColorCFOrange).
		Bold(true)

	return t
}

var currentTheme = DefaultTheme()

func Current() *Theme {
	return currentTheme
}

func BadgeHealthy(label string) string {
	if label == "" {
		label = "Healthy"
	}
	return currentTheme.BadgeHealthy.Render(SymbolHealthy + " " + label)
}

func BadgeProxied() string {
	return currentTheme.BadgeHealthy.Render(SymbolHealthy + " Proxied")
}

func BadgeActive() string {
	return currentTheme.BadgeHealthy.Render(SymbolHealthy + " Active")
}

func BadgeDegraded(label string) string {
	if label == "" {
		label = "Degraded"
	}
	return currentTheme.BadgeWarning.Render(SymbolDegraded + " " + label)
}

func BadgeDNSOnly() string {
	return currentTheme.BadgeWarning.Render(SymbolDegraded + " DNS Only")
}

func BadgePending() string {
	return currentTheme.BadgeWarning.Render(SymbolDegraded + " Pending")
}

func BadgeDown(label string) string {
	if label == "" {
		label = "Down"
	}
	return currentTheme.BadgeError.Render(SymbolDown + " " + label)
}

func BadgeInactive() string {
	return currentTheme.BadgeError.Render(SymbolDown + " Inactive")
}

func BadgeError(label string) string {
	if label == "" {
		label = "Error"
	}
	return currentTheme.BadgeError.Render(SymbolError + " " + label)
}

func BadgeAuto(label string) string {
	if label == "" {
		label = "Auto"
	}
	return currentTheme.BadgeMuted.Render(SymbolAuto + " " + label)
}

func FormatKeyBadge(key string) string {
	return currentTheme.KeyBadge.Render("[" + key + "]")
}

func RenderTitle(s string) string {
	return currentTheme.Title.Render(s)
}

func RenderSubtitle(s string) string {
	return currentTheme.Subtitle.Render(s)
}

func RenderMuted(s string) string {
	return currentTheme.Muted.Render(s)
}

func RenderHighlight(s string) string {
	return currentTheme.Highlight.Render(s)
}

func RenderError(s string) string {
	return currentTheme.ErrorText.Render(s)
}
