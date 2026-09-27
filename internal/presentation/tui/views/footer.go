package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type Footer struct {
	Shortcuts []Shortcut
	Message   string
	IsError   bool
	MsgExpiry time.Time
	Width     int
	Height    int
}

func NewFooter() *Footer {
	return &Footer{
		Shortcuts: []Shortcut{
			{Key: "1-6", Desc: "switch view"},
			{Key: ":", Desc: "commands"},
			{Key: "?", Desc: "help"},
			{Key: "q", Desc: "quit"},
		},
		Width:  80,
		Height: 1,
	}
}

func (f *Footer) SetDimensions(width, height int) {
	f.Width = width
	f.Height = height
}

func (f *Footer) SetShortcuts(sc []Shortcut) {
	f.Shortcuts = sc
}

func (f *Footer) SetStatus(msg string, isError bool, ttl time.Duration) {
	f.Message = msg
	f.IsError = isError
	if ttl > 0 {
		f.MsgExpiry = time.Now().Add(ttl)
	} else {
		f.MsgExpiry = time.Time{}
	}
}

func (f *Footer) ClearStatus() {
	f.Message = ""
	f.IsError = false
	f.MsgExpiry = time.Time{}
}

func (f *Footer) View() string {
	th := theme.Current()

	hasMessage := f.Message != ""
	if hasMessage && !f.MsgExpiry.IsZero() && time.Now().After(f.MsgExpiry) {
		f.Message = ""
		hasMessage = false
	}

	var leftPart string
	if hasMessage {
		if f.IsError {
			leftPart = th.ErrorText.Render("✗ " + f.Message)
		} else {
			leftPart = th.SuccessText.Render("✓ " + f.Message)
		}
	} else {

		var scParts []string
		for _, sc := range f.Shortcuts {
			k := th.KeyBadge.Render(sc.Key)
			d := th.Muted.Render(sc.Desc)
			scParts = append(scParts, fmt.Sprintf("%s %s", k, d))
		}
		leftPart = strings.Join(scParts, "  ")
	}

	cmdHint := fmt.Sprintf("%s %s  %s %s",
		th.KeyBadge.Render(":"), th.Muted.Render("commands"),
		th.KeyBadge.Render("?"), th.Muted.Render("help"),
	)

	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(cmdHint)
	spacing := f.Width - leftWidth - rightWidth - 2

	var content string
	if spacing > 0 {
		content = leftPart + strings.Repeat(" ", spacing) + cmdHint
	} else {
		content = leftPart
	}

	footerStyle := lipgloss.NewStyle().
		Width(f.Width).
		Height(f.Height).
		Background(theme.ColorSurface).
		Foreground(theme.ColorForeground).
		Padding(0, 1)

	return footerStyle.Render(content)
}
