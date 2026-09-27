package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sekai-labs/kumo/internal/core/application"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/core/ports"
	"github.com/sekai-labs/kumo/internal/presentation/tui/command"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
	"github.com/sekai-labs/kumo/internal/presentation/tui/modals"
	"github.com/sekai-labs/kumo/internal/presentation/tui/theme"
	"github.com/sekai-labs/kumo/internal/presentation/tui/views"
)

type ModalState int

const (
	ModalNone ModalState = iota
	ModalConfirm
	ModalDNSForm
	ModalHelp
	ModalZonePicker
	ModalCommandPalette
)

type AppModel struct {
	service   *application.AppService
	profile   string
	layoutMgr *layout.Manager
	dims      layout.Dimensions
	theme     *theme.Theme
	registry  *command.Registry

	activeViewID string
	activeModal  ModalState
	loading      bool
	err          error

	tokenInfo account.TokenInfo
	accounts  []account.Account
	activeAcc account.Account
	zones     []zone.Zone
	activeZn  zone.Zone

	nav    *views.Nav
	header *views.Header
	footer *views.Footer

	views map[string]views.View

	confirmModal *modals.ConfirmModal
	dnsModal     *modals.DNSFormModal
	helpModal    *modals.HelpModal
	zoneModal    *modals.ZonePickerModal
	paletteModal *command.Palette
}

type Config struct {
	Service        *application.AppService
	Profile        string
	InitialAccount string
	InitialZone    string
}

func New(cfg Config) *AppModel {
	reg := command.NewRegistry()
	lm := layout.NewManager()
	th := theme.Current()

	navComp := views.NewNav()
	headerComp := views.NewHeader()
	if cfg.Profile != "" {
		headerComp.SetProfile(cfg.Profile)
	}
	footerComp := views.NewFooter()

	viewMap := make(map[string]views.View)
	viewMap[views.ViewOverview] = views.NewOverviewView()
	viewMap[views.ViewDNS] = views.NewDNSView()
	viewMap[views.ViewTunnels] = views.NewTunnelView()
	viewMap[views.ViewWorkers] = views.NewWorkersPagesView()
	viewMap[views.ViewRulesets] = views.NewRulesetsView()
	viewMap[views.ViewAnalytics] = views.NewAnalyticsView()

	m := &AppModel{
		service:      cfg.Service,
		profile:      cfg.Profile,
		layoutMgr:    lm,
		theme:        th,
		registry:     reg,
		activeViewID: views.ViewOverview,
		activeModal:  ModalNone,
		nav:          navComp,
		header:       headerComp,
		footer:       footerComp,
		views:        viewMap,
		helpModal:    modals.NewHelpModal(),
		paletteModal: command.NewPalette(reg),
	}

	return m
}

func (m *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.header.Init())

	if m.service != nil {
		cmds = append(cmds, m.fetchTokenInfoCmd(false))
		cmds = append(cmds, m.fetchAccountsCmd(false))
	}

	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.dims = m.layoutMgr.HandleWindowSize(msg)
		m.updateDimensions()

		for _, v := range m.views {
			v.SetDimensions(m.dims)
		}
		if m.confirmModal != nil {
			m.confirmModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		}
		if m.dnsModal != nil {
			m.dnsModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		}
		m.helpModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		if m.zoneModal != nil {
			m.zoneModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		}
		m.paletteModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		return m, nil

	case views.TokenVerifiedMsg:
		m.tokenInfo = msg.TokenInfo
		m.header.SetTokenStatus(msg.TokenInfo.Status == "active", string(msg.TokenInfo.Status))
		return m, nil

	case views.AccountsLoadedMsg:
		m.accounts = msg.Accounts
		if len(m.accounts) > 0 && m.activeAcc.ID == "" {
			m.activeAcc = m.accounts[0]
			m.header.SetAccount(m.activeAcc.Name, string(m.activeAcc.ID))
			cmds = append(cmds, m.fetchZonesCmd(m.activeAcc.ID, false))
		}
		m.broadcastToViews(msg)
		return m, tea.Batch(cmds...)

	case views.ZonesLoadedMsg:
		m.zones = msg.Zones
		m.nav.SetBadge(views.ViewOverview, fmt.Sprintf("%d", len(m.zones)))
		if len(m.zones) > 0 && m.activeZn.ID == "" {
			m.activeZn = m.zones[0]
			m.header.SetZone(m.activeZn.Name)
			cmds = append(cmds, m.fetchCurrentViewDataCmd(false))
		}
		m.broadcastToViews(msg)
		return m, tea.Batch(cmds...)

	case views.DNSLoadedMsg:
		m.nav.SetBadge(views.ViewDNS, fmt.Sprintf("%d", len(msg.Records)))
		m.broadcastToViews(msg)
		return m, nil

	case views.TunnelsLoadedMsg:
		m.nav.SetBadge(views.ViewTunnels, fmt.Sprintf("%d", len(msg.Tunnels)))
		m.broadcastToViews(msg)
		return m, nil

	case views.WorkersLoadedMsg:
		m.broadcastToViews(msg)
		return m, nil

	case views.PagesLoadedMsg:
		m.broadcastToViews(msg)
		return m, nil

	case views.RulesetsLoadedMsg:
		m.nav.SetBadge(views.ViewRulesets, fmt.Sprintf("%d", len(msg.Rulesets)))
		m.broadcastToViews(msg)
		return m, nil

	case views.AnalyticsLoadedMsg:
		m.broadcastToViews(msg)
		return m, nil

	case views.StatusMsg:
		m.footer.SetStatus(msg.Message, msg.IsError, 4*time.Second)
		return m, nil

	case views.ErrorMsg:
		m.err = msg.Err
		m.footer.SetStatus("Error: "+msg.Err.Error(), true, 6*time.Second)
		return m, nil

	case views.RecordCopiedMsg:
		m.footer.SetStatus("Yanked: "+msg.Content, false, 3*time.Second)
		return m, nil

	case command.SwitchViewMsg:
		return m, m.switchView(msg.View)

	case views.SelectNavMsg:
		return m, m.switchView(msg.ID)

	case command.OpenCommandPaletteMsg:
		m.paletteModal.Reset()
		m.activeModal = ModalCommandPalette
		return m, m.paletteModal.Init()

	case command.ToggleHelpMsg:
		if m.activeModal == ModalHelp {
			m.activeModal = ModalNone
		} else {
			m.activeModal = ModalHelp
		}
		return m, nil

	case views.OpenNewRecordModalMsg:
		m.dnsModal = modals.NewDNSCreateForm(m.activeZn.ID)
		m.dnsModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		m.activeModal = ModalDNSForm
		return m, m.dnsModal.Init()

	case views.OpenEditRecordModalMsg:
		m.dnsModal = modals.NewDNSEditForm(msg.Record)
		m.dnsModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		m.activeModal = ModalDNSForm
		return m, m.dnsModal.Init()

	case views.OpenDeleteRecordModalMsg:
		rec := msg.Record
		m.confirmModal = modals.NewConfirmModal(
			string(rec.ID),
			"Delete DNS Record",
			fmt.Sprintf("Are you sure you want to delete %s record '%s' (%s)?", rec.Type, rec.Name, rec.Content),
			rec,
		)
		m.confirmModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
		m.activeModal = ModalConfirm
		return m, m.confirmModal.Init()

	case modals.ConfirmResultMsg:
		m.activeModal = ModalNone
		if msg.Confirmed {
			if rec, ok := msg.Data.(dns.DNSRecord); ok {
				return m, m.deleteRecordCmd(rec.ZoneID, rec.ID)
			}
		}
		return m, nil

	case modals.DNSFormCancelMsg:
		m.activeModal = ModalNone
		return m, nil

	case modals.DNSFormSubmitMsg:
		m.activeModal = ModalNone
		if msg.Mode == modals.DNSFormCreate {
			return m, m.createRecordCmd(msg.Record.ZoneID, msg.Record)
		} else {
			return m, m.updateRecordCmd(msg.Record.ZoneID, msg.Record)
		}

	case modals.CloseHelpModalMsg:
		m.activeModal = ModalNone
		return m, nil

	case command.PaletteCloseMsg:
		m.activeModal = ModalNone
		return m, nil

	case modals.ZonePickerCloseMsg:
		m.activeModal = ModalNone
		return m, nil

	case modals.ZoneSelectedMsg:
		m.activeModal = ModalNone
		m.activeZn = msg.Zone
		m.header.SetZone(msg.Zone.Name)
		if msg.Account.ID != "" {
			m.activeAcc = msg.Account
			m.header.SetAccount(msg.Account.Name, string(msg.Account.ID))
		}
		m.broadcastToViews(views.ActiveZoneChangedMsg{Zone: msg.Zone})
		return m, m.fetchCurrentViewDataCmd(false)

	case modals.AccountSelectedMsg:
		m.activeModal = ModalNone
		m.activeAcc = msg.Account
		m.header.SetAccount(msg.Account.Name, string(msg.Account.ID))
		m.broadcastToViews(views.ActiveAccountChangedMsg{Account: msg.Account})
		return m, m.fetchZonesCmd(msg.Account.ID, false)

	case views.ToggleDNSProxyReqMsg:
		return m, m.toggleProxyCmd(msg.ZoneID, msg.RecordID)

	case command.RefreshMsg:
		return m, m.refreshAllCmd()
	}

	if m.activeModal != ModalNone {
		return m.updateModal(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		k := keyMsg.String()
		switch k {
		case "ctrl+c", "q":
			return m, tea.Quit
		case ":":
			m.paletteModal.Reset()
			m.activeModal = ModalCommandPalette
			return m, m.paletteModal.Init()
		case "?":
			m.activeModal = ModalHelp
			return m, nil
		case "r":
			return m, m.refreshAllCmd()
		case "Z":
			m.zoneModal = modals.NewZonePickerModal(m.accounts, m.zones, m.activeAcc, m.activeZn)
			m.zoneModal.SetSize(m.dims.TotalWidth, m.dims.TotalHeight)
			m.activeModal = ModalZonePicker
			return m, m.zoneModal.Init()
		case "1":
			return m, m.switchView(views.ViewOverview)
		case "2":
			return m, m.switchView(views.ViewDNS)
		case "3":
			return m, m.switchView(views.ViewTunnels)
		case "4":
			return m, m.switchView(views.ViewWorkers)
		case "5":
			return m, m.switchView(views.ViewRulesets)
		case "6":
			return m, m.switchView(views.ViewAnalytics)
		case "tab":

			if m.dims.Breakpoint == layout.BreakpointCompact {
				switch m.dims.ActivePane {
				case layout.PaneNav:
					m.dims = m.layoutMgr.SetActivePane(layout.PaneList)
				case layout.PaneList:
					m.dims = m.layoutMgr.SetActivePane(layout.PaneDetail)
				case layout.PaneDetail:
					m.dims = m.layoutMgr.SetActivePane(layout.PaneNav)
				}
				m.updateDimensions()
			}
		}
	}

	if curView, exists := m.views[m.activeViewID]; exists {
		var viewCmd tea.Cmd
		updatedView, viewCmd := curView.Update(msg)
		m.views[m.activeViewID] = updatedView
		cmds = append(cmds, viewCmd)

		m.footer.SetShortcuts(updatedView.Shortcuts())
	}

	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(msg)
	cmds = append(cmds, headerCmd)

	return m, tea.Batch(cmds...)
}

func (m *AppModel) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeModal {
	case ModalConfirm:
		if m.confirmModal != nil {
			var cmd tea.Cmd
			m.confirmModal, cmd = m.confirmModal.Update(msg)
			return m, cmd
		}
	case ModalDNSForm:
		if m.dnsModal != nil {
			var cmd tea.Cmd
			m.dnsModal, cmd = m.dnsModal.Update(msg)
			return m, cmd
		}
	case ModalHelp:
		var cmd tea.Cmd
		m.helpModal, cmd = m.helpModal.Update(msg)
		return m, cmd
	case ModalZonePicker:
		if m.zoneModal != nil {
			var cmd tea.Cmd
			m.zoneModal, cmd = m.zoneModal.Update(msg)
			return m, cmd
		}
	case ModalCommandPalette:
		var cmd tea.Cmd
		m.paletteModal, cmd = m.paletteModal.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *AppModel) switchView(viewID string) tea.Cmd {
	if _, ok := m.views[viewID]; !ok {
		return nil
	}
	m.activeViewID = viewID
	m.nav.SelectByID(viewID)

	switch viewID {
	case views.ViewOverview:
		m.header.SetBreadcrumb("Overview")
	case views.ViewDNS:
		m.header.SetBreadcrumb("DNS Records")
	case views.ViewTunnels:
		m.header.SetBreadcrumb("Cloudflare Tunnels")
	case views.ViewWorkers:
		m.header.SetBreadcrumb("Workers & Pages")
	case views.ViewRulesets:
		m.header.SetBreadcrumb("Rulesets")
	case views.ViewAnalytics:
		m.header.SetBreadcrumb("Traffic Analytics")
	}

	if curView, ok := m.views[viewID]; ok {
		m.footer.SetShortcuts(curView.Shortcuts())
	}

	return m.fetchCurrentViewDataCmd(false)
}

func (m *AppModel) updateDimensions() {
	m.header.SetDimensions(m.dims.Header.Width, m.dims.Header.Height)
	m.nav.SetDimensions(m.dims.Nav.Width, m.dims.Nav.Height)
	m.footer.SetDimensions(m.dims.Footer.Width, m.dims.Footer.Height)
}

func (m *AppModel) broadcastToViews(msg tea.Msg) {
	for k, v := range m.views {
		updated, _ := v.Update(msg)
		m.views[k] = updated
	}
}

func (m *AppModel) View() string {
	if m.dims.TotalWidth <= 0 || m.dims.TotalHeight <= 0 {
		return ""
	}

	headerView := m.header.View()
	footerView := m.footer.View()

	var contentView string
	if m.dims.Breakpoint == layout.BreakpointCompact {

		switch m.dims.ActivePane {
		case layout.PaneNav:
			contentView = m.nav.View()
		default:
			if curView, exists := m.views[m.activeViewID]; exists {
				contentView = curView.View()
			}
		}
	} else {

		navView := m.nav.View()
		var mainView string
		if curView, exists := m.views[m.activeViewID]; exists {
			mainView = curView.View()
		}
		contentView = lipgloss.JoinHorizontal(lipgloss.Top, navView, mainView)
	}

	baseApp := lipgloss.JoinVertical(
		lipgloss.Left,
		headerView,
		contentView,
		footerView,
	)

	if m.activeModal != ModalNone {
		var modalBox string
		switch m.activeModal {
		case ModalConfirm:
			if m.confirmModal != nil {
				modalBox = m.confirmModal.View()
			}
		case ModalDNSForm:
			if m.dnsModal != nil {
				modalBox = m.dnsModal.View()
			}
		case ModalHelp:
			modalBox = m.helpModal.View()
		case ModalZonePicker:
			if m.zoneModal != nil {
				modalBox = m.zoneModal.View()
			}
		case ModalCommandPalette:
			modalBox = m.paletteModal.View()
		}

		if modalBox != "" {
			return OverlayCenter(baseApp, modalBox, m.dims.TotalWidth, m.dims.TotalHeight)
		}
	}

	return baseApp
}

func OverlayCenter(base, overlay string, width, height int) string {
	if width <= 0 || height <= 0 {
		return overlay
	}

	overlayLines := strings.Split(overlay, "\n")
	overlayHeight := len(overlayLines)
	overlayWidth := 0
	for _, l := range overlayLines {
		w := lipgloss.Width(l)
		if w > overlayWidth {
			overlayWidth = w
		}
	}

	topPad := (height - overlayHeight) / 2
	if topPad < 0 {
		topPad = 0
	}
	leftPad := (width - overlayWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	placedOverlay := lipgloss.NewStyle().
		MarginLeft(leftPad).
		MarginTop(topPad).
		Render(overlay)

	return placeOver(base, placedOverlay, width, height, leftPad, topPad, overlayWidth, overlayHeight)
}

func placeOver(base, overlay string, totalW, totalH, startX, startY, overW, overH int) string {
	baseLines := strings.Split(base, "\n")
	overLines := strings.Split(overlay, "\n")

	for len(baseLines) < totalH {
		baseLines = append(baseLines, strings.Repeat(" ", totalW))
	}

	var result []string
	for y := 0; y < totalH; y++ {
		if y >= len(baseLines) {
			break
		}
		if y >= startY && y < startY+overH && (y-startY) < len(overLines) {

			oLine := overLines[y-startY]

			result = append(result, oLine)
		} else {
			result = append(result, baseLines[y])
		}
	}

	return strings.Join(result, "\n")
}

func (m *AppModel) fetchTokenInfoCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		info, err := m.service.VerifyToken(context.Background(), force)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return views.TokenVerifiedMsg{TokenInfo: info}
	}
}

func (m *AppModel) fetchAccountsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		accs, err := m.service.ListAccounts(context.Background(), force)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return views.AccountsLoadedMsg{Accounts: accs}
	}
}

func (m *AppModel) fetchZonesCmd(accountID account.AccountID, force bool) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		zones, err := m.service.ListZones(context.Background(), accountID, force)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return views.ZonesLoadedMsg{Zones: zones}
	}
}

func (m *AppModel) fetchCurrentViewDataCmd(force bool) tea.Cmd {
	if m.service == nil {
		return nil
	}

	switch m.activeViewID {
	case views.ViewOverview:
		var cmds []tea.Cmd
		if m.activeAcc.ID != "" {
			cmds = append(cmds, func() tea.Msg {
				tuns, err := m.service.ListTunnels(context.Background(), m.activeAcc.ID, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.TunnelsLoadedMsg{Tunnels: tuns}
			})
			cmds = append(cmds, func() tea.Msg {
				wrk, err := m.service.ListWorkers(context.Background(), m.activeAcc.ID, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.WorkersLoadedMsg{Workers: wrk}
			})
			cmds = append(cmds, func() tea.Msg {
				pgs, err := m.service.ListPages(context.Background(), m.activeAcc.ID, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.PagesLoadedMsg{Pages: pgs}
			})
		}
		if m.activeZn.ID != "" {
			cmds = append(cmds, func() tea.Msg {
				recs, err := m.service.ListDNSRecords(context.Background(), m.activeZn.ID, ports.DNSFilter{}, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.DNSLoadedMsg{Records: recs}
			})
		}
		return tea.Batch(cmds...)

	case views.ViewDNS:
		if m.activeZn.ID == "" {
			return nil
		}
		return func() tea.Msg {
			recs, err := m.service.ListDNSRecords(context.Background(), m.activeZn.ID, ports.DNSFilter{}, force)
			if err != nil {
				return views.ErrorMsg{Err: err}
			}
			return views.DNSLoadedMsg{Records: recs}
		}

	case views.ViewTunnels:
		if m.activeAcc.ID == "" {
			return nil
		}
		return func() tea.Msg {
			tuns, err := m.service.ListTunnels(context.Background(), m.activeAcc.ID, force)
			if err != nil {
				return views.ErrorMsg{Err: err}
			}
			return views.TunnelsLoadedMsg{Tunnels: tuns}
		}

	case views.ViewWorkers:
		if m.activeAcc.ID == "" {
			return nil
		}
		return tea.Batch(
			func() tea.Msg {
				wrk, err := m.service.ListWorkers(context.Background(), m.activeAcc.ID, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.WorkersLoadedMsg{Workers: wrk}
			},
			func() tea.Msg {
				pgs, err := m.service.ListPages(context.Background(), m.activeAcc.ID, force)
				if err != nil {
					return views.ErrorMsg{Err: err}
				}
				return views.PagesLoadedMsg{Pages: pgs}
			},
		)

	case views.ViewRulesets:
		if m.activeZn.ID == "" {
			return nil
		}
		return func() tea.Msg {
			rules, err := m.service.ListZoneRulesets(context.Background(), m.activeZn.ID, force)
			if err != nil {
				return views.ErrorMsg{Err: err}
			}
			return views.RulesetsLoadedMsg{Rulesets: rules}
		}

	case views.ViewAnalytics:
		if m.activeZn.ID == "" {
			return nil
		}
		return func() tea.Msg {
			summary, err := m.service.GetZoneTrafficSummary(context.Background(), m.activeZn.ID, 24*time.Hour, force)
			if err != nil {
				return views.ErrorMsg{Err: err}
			}
			return views.AnalyticsLoadedMsg{Summary: summary}
		}
	}

	return nil
}

func (m *AppModel) refreshAllCmd() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		return views.StatusMsg{Message: "Refreshed from API", IsError: false}
	})
	if m.service != nil {
		cmds = append(cmds, m.fetchTokenInfoCmd(true))
		cmds = append(cmds, m.fetchAccountsCmd(true))
		if m.activeAcc.ID != "" {
			cmds = append(cmds, m.fetchZonesCmd(m.activeAcc.ID, true))
		}
		cmds = append(cmds, m.fetchCurrentViewDataCmd(true))
	}
	return tea.Batch(cmds...)
}

func (m *AppModel) toggleProxyCmd(zoneID zone.ZoneID, recID dns.RecordID) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		rec, err := m.service.ToggleDNSProxy(context.Background(), zoneID, recID)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		status := "DNS Only"
		if rec.Proxied {
			status = "Proxied"
		}
		return views.StatusMsg{
			Message: fmt.Sprintf("Updated %s proxy status to: %s", rec.Name, status),
			IsError: false,
		}
	}
}

func (m *AppModel) createRecordCmd(zoneID zone.ZoneID, record dns.DNSRecord) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		rec, err := m.service.CreateDNSRecord(context.Background(), zoneID, record)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return tea.Batch(
			func() tea.Msg {
				return views.StatusMsg{
					Message: fmt.Sprintf("Created %s record: %s", rec.Type, rec.Name),
					IsError: false,
				}
			},
			m.fetchCurrentViewDataCmd(true),
		)()
	}
}

func (m *AppModel) updateRecordCmd(zoneID zone.ZoneID, record dns.DNSRecord) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		rec, err := m.service.UpdateDNSRecord(context.Background(), zoneID, record)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return tea.Batch(
			func() tea.Msg {
				return views.StatusMsg{
					Message: fmt.Sprintf("Updated %s record: %s", rec.Type, rec.Name),
					IsError: false,
				}
			},
			m.fetchCurrentViewDataCmd(true),
		)()
	}
}

func (m *AppModel) deleteRecordCmd(zoneID zone.ZoneID, recID dns.RecordID) tea.Cmd {
	return func() tea.Msg {
		if m.service == nil {
			return nil
		}
		err := m.service.DeleteDNSRecord(context.Background(), zoneID, recID)
		if err != nil {
			return views.ErrorMsg{Err: err}
		}
		return tea.Batch(
			func() tea.Msg {
				return views.StatusMsg{
					Message: fmt.Sprintf("Deleted DNS record: %s", recID),
					IsError: false,
				}
			},
			m.fetchCurrentViewDataCmd(true),
		)()
	}
}
