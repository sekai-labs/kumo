package layout_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sekai-labs/kumo/internal/presentation/tui/layout"
)

func TestBreakpoints(t *testing.T) {
	tests := []struct {
		width    int
		expected layout.Breakpoint
	}{
		{140, layout.BreakpointWide},
		{120, layout.BreakpointWide},
		{119, layout.BreakpointMedium},
		{80, layout.BreakpointMedium},
		{79, layout.BreakpointCompact},
		{40, layout.BreakpointCompact},
	}

	for _, tt := range tests {
		bp := layout.DetermineBreakpoint(tt.width)
		if bp != tt.expected {
			t.Errorf("DetermineBreakpoint(%d) = %v; want %v", tt.width, bp, tt.expected)
		}
	}
}

func TestWideLayoutDimensions(t *testing.T) {
	mgr := layout.NewManager()
	dims := mgr.SetSize(140, 40)

	if dims.Breakpoint != layout.BreakpointWide {
		t.Fatalf("expected BreakpointWide, got %v", dims.Breakpoint)
	}

	if dims.Header.Height != 2 || dims.Footer.Height != 1 {
		t.Errorf("unexpected header/footer height: %d, %d", dims.Header.Height, dims.Footer.Height)
	}

	if dims.Nav.Width <= 0 || dims.List.Width <= 0 || dims.Detail.Width <= 0 {
		t.Errorf("all 3 panes should have width > 0, got nav=%d, list=%d, detail=%d",
			dims.Nav.Width, dims.List.Width, dims.Detail.Width)
	}

	sumWidth := dims.Nav.Width + dims.List.Width + dims.Detail.Width
	if sumWidth != 140 {
		t.Errorf("sum of widths = %d, expected 140", sumWidth)
	}

	expectedH := 40 - 2 - 1
	if dims.Nav.Height != expectedH || dims.List.Height != expectedH || dims.Detail.Height != expectedH {
		t.Errorf("pane heights mismatch, expected %d", expectedH)
	}
}

func TestMediumLayoutDimensions(t *testing.T) {
	mgr := layout.NewManager()
	dims := mgr.HandleWindowSize(tea.WindowSizeMsg{Width: 100, Height: 30})

	if dims.Breakpoint != layout.BreakpointMedium {
		t.Fatalf("expected BreakpointMedium, got %v", dims.Breakpoint)
	}

	if dims.Nav.Width != 20 {
		t.Errorf("expected nav width 20, got %d", dims.Nav.Width)
	}

	sumWidth := dims.Nav.Width + dims.List.Width + dims.Detail.Width
	if sumWidth != 100 {
		t.Errorf("sum of widths = %d, expected 100", sumWidth)
	}
}

func TestCompactLayoutDimensions(t *testing.T) {
	mgr := layout.NewManager()
	mgr.SetSize(70, 24)

	dims := mgr.Dimensions()
	if dims.Breakpoint != layout.BreakpointCompact {
		t.Fatalf("expected BreakpointCompact, got %v", dims.Breakpoint)
	}

	if dims.List.Width != 70 {
		t.Errorf("expected list width 70, got %d", dims.List.Width)
	}
	if dims.Nav.Width != 0 || dims.Detail.Width != 0 {
		t.Errorf("inactive panes should have width 0 in compact mode")
	}

	dims = mgr.SetActivePane(layout.PaneDetail)
	if dims.Detail.Width != 70 {
		t.Errorf("expected detail width 70, got %d", dims.Detail.Width)
	}
	if dims.List.Width != 0 || dims.Nav.Width != 0 {
		t.Errorf("inactive panes should have width 0 in compact mode")
	}
}

func TestRectInnerDimensions(t *testing.T) {
	r := layout.Rect{X: 0, Y: 0, Width: 10, Height: 8}
	if r.InnerWidth() != 8 {
		t.Errorf("expected inner width 8, got %d", r.InnerWidth())
	}
	if r.InnerHeight() != 6 {
		t.Errorf("expected inner height 6, got %d", r.InnerHeight())
	}

	zeroR := layout.Rect{X: 0, Y: 0, Width: 1, Height: 1}
	if zeroR.InnerWidth() != 0 || zeroR.InnerHeight() != 0 {
		t.Errorf("expected 0 inner dims for small rect")
	}
}
