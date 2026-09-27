package modals_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/modals"
)

func TestConfirmModal(t *testing.T) {
	m := modals.NewConfirmModal("del-1", "Delete Record", "Are you sure?", "extra-data")
	m.SetSize(80, 24)

	rendered := m.View()
	if rendered == "" {
		t.Fatal("expected non-empty render from confirm modal")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}
	msg := cmd()
	res, ok := msg.(modals.ConfirmResultMsg)
	if !ok {
		t.Fatalf("expected ConfirmResultMsg, got %T", msg)
	}
	if res.Confirmed {
		t.Errorf("expected confirmed=false on Esc")
	}
	if res.ID != "del-1" {
		t.Errorf("expected ID 'del-1', got %s", res.ID)
	}

	m2 := modals.NewConfirmModal("del-2", "Delete", "Confirm?", nil)
	_, cmd2 := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd2 == nil {
		t.Fatal("expected command on Enter")
	}
	msg2 := cmd2()
	res2, ok := msg2.(modals.ConfirmResultMsg)
	if !ok || !res2.Confirmed {
		t.Errorf("expected confirmed=true on Enter, got %+v", res2)
	}

	m3 := modals.NewConfirmModal("del-3", "Delete", "Confirm?", nil)
	m3, _ = m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	_, cmd3 := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd3 == nil {
		t.Fatal("expected command on Enter")
	}
	msg3 := cmd3()
	res3 := msg3.(modals.ConfirmResultMsg)
	if res3.Confirmed {
		t.Errorf("expected confirmed=false after tab to cancel")
	}
}

func TestDNSFormModal(t *testing.T) {
	zID := zone.ZoneID("zone-123")
	createForm := modals.NewDNSCreateForm(zID)
	createForm.SetSize(100, 30)

	rendered := createForm.View()
	if rendered == "" {
		t.Fatal("expected non-empty render from create form")
	}

	_, cancelCmd := createForm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cancelCmd == nil {
		t.Fatal("expected cancel cmd on Esc")
	}
	if _, ok := cancelCmd().(modals.DNSFormCancelMsg); !ok {
		t.Errorf("expected DNSFormCancelMsg")
	}

	rec, err := dns.NewDNSRecord("rec-1", zID, dns.TypeA, "api.example.com", "1.2.3.4", true, 300, "note", nil, time.Now())
	if err != nil {
		t.Fatalf("failed to create DNS record: %v", err)
	}

	editForm := modals.NewDNSEditForm(rec)
	editRendered := editForm.View()
	if editRendered == "" {
		t.Fatal("expected non-empty render from edit form")
	}

	_, submitCmd := editForm.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if submitCmd == nil {
		t.Fatal("expected submit cmd on ctrl+s")
	}
	submitMsg := submitCmd()
	sub, ok := submitMsg.(modals.DNSFormSubmitMsg)
	if !ok {
		t.Fatalf("expected DNSFormSubmitMsg, got %T", submitMsg)
	}
	if sub.Record.Name != "api.example.com" {
		t.Errorf("expected name 'api.example.com', got %s", sub.Record.Name)
	}
	if sub.Record.Content != "1.2.3.4" {
		t.Errorf("expected content '1.2.3.4', got %s", sub.Record.Content)
	}
}

func TestHelpModal(t *testing.T) {
	h := modals.NewHelpModal()
	h.SetSize(80, 24)

	rendered := h.View()
	if rendered == "" {
		t.Fatal("expected non-empty render from help modal")
	}

	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}
	if _, ok := cmd().(modals.CloseHelpModalMsg); !ok {
		t.Errorf("expected CloseHelpModalMsg on Esc")
	}
}

func TestZonePickerModal(t *testing.T) {
	accs := []account.Account{
		{ID: "acc-1", Name: "Personal"},
		{ID: "acc-2", Name: "Work"},
	}
	zones := []zone.Zone{
		{ID: "zone-1", Name: "example.com", AccountID: "acc-1", Plan: "Pro"},
		{ID: "zone-2", Name: "sekai.dev", AccountID: "acc-2", Plan: "Free"},
	}

	picker := modals.NewZonePickerModal(accs, zones, accs[0], zones[0])
	picker.SetSize(100, 30)

	rendered := picker.View()
	if rendered == "" {
		t.Fatal("expected non-empty render from zone picker")
	}

	_, selectCmd := picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if selectCmd == nil {
		t.Fatal("expected select cmd on Enter")
	}
	msg := selectCmd()
	sel, ok := msg.(modals.ZoneSelectedMsg)
	if !ok {
		t.Fatalf("expected ZoneSelectedMsg, got %T", msg)
	}
	if sel.Zone.ID != "zone-1" {
		t.Errorf("expected zone-1, got %s", sel.Zone.ID)
	}

	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyTab})
	if picker.ActiveTab != modals.TabAccounts {
		t.Errorf("expected TabAccounts after Tab key")
	}

	_, accCmd := picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if accCmd == nil {
		t.Fatal("expected command on Enter in Accounts tab")
	}
	accMsg := accCmd()
	accSel, ok := accMsg.(modals.AccountSelectedMsg)
	if !ok {
		t.Fatalf("expected AccountSelectedMsg, got %T", accMsg)
	}
	if accSel.Account.ID != "acc-1" {
		t.Errorf("expected acc-1, got %s", accSel.Account.ID)
	}

	_, closeCmd := picker.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if closeCmd == nil {
		t.Fatal("expected close cmd on Esc")
	}
	if _, ok := closeCmd().(modals.ZonePickerCloseMsg); !ok {
		t.Errorf("expected ZonePickerCloseMsg")
	}
}
