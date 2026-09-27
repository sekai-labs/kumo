package modals

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
)

type DNSFormMode int

const (
	DNSFormCreate DNSFormMode = iota
	DNSFormEdit
)

type DNSFormSubmitMsg struct {
	Mode   DNSFormMode
	Record dns.DNSRecord
}

type DNSFormCancelMsg struct{}

type DNSFormField int

const (
	FieldType DNSFormField = iota
	FieldName
	FieldContent
	FieldProxied
	FieldTTL
	FieldComment
	FieldCount
)

var allowedTypes = []dns.RecordType{
	dns.TypeA,
	dns.TypeAAAA,
	dns.TypeCNAME,
	dns.TypeTXT,
	dns.TypeMX,
	dns.TypeNS,
	dns.TypeSRV,
	dns.TypeCAA,
}

type DNSFormModal struct {
	Mode         DNSFormMode
	ZoneID       zone.ZoneID
	RecordID     dns.RecordID
	focusedField DNSFormField

	typeIndex int
	nameInput textinput.Model
	contInput textinput.Model
	proxied   bool
	ttlInput  textinput.Model
	commInput textinput.Model

	errMessage string
	Width      int
	Height     int
}

func NewDNSCreateForm(zoneID zone.ZoneID) *DNSFormModal {
	m := &DNSFormModal{
		Mode:         DNSFormCreate,
		ZoneID:       zoneID,
		focusedField: FieldType,
		typeIndex:    0,
		proxied:      true,
		Width:        64,
	}
	m.initInputs("", "", "1", "")
	return m
}

func NewDNSEditForm(rec dns.DNSRecord) *DNSFormModal {
	m := &DNSFormModal{
		Mode:         DNSFormEdit,
		ZoneID:       rec.ZoneID,
		RecordID:     rec.ID,
		focusedField: FieldName,
		proxied:      rec.Proxied,
		Width:        64,
	}

	for i, t := range allowedTypes {
		if t == rec.Type {
			m.typeIndex = i
			break
		}
	}

	ttlStr := "1"
	if rec.TTL > 1 {
		ttlStr = strconv.Itoa(int(rec.TTL))
	}

	m.initInputs(rec.Name, rec.Content, ttlStr, rec.Comment)
	return m
}

func (m *DNSFormModal) initInputs(name, content, ttl, comment string) {
	ni := textinput.New()
	ni.Placeholder = "subdomain (e.g. api or @)"
	ni.SetValue(name)
	ni.Prompt = ""
	ni.CharLimit = 255
	m.nameInput = ni

	ci := textinput.New()
	ci.Placeholder = "target IP or destination hostname"
	ci.SetValue(content)
	ci.Prompt = ""
	ci.CharLimit = 1024
	m.contInput = ci

	ti := textinput.New()
	ti.Placeholder = "1 for Auto, or 60-86400"
	ti.SetValue(ttl)
	ti.Prompt = ""
	ti.CharLimit = 8
	m.ttlInput = ti

	commi := textinput.New()
	commi.Placeholder = "optional notes"
	commi.SetValue(comment)
	commi.Prompt = ""
	commi.CharLimit = 255
	m.commInput = commi
}

func (m *DNSFormModal) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *DNSFormModal) Init() tea.Cmd {
	return nil
}

func (m *DNSFormModal) Update(msg tea.Msg) (*DNSFormModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return DNSFormCancelMsg{}
			}
		case "tab", "down":
			m.nextField()
			return m, nil
		case "shift+tab", "up":
			m.prevField()
			return m, nil
		case "enter":

			if m.focusedField == FieldComment || m.focusedField == FieldTTL {
				return m, m.submit()
			}
			m.nextField()
			return m, nil
		case "ctrl+s":
			return m, m.submit()
		}

		switch m.focusedField {
		case FieldType:
			switch msg.String() {
			case "left", "h":
				if m.typeIndex > 0 {
					m.typeIndex--
					m.updateProxySupport()
				}
				return m, nil
			case "right", "l", " ":
				if m.typeIndex < len(allowedTypes)-1 {
					m.typeIndex++
					m.updateProxySupport()
				} else if msg.String() == " " {
					m.typeIndex = 0
					m.updateProxySupport()
				}
				return m, nil
			}
		case FieldProxied:
			switch msg.String() {
			case " ", "p", "enter":
				selectedType := allowedTypes[m.typeIndex]
				if dns.ProxiableTypes[selectedType] {
					m.proxied = !m.proxied
				}
				return m, nil
			}
		case FieldName:
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd
		case FieldContent:
			var cmd tea.Cmd
			m.contInput, cmd = m.contInput.Update(msg)
			return m, cmd
		case FieldTTL:
			var cmd tea.Cmd
			m.ttlInput, cmd = m.ttlInput.Update(msg)
			return m, cmd
		case FieldComment:
			var cmd tea.Cmd
			m.commInput, cmd = m.commInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *DNSFormModal) nextField() {
	m.focusedField = (m.focusedField + 1) % FieldCount
	m.syncFocus()
}

func (m *DNSFormModal) prevField() {
	if m.focusedField == 0 {
		m.focusedField = FieldCount - 1
	} else {
		m.focusedField--
	}
	m.syncFocus()
}

func (m *DNSFormModal) syncFocus() {
	m.nameInput.Blur()
	m.contInput.Blur()
	m.ttlInput.Blur()
	m.commInput.Blur()

	switch m.focusedField {
	case FieldName:
		m.nameInput.Focus()
	case FieldContent:
		m.contInput.Focus()
	case FieldTTL:
		m.ttlInput.Focus()
	case FieldComment:
		m.commInput.Focus()
	}
}

func (m *DNSFormModal) updateProxySupport() {
	selectedType := allowedTypes[m.typeIndex]
	if !dns.ProxiableTypes[selectedType] {
		m.proxied = false
	}
}

func (m *DNSFormModal) submit() tea.Cmd {
	selectedType := allowedTypes[m.typeIndex]
	name := strings.TrimSpace(m.nameInput.Value())
	content := strings.TrimSpace(m.contInput.Value())
	ttlStr := strings.TrimSpace(m.ttlInput.Value())
	comment := strings.TrimSpace(m.commInput.Value())

	if name == "" {
		m.errMessage = "Record name cannot be empty"
		return nil
	}
	if content == "" {
		m.errMessage = "Record content cannot be empty"
		return nil
	}

	ttlVal := 1
	if ttlStr != "" && ttlStr != "1" && strings.ToLower(ttlStr) != "auto" {
		parsed, err := strconv.Atoi(ttlStr)
		if err != nil || (parsed != 1 && (parsed < 60 || parsed > 86400)) {
			m.errMessage = "TTL must be 1 (Auto) or between 60 and 86400"
			return nil
		}
		ttlVal = parsed
	}

	recordID := m.RecordID
	if m.Mode == DNSFormCreate {
		recordID = dns.RecordID(fmt.Sprintf("temp-%d", time.Now().UnixNano()))
	}

	rec, err := dns.NewDNSRecord(
		string(recordID),
		m.ZoneID,
		selectedType,
		name,
		content,
		m.proxied,
		ttlVal,
		comment,
		nil,
		time.Now(),
	)
	if err != nil {
		m.errMessage = err.Error()
		return nil
	}

	m.errMessage = ""
	return func() tea.Msg {
		return DNSFormSubmitMsg{
			Mode:   m.Mode,
			Record: rec,
		}
	}
}

func (m *DNSFormModal) View() string {
	modalWidth := 62
	if m.Width > 0 && m.Width < 68 {
		modalWidth = m.Width - 4
		if modalWidth < 40 {
			modalWidth = 40
		}
	}
	contentWidth := modalWidth - 4

	titleText := "Add New DNS Record"
	if m.Mode == DNSFormEdit {
		titleText = fmt.Sprintf("Edit DNS Record (%s)", m.RecordID)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ColorCFOrange).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Width(12).
		Foreground(theme.ColorCFBlue).
		Bold(true)

	fieldFocused := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.ColorCFOrange).
		Width(contentWidth - 14)

	fieldUnfocused := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.ColorBorder).
		Width(contentWidth - 14)

	var typeItems []string
	for i, t := range allowedTypes {
		if i == m.typeIndex {
			typeItems = append(typeItems, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(theme.ColorCFBlue).
				Bold(true).
				Padding(0, 1).
				Render(string(t)))
		} else {
			typeItems = append(typeItems, lipgloss.NewStyle().
				Foreground(theme.ColorTextDim).
				Padding(0, 1).
				Render(string(t)))
		}
	}
	typeRow := lipgloss.JoinHorizontal(lipgloss.Center, typeItems...)
	if m.focusedField == FieldType {
		typeRow = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(theme.ColorCFOrange).
			Render(typeRow + "  ←/→ to select")
	}

	var proxyDisplay string
	selectedType := allowedTypes[m.typeIndex]
	if !dns.ProxiableTypes[selectedType] {
		proxyDisplay = theme.RenderMuted("Not supported for " + string(selectedType))
	} else if m.proxied {
		proxyDisplay = theme.BadgeHealthy("Proxied (Orange Cloud)") + " [Space to toggle]"
	} else {
		proxyDisplay = theme.BadgeDegraded("DNS Only (Grey Cloud)") + " [Space to toggle]"
	}

	if m.focusedField == FieldProxied {
		proxyDisplay = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ColorCFOrange).
			Render("> " + proxyDisplay)
	}

	renderRow := func(label string, content string, isFocused bool) string {
		style := fieldUnfocused
		if isFocused {
			style = fieldFocused
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, labelStyle.Render(label), style.Render(content))
	}

	var rows []string
	rows = append(rows, renderRow("Type:", typeRow, m.focusedField == FieldType))
	rows = append(rows, renderRow("Name:", m.nameInput.View(), m.focusedField == FieldName))
	rows = append(rows, renderRow("Content:", m.contInput.View(), m.focusedField == FieldContent))
	rows = append(rows, renderRow("Proxy:", proxyDisplay, m.focusedField == FieldProxied))
	rows = append(rows, renderRow("TTL:", m.ttlInput.View(), m.focusedField == FieldTTL))
	rows = append(rows, renderRow("Comment:", m.commInput.View(), m.focusedField == FieldComment))

	var errView string
	if m.errMessage != "" {
		errView = lipgloss.NewStyle().
			Foreground(theme.ColorStatusDown).
			Bold(true).
			Render("! " + m.errMessage)
	}

	helpLine := theme.RenderMuted("[Tab] Next  [Shift+Tab] Prev  [Enter/Ctrl+S] Save  [Esc] Cancel")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(titleText),
		lipgloss.JoinVertical(lipgloss.Left, rows...),
		errView,
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
