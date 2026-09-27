package layout

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Breakpoint string

const (
	BreakpointWide    Breakpoint = "wide"
	BreakpointMedium  Breakpoint = "medium"
	BreakpointCompact Breakpoint = "compact"
)

type PaneID string

const (
	PaneNav    PaneID = "nav"
	PaneList   PaneID = "list"
	PaneDetail PaneID = "detail"
)

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r Rect) InnerWidth() int {
	if r.Width <= 2 {
		return 0
	}
	return r.Width - 2
}

func (r Rect) InnerHeight() int {
	if r.Height <= 2 {
		return 0
	}
	return r.Height - 2
}

type Dimensions struct {
	TotalWidth  int
	TotalHeight int

	Breakpoint Breakpoint

	Header Rect
	Footer Rect

	Nav    Rect
	List   Rect
	Detail Rect

	ActivePane PaneID
}

type Manager struct {
	width      int
	height     int
	activePane PaneID
	dims       Dimensions
}

func NewManager() *Manager {
	return &Manager{
		activePane: PaneList,
	}
}

func (m *Manager) SetSize(width, height int) Dimensions {
	m.width = width
	m.height = height
	m.dims = Calculate(width, height, m.activePane)
	return m.dims
}

func (m *Manager) HandleWindowSize(msg tea.WindowSizeMsg) Dimensions {
	return m.SetSize(msg.Width, msg.Height)
}

func (m *Manager) SetActivePane(pane PaneID) Dimensions {
	m.activePane = pane
	m.dims = Calculate(m.width, m.height, m.activePane)
	return m.dims
}

func (m *Manager) Dimensions() Dimensions {
	return m.dims
}

func DetermineBreakpoint(width int) Breakpoint {
	if width >= 120 {
		return BreakpointWide
	}
	if width >= 80 {
		return BreakpointMedium
	}
	return BreakpointCompact
}

func Calculate(width, height int, activePane PaneID) Dimensions {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	bp := DetermineBreakpoint(width)

	headerHeight := 2
	footerHeight := 1

	if height < 10 {
		headerHeight = 1
		footerHeight = 1
	}

	contentHeight := height - headerHeight - footerHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	headerRect := Rect{
		X:      0,
		Y:      0,
		Width:  width,
		Height: headerHeight,
	}

	footerRect := Rect{
		X:      0,
		Y:      height - footerHeight,
		Width:  width,
		Height: footerHeight,
	}

	var navRect, listRect, detailRect Rect

	switch bp {
	case BreakpointWide:

		navWidth := 24
		if width*20/100 > navWidth {
			navWidth = width * 20 / 100
		}
		if navWidth < 20 {
			navWidth = 20
		}

		remainingWidth := width - navWidth
		detailWidth := remainingWidth * 38 / 100
		listWidth := remainingWidth - detailWidth

		navRect = Rect{
			X:      0,
			Y:      headerHeight,
			Width:  navWidth,
			Height: contentHeight,
		}

		listRect = Rect{
			X:      navWidth,
			Y:      headerHeight,
			Width:  listWidth,
			Height: contentHeight,
		}

		detailRect = Rect{
			X:      navWidth + listWidth,
			Y:      headerHeight,
			Width:  detailWidth,
			Height: contentHeight,
		}

	case BreakpointMedium:

		navWidth := 20
		contentWidth := width - navWidth
		if contentWidth < 20 {
			contentWidth = 20
		}

		detailWidth := contentWidth * 45 / 100
		listWidth := contentWidth - detailWidth

		navRect = Rect{
			X:      0,
			Y:      headerHeight,
			Width:  navWidth,
			Height: contentHeight,
		}

		listRect = Rect{
			X:      navWidth,
			Y:      headerHeight,
			Width:  listWidth,
			Height: contentHeight,
		}

		detailRect = Rect{
			X:      navWidth + listWidth,
			Y:      headerHeight,
			Width:  detailWidth,
			Height: contentHeight,
		}

	case BreakpointCompact:

		contentRect := Rect{
			X:      0,
			Y:      headerHeight,
			Width:  width,
			Height: contentHeight,
		}

		switch activePane {
		case PaneNav:
			navRect = contentRect
			listRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
			detailRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
		case PaneDetail:
			navRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
			listRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
			detailRect = contentRect
		case PaneList:
			fallthrough
		default:
			navRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
			listRect = contentRect
			detailRect = Rect{X: 0, Y: headerHeight, Width: 0, Height: 0}
		}
	}

	return Dimensions{
		TotalWidth:  width,
		TotalHeight: height,
		Breakpoint:  bp,
		Header:      headerRect,
		Footer:      footerRect,
		Nav:         navRect,
		List:        listRect,
		Detail:      detailRect,
		ActivePane:  activePane,
	}
}
