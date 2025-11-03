package gui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
)

// StatCard represents an enhanced statistics card
type StatCard struct {
	Label      string
	Value      string
	Icon       IconType
	Trend      TrendType
	TrendValue string // e.g., "+12%", "-5%"
	Color      color.NRGBA
}

// TrendType represents the trend direction
type TrendType int

const (
	TrendNone TrendType = iota
	TrendUp
	TrendDown
	TrendFlat
)

// TrendSymbols maps trend types to Unicode symbols
var TrendSymbols = map[TrendType]string{
	TrendNone: "",
	TrendUp:   "↑",
	TrendDown: "↓",
	TrendFlat: "→",
}

// NewStatCard creates a new statistics card
func NewStatCard(label, value string, icon IconType, theme *AppTheme) *StatCard {
	return &StatCard{
		Label: label,
		Value: value,
		Icon:  icon,
		Trend: TrendNone,
		Color: theme.Primary,
	}
}

// WithTrend adds trend information to the stat card
func (sc *StatCard) WithTrend(trend TrendType, value string) *StatCard {
	sc.Trend = trend
	sc.TrendValue = value
	return sc
}

// WithColor sets a custom color for the stat card
func (sc *StatCard) WithColor(c color.NRGBA) *StatCard {
	sc.Color = c
	return sc
}

// Layout renders the enhanced stat card
func (sc *StatCard) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Calculate gradient colors
		bgStart := sc.Color
		bgEnd := sc.adjustColorBrightness(sc.Color, 0.85) // Slightly darker

		// Draw card with gradient background
		return sc.drawGradientCard(gtx, bgStart, bgEnd, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					// Icon + Label row
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis:      layout.Horizontal,
							Alignment: layout.Middle,
						}.Layout(gtx,
							// Icon
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if sc.Icon == 0 {
									return layout.Dimensions{}
								}
								icon := NewIcon(sc.Icon, theme).WithColor(theme.OnPrimary).WithSize(unit.Sp(18))
								return icon.Layout(gtx, theme)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							// Label
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := theme.Body2(sc.Label)
								lbl.Color = theme.OnPrimary
								lbl.Alignment = text.Start
								return lbl.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					// Value
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						val := theme.H3(sc.Value)
						val.Color = theme.OnPrimary
						val.Alignment = text.Start
						return val.Layout(gtx)
					}),
					// Trend indicator (if present)
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if sc.Trend == TrendNone || sc.TrendValue == "" {
							return layout.Dimensions{}
						}
						return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							trendText := TrendSymbols[sc.Trend] + " " + sc.TrendValue
							trendColor := theme.OnPrimary

							// Adjust opacity for trend
							trendColor.A = 200 // Slightly transparent

							trend := theme.Caption(trendText)
							trend.Color = trendColor
							trend.Alignment = text.Start
							return trend.Layout(gtx)
						})
					}),
				)
			})
		})
	})
}

// drawGradientCard draws a card with gradient background
func (sc *StatCard) drawGradientCard(gtx layout.Context, colorStart, colorEnd color.NRGBA, content layout.Widget) layout.Dimensions {
	// Draw rounded rectangle background with solid color
	rect := image.Rectangle{Max: gtx.Constraints.Min}
	rr := clip.UniformRRect(rect, gtx.Dp(unit.Dp(12)))
	paint.FillShape(gtx.Ops, colorStart, rr.Op(gtx.Ops))

	// Render content on top
	return content(gtx)
} // adjustColorBrightness adjusts the brightness of a color by a factor
func (sc *StatCard) adjustColorBrightness(c color.NRGBA, factor float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(float32(c.R) * factor),
		G: uint8(float32(c.G) * factor),
		B: uint8(float32(c.B) * factor),
		A: c.A,
	}
}

// StatCardGrid renders a grid of stat cards
type StatCardGrid struct {
	Cards []*StatCard
}

// NewStatCardGrid creates a new stat card grid
func NewStatCardGrid(cards ...*StatCard) *StatCardGrid {
	return &StatCardGrid{
		Cards: cards,
	}
}

// Layout renders the stat card grid (non-responsive)
func (scg *StatCardGrid) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	if len(scg.Cards) == 0 {
		return layout.Dimensions{}
	}

	// Create flexible layout based on number of cards
	children := make([]layout.FlexChild, 0, len(scg.Cards)*2)

	for i, card := range scg.Cards {
		card := card // Capture for closure
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return card.Layout(gtx, theme)
		}))

		// Add spacer between cards (but not after last card)
		if i < len(scg.Cards)-1 {
			children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout))
		}
	}

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx, children...)
}

// LayoutResponsive renders the stat card grid with responsive columns
func (scg *StatCardGrid) LayoutResponsive(gtx layout.Context, theme *AppTheme, columns int) layout.Dimensions {
	if len(scg.Cards) == 0 {
		return layout.Dimensions{}
	}

	// Convert cards to widgets
	widgets := make([]layout.Widget, len(scg.Cards))
	for i, card := range scg.Cards {
		card := card // Capture for closure
		widgets[i] = func(gtx layout.Context) layout.Dimensions {
			return card.Layout(gtx, theme)
		}
	}

	// Use responsive grid
	grid := NewResponsiveGrid(columns, widgets...)
	return grid.Layout(gtx)
} // Helper function to format large numbers
func FormatNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
