package command_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/presentation/tui/command"
)

func TestRegistryDefaults(t *testing.T) {
	reg := command.NewRegistry()
	cmds := reg.All()

	if len(cmds) == 0 {
		t.Fatal("expected registered default commands")
	}

	cmd, found := reg.FindByID("app.quit")
	if !found {
		t.Fatalf("expected app.quit command to be registered")
	}

	if cmd.Shortcut != "q" {
		t.Errorf("expected shortcut 'q', got %s", cmd.Shortcut)
	}

	msg := cmd.Action()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg from quit action, got %T", msg)
	}
}

func TestRegistryFilter(t *testing.T) {
	reg := command.NewRegistry()

	all := reg.Filter("")
	if len(all) != len(reg.All()) {
		t.Errorf("empty filter should return all commands, got %d vs %d", len(all), len(reg.All()))
	}

	dnsMatches := reg.Filter("DNS")
	if len(dnsMatches) == 0 {
		t.Errorf("expected matches for 'DNS'")
	}
	for _, m := range dnsMatches {
		t.Logf("DNS match: %s - %s", m.ID, m.Title)
	}

	pMatches := reg.Filter("p")
	if len(pMatches) == 0 {
		t.Errorf("expected matches for shortcut 'p'")
	}

	none := reg.Filter("nonexistent_random_command_123")
	if len(none) != 0 {
		t.Errorf("expected 0 matches, got %d", len(none))
	}
}

func TestRegistryMatchKey(t *testing.T) {
	reg := command.NewRegistry()

	keyQ := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	cmd, matched := reg.MatchKey(keyQ)
	if !matched {
		t.Fatalf("expected key 'q' to match a command")
	}
	if cmd.ID != "app.quit" {
		t.Errorf("expected 'app.quit', got %s", cmd.ID)
	}

	keyHelp := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	cmd, matched = reg.MatchKey(keyHelp)
	if !matched {
		t.Fatalf("expected key '?' to match help command")
	}
	if cmd.ID != "app.help" {
		t.Errorf("expected 'app.help', got %s", cmd.ID)
	}

	keyP := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	cmd, matched = reg.MatchKey(keyP)
	if !matched {
		t.Fatalf("expected key 'p' to match toggle proxy")
	}
	if cmd.ID != "dns.toggle_proxy" {
		t.Errorf("expected 'dns.toggle_proxy', got %s", cmd.ID)
	}
}

func TestRegisterCustomCommand(t *testing.T) {
	reg := command.NewRegistry()
	custom := command.Command{
		ID:          "custom.test",
		Title:       "Custom Test",
		Category:    command.CategoryActions,
		Shortcut:    "x",
		Description: "A custom test command",
	}

	reg.Register(custom)

	found, ok := reg.FindByID("custom.test")
	if !ok {
		t.Fatalf("expected custom command to be found")
	}
	if found.Title != "Custom Test" {
		t.Errorf("unexpected title %s", found.Title)
	}

	custom.Title = "Updated Custom Test"
	reg.Register(custom)
	found, _ = reg.FindByID("custom.test")
	if found.Title != "Updated Custom Test" {
		t.Errorf("expected updated title, got %s", found.Title)
	}
}
