package gui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// BreakpointType represents different screen size categories
type BreakpointType int

const (
	BreakpointMobile BreakpointType = iota
	BreakpointTablet
	BreakpointDesktop
	BreakpointWide
)

// Breakpoint values in dp (density-independent pixels)
const (
	BreakpointMobileMax  = 600  // 0-600dp: Mobile phones
	BreakpointTabletMax  = 900  // 600-900dp: Tablets
	BreakpointDesktopMax = 1200 // 900-1200dp: Small desktops
	// >1200dp: Wide desktops
)

// ResponsiveLayout provides responsive layout utilities
type ResponsiveLayout struct {
	currentBreakpoint BreakpointType
}

// NewResponsiveLayout creates a new responsive layout helper
func NewResponsiveLayout() *ResponsiveLayout {
	return &ResponsiveLayout{
		currentBreakpoint: BreakpointDesktop,
	}
}

// GetBreakpoint determines the current breakpoint based on available width
func (rl *ResponsiveLayout) GetBreakpoint(gtx layout.Context) BreakpointType {
	widthDp := gtx.Metric.PxToDp(gtx.Constraints.Max.X)

	if widthDp < BreakpointMobileMax {
		rl.currentBreakpoint = BreakpointMobile
	} else if widthDp < BreakpointTabletMax {
		rl.currentBreakpoint = BreakpointTablet
	} else if widthDp < BreakpointDesktopMax {
		rl.currentBreakpoint = BreakpointDesktop
	} else {
		rl.currentBreakpoint = BreakpointWide
	}

	return rl.currentBreakpoint
}

// IsMobile returns true if current breakpoint is mobile
func (rl *ResponsiveLayout) IsMobile(gtx layout.Context) bool {
	return rl.GetBreakpoint(gtx) == BreakpointMobile
}

// IsTablet returns true if current breakpoint is tablet
func (rl *ResponsiveLayout) IsTablet(gtx layout.Context) bool {
	return rl.GetBreakpoint(gtx) == BreakpointTablet
}

// IsDesktop returns true if current breakpoint is desktop or wider
func (rl *ResponsiveLayout) IsDesktop(gtx layout.Context) bool {
	bp := rl.GetBreakpoint(gtx)
	return bp == BreakpointDesktop || bp == BreakpointWide
}

// GetCardColumns returns the number of columns for stat cards based on breakpoint
func (rl *ResponsiveLayout) GetCardColumns(gtx layout.Context) int {
	switch rl.GetBreakpoint(gtx) {
	case BreakpointMobile:
		return 1 // Stack cards vertically
	case BreakpointTablet:
		return 2 // Two columns
	case BreakpointDesktop:
		return 4 // Four columns
	case BreakpointWide:
		return 4 // Four columns
	default:
		return 4
	}
}

// GetNavigationAxis returns the axis for navigation layout
func (rl *ResponsiveLayout) GetNavigationAxis(gtx layout.Context) layout.Axis {
	if rl.IsMobile(gtx) {
		return layout.Vertical // Stack vertically on mobile
	}
	return layout.Horizontal // Horizontal on larger screens
}

// GetSpacing returns appropriate spacing based on screen size
func (rl *ResponsiveLayout) GetSpacing(gtx layout.Context) unit.Dp {
	switch rl.GetBreakpoint(gtx) {
	case BreakpointMobile:
		return unit.Dp(8) // Tighter spacing on mobile
	case BreakpointTablet:
		return unit.Dp(12)
	default:
		return unit.Dp(16) // Comfortable spacing on desktop
	}
}

// GetInset returns appropriate insets based on screen size
func (rl *ResponsiveLayout) GetInset(gtx layout.Context) layout.Inset {
	spacing := rl.GetSpacing(gtx)
	return layout.UniformInset(spacing)
}

// GetCardInset returns insets for cards based on screen size
func (rl *ResponsiveLayout) GetCardInset(gtx layout.Context) layout.Inset {
	switch rl.GetBreakpoint(gtx) {
	case BreakpointMobile:
		return layout.Inset{
			Top: unit.Dp(8), Bottom: unit.Dp(8),
			Left: unit.Dp(8), Right: unit.Dp(8),
		}
	case BreakpointTablet:
		return layout.Inset{
			Top: unit.Dp(12), Bottom: unit.Dp(12),
			Left: unit.Dp(12), Right: unit.Dp(12),
		}
	default:
		return layout.Inset{
			Top: unit.Dp(16), Bottom: unit.Dp(16),
			Left: unit.Dp(16), Right: unit.Dp(16),
		}
	}
}

// ResponsiveGrid creates a responsive grid layout
type ResponsiveGrid struct {
	Children []layout.Widget
	Columns  int
}

// NewResponsiveGrid creates a new responsive grid
func NewResponsiveGrid(columns int, children ...layout.Widget) *ResponsiveGrid {
	return &ResponsiveGrid{
		Children: children,
		Columns:  columns,
	}
}

// Layout renders the responsive grid
func (rg *ResponsiveGrid) Layout(gtx layout.Context) layout.Dimensions {
	if len(rg.Children) == 0 {
		return layout.Dimensions{}
	}

	// Handle single column case (mobile)
	if rg.Columns == 1 {
		return rg.layoutVertical(gtx)
	}

	// Handle multi-column layout
	return rg.layoutGrid(gtx)
}

// layoutVertical stacks all children vertically
func (rg *ResponsiveGrid) layoutVertical(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(rg.Children)*2)

	for i, child := range rg.Children {
		child := child // Capture for closure
		children = append(children, layout.Rigid(child))

		// Add spacer between items
		if i < len(rg.Children)-1 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout))
		}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}

// layoutGrid arranges children in a grid
func (rg *ResponsiveGrid) layoutGrid(gtx layout.Context) layout.Dimensions {
	// Calculate rows needed
	rows := (len(rg.Children) + rg.Columns - 1) / rg.Columns

	rowChildren := make([]layout.FlexChild, 0, rows*2)

	for row := 0; row < rows; row++ {
		startIdx := row * rg.Columns
		endIdx := startIdx + rg.Columns
		if endIdx > len(rg.Children) {
			endIdx = len(rg.Children)
		}

		rowItems := rg.Children[startIdx:endIdx]

		// Create a row
		rowChildren = append(rowChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			colChildren := make([]layout.FlexChild, 0, len(rowItems)*2)

			for i, item := range rowItems {
				item := item // Capture for closure
				colChildren = append(colChildren, layout.Flexed(1, item))

				// Add spacer between columns
				if i < len(rowItems)-1 {
					colChildren = append(colChildren, layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout))
				}
			}

			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx, colChildren...)
		}))

		// Add spacer between rows
		if row < rows-1 {
			rowChildren = append(rowChildren, layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout))
		}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, rowChildren...)
}

// ResponsiveStack conditionally renders different layouts based on breakpoint
type ResponsiveStack struct {
	Mobile  layout.Widget
	Tablet  layout.Widget
	Desktop layout.Widget
}

// NewResponsiveStack creates a new responsive stack with breakpoint-specific layouts
func NewResponsiveStack(mobile, tablet, desktop layout.Widget) *ResponsiveStack {
	return &ResponsiveStack{
		Mobile:  mobile,
		Tablet:  tablet,
		Desktop: desktop,
	}
}

// Layout renders the appropriate widget based on current breakpoint
func (rs *ResponsiveStack) Layout(gtx layout.Context, rl *ResponsiveLayout) layout.Dimensions {
	bp := rl.GetBreakpoint(gtx)

	switch bp {
	case BreakpointMobile:
		if rs.Mobile != nil {
			return rs.Mobile(gtx)
		}
	case BreakpointTablet:
		if rs.Tablet != nil {
			return rs.Tablet(gtx)
		}
		// Fallback to mobile if tablet not defined
		if rs.Mobile != nil {
			return rs.Mobile(gtx)
		}
	case BreakpointDesktop, BreakpointWide:
		if rs.Desktop != nil {
			return rs.Desktop(gtx)
		}
		// Fallback to tablet if desktop not defined
		if rs.Tablet != nil {
			return rs.Tablet(gtx)
		}
	}

	// Ultimate fallback
	if rs.Desktop != nil {
		return rs.Desktop(gtx)
	}
	if rs.Tablet != nil {
		return rs.Tablet(gtx)
	}
	if rs.Mobile != nil {
		return rs.Mobile(gtx)
	}

	return layout.Dimensions{}
}
