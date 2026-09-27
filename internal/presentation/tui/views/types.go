package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/core/domain/account"
	"github.com/sekai-labs/kumo/internal/core/domain/analytics"
	"github.com/sekai-labs/kumo/internal/core/domain/dns"
	"github.com/sekai-labs/kumo/internal/core/domain/pages"
	"github.com/sekai-labs/kumo/internal/core/domain/ruleset"
	"github.com/sekai-labs/kumo/internal/core/domain/tunnel"
	"github.com/sekai-labs/kumo/internal/core/domain/worker"
	"github.com/sekai-labs/kumo/internal/core/domain/zone"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
)

const (
	ViewOverview  = "overview"
	ViewDNS       = "dns"
	ViewTunnels   = "tunnels"
	ViewWorkers   = "workers"
	ViewRulesets  = "rulesets"
	ViewAnalytics = "analytics"
)

type Shortcut struct {
	Key  string
	Desc string
}

type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
	SetDimensions(dims layout.Dimensions)
	Shortcuts() []Shortcut
}

type ZonesLoadedMsg struct {
	Zones []zone.Zone
}

type AccountsLoadedMsg struct {
	Accounts []account.Account
}

type DNSLoadedMsg struct {
	Records []dns.DNSRecord
}

type TunnelsLoadedMsg struct {
	Tunnels []tunnel.Tunnel
}

type WorkersLoadedMsg struct {
	Workers []worker.WorkerScript
}

type PagesLoadedMsg struct {
	Pages []pages.PagesProject
}

type RulesetsLoadedMsg struct {
	Rulesets []ruleset.Ruleset
}

type AnalyticsLoadedMsg struct {
	Summary analytics.TrafficSummary
}

type ErrorMsg struct {
	Err error
}

type StatusMsg struct {
	Message string
	IsError bool
}

type ActiveZoneChangedMsg struct {
	Zone zone.Zone
}

type ActiveAccountChangedMsg struct {
	Account account.Account
}

type TokenVerifiedMsg struct {
	TokenInfo account.TokenInfo
}

type OpenNewRecordModalMsg struct {
	ZoneID zone.ZoneID
}

type OpenEditRecordModalMsg struct {
	Record dns.DNSRecord
}

type OpenDeleteRecordModalMsg struct {
	Record dns.DNSRecord
}

type ToggleDNSProxyReqMsg struct {
	ZoneID   zone.ZoneID
	RecordID dns.RecordID
}

type RecordCopiedMsg struct {
	Content string
}

type SelectNavMsg struct {
	Index int
	ID    string
}
