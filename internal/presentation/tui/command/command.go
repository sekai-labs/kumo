package command

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Category string

const (
	CategoryNavigation Category = "Navigation"
	CategoryDNS        Category = "DNS"
	CategoryTunnels    Category = "Tunnels"
	CategoryGeneral    Category = "General"
	CategoryActions    Category = "Actions"
)

type Command struct {
	ID          string
	Title       string
	Category    Category
	Shortcut    string
	Description string
	Action      func() tea.Msg
	KeyBinding  key.Binding
}

type Registry struct {
	commands []Command
}

func NewRegistry() *Registry {
	r := &Registry{
		commands: make([]Command, 0),
	}
	r.registerDefaults()
	return r
}

func (r *Registry) Register(cmd Command) {

	for i, existing := range r.commands {
		if existing.ID == cmd.ID {
			r.commands[i] = cmd
			return
		}
	}
	r.commands = append(r.commands, cmd)
}

func (r *Registry) All() []Command {
	out := make([]Command, len(r.commands))
	copy(out, r.commands)
	return out
}

func (r *Registry) FindByID(id string) (Command, bool) {
	for _, cmd := range r.commands {
		if cmd.ID == id {
			return cmd, true
		}
	}
	return Command{}, false
}

func (r *Registry) Filter(query string) []Command {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return r.All()
	}

	var matched []Command
	for _, cmd := range r.commands {
		title := strings.ToLower(cmd.Title)
		desc := strings.ToLower(cmd.Description)
		shortcut := strings.ToLower(cmd.Shortcut)
		id := strings.ToLower(cmd.ID)
		cat := strings.ToLower(string(cmd.Category))

		if strings.Contains(title, q) ||
			strings.Contains(desc, q) ||
			strings.Contains(shortcut, q) ||
			strings.Contains(id, q) ||
			strings.Contains(cat, q) {
			matched = append(matched, cmd)
		}
	}
	return matched
}

func (r *Registry) MatchKey(msg tea.KeyMsg) (Command, bool) {
	for _, cmd := range r.commands {
		if cmd.KeyBinding.Keys() != nil && key.Matches(msg, cmd.KeyBinding) {
			return cmd, true
		}
	}
	return Command{}, false
}

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Tab        key.Binding
	Backtab    key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Filter     key.Binding
	Command    key.Binding
	Refresh    key.Binding
	ToggleProx key.Binding
	New        key.Binding
	Edit       key.Binding
	Delete     key.Binding
	Help       key.Binding
	Quit       key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/→", "right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "next pane"),
		),
		Backtab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("Shift+Tab", "previous pane"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select/confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "back/cancel"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter/search"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command palette"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleProx: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle proxy"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new record"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit record"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete record"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

type SwitchViewMsg struct {
	View string
}

type RefreshMsg struct{}

type ToggleHelpMsg struct{}

type QuitMsg struct{}

type OpenCommandPaletteMsg struct{}

type FilterMsg struct{}

type ToggleProxyMsg struct{}

type NewRecordMsg struct{}

type EditRecordMsg struct{}

type DeleteRecordMsg struct{}

func (r *Registry) registerDefaults() {
	km := DefaultKeyMap()

	r.Register(Command{
		ID:          "app.quit",
		Title:       "Quit",
		Category:    CategoryGeneral,
		Shortcut:    "q",
		Description: "Exit Kumo terminal",
		KeyBinding:  km.Quit,
		Action: func() tea.Msg {
			return tea.Quit()
		},
	})

	r.Register(Command{
		ID:          "app.help",
		Title:       "Help",
		Category:    CategoryGeneral,
		Shortcut:    "?",
		Description: "Toggle help view",
		KeyBinding:  km.Help,
		Action: func() tea.Msg {
			return ToggleHelpMsg{}
		},
	})

	r.Register(Command{
		ID:          "app.palette",
		Title:       "Command Palette",
		Category:    CategoryGeneral,
		Shortcut:    ":",
		Description: "Open command palette",
		KeyBinding:  km.Command,
		Action: func() tea.Msg {
			return OpenCommandPaletteMsg{}
		},
	})

	r.Register(Command{
		ID:          "app.refresh",
		Title:       "Refresh",
		Category:    CategoryGeneral,
		Shortcut:    "r",
		Description: "Force refresh current view and bypass cache",
		KeyBinding:  km.Refresh,
		Action: func() tea.Msg {
			return RefreshMsg{}
		},
	})

	r.Register(Command{
		ID:          "app.filter",
		Title:       "Filter",
		Category:    CategoryNavigation,
		Shortcut:    "/",
		Description: "Filter list items",
		KeyBinding:  km.Filter,
		Action: func() tea.Msg {
			return FilterMsg{}
		},
	})

	r.Register(Command{
		ID:          "dns.toggle_proxy",
		Title:       "Toggle DNS Proxy",
		Category:    CategoryDNS,
		Shortcut:    "p",
		Description: "Toggle Cloudflare proxy on selected DNS record",
		KeyBinding:  km.ToggleProx,
		Action: func() tea.Msg {
			return ToggleProxyMsg{}
		},
	})

	r.Register(Command{
		ID:          "dns.new",
		Title:       "New DNS Record",
		Category:    CategoryDNS,
		Shortcut:    "n",
		Description: "Add a new DNS record to the active zone",
		KeyBinding:  km.New,
		Action: func() tea.Msg {
			return NewRecordMsg{}
		},
	})

	r.Register(Command{
		ID:          "dns.edit",
		Title:       "Edit DNS Record",
		Category:    CategoryDNS,
		Shortcut:    "e",
		Description: "Edit selected DNS record",
		KeyBinding:  km.Edit,
		Action: func() tea.Msg {
			return EditRecordMsg{}
		},
	})

	r.Register(Command{
		ID:          "dns.delete",
		Title:       "Delete DNS Record",
		Category:    CategoryDNS,
		Shortcut:    "d",
		Description: "Delete selected DNS record",
		KeyBinding:  km.Delete,
		Action: func() tea.Msg {
			return DeleteRecordMsg{}
		},
	})

	r.Register(Command{
		ID:          "view.overview",
		Title:       "Go to Overview",
		Category:    CategoryNavigation,
		Shortcut:    "1",
		Description: "Switch to system overview dashboard",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "overview"}
		},
	})

	r.Register(Command{
		ID:          "view.dns",
		Title:       "Go to DNS Records",
		Category:    CategoryNavigation,
		Shortcut:    "2",
		Description: "Switch to DNS management view",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "dns"}
		},
	})

	r.Register(Command{
		ID:          "view.tunnels",
		Title:       "Go to Cloudflare Tunnels",
		Category:    CategoryNavigation,
		Shortcut:    "3",
		Description: "Switch to Cloudflare Tunnels monitor",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "tunnels"}
		},
	})

	r.Register(Command{
		ID:          "view.workers",
		Title:       "Go to Workers & Pages",
		Category:    CategoryNavigation,
		Shortcut:    "4",
		Description: "Switch to Workers and Pages view",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "workers"}
		},
	})

	r.Register(Command{
		ID:          "view.rulesets",
		Title:       "Go to Rulesets",
		Category:    CategoryNavigation,
		Shortcut:    "5",
		Description: "Switch to Rulesets view",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "rulesets"}
		},
	})

	r.Register(Command{
		ID:          "view.analytics",
		Title:       "Go to Analytics",
		Category:    CategoryNavigation,
		Shortcut:    "6",
		Description: "Switch to Traffic Analytics view",
		Action: func() tea.Msg {
			return SwitchViewMsg{View: "analytics"}
		},
	})
}
