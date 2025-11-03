package gui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// Tooltip represents a tooltip widget
type Tooltip struct {
	text      string
	visible   bool
	showTime  time.Time
	delayTime time.Duration
	hoverArea *widget.Clickable
}

// NewTooltip creates a new tooltip
func NewTooltip(text string) *Tooltip {
	return &Tooltip{
		text:      text,
		delayTime: 500 * time.Millisecond, // Show after 500ms hover
		hoverArea: new(widget.Clickable),
	}
}

// Layout renders the content with tooltip on hover
func (t *Tooltip) Layout(gtx layout.Context, theme *AppTheme, content layout.Widget) layout.Dimensions {
	// Check for hover state
	for t.hoverArea.Clicked(gtx) {
		// Consume clicks but don't do anything
	}

	// Track hover time
	if t.hoverArea.Hovered() {
		if !t.visible {
			if t.showTime.IsZero() {
				t.showTime = time.Now()
			} else if time.Since(t.showTime) >= t.delayTime {
				t.visible = true
			}
		}
	} else {
		t.visible = false
		t.showTime = time.Time{}
	}

	// Render with tooltip
	return layout.Stack{}.Layout(gtx,
		// Main content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return t.hoverArea.Layout(gtx, content)
		}),
		// Tooltip overlay (if visible)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if !t.visible || t.text == "" {
				return layout.Dimensions{}
			}

			// Position tooltip below the content
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return t.layoutTooltipBox(gtx, theme)
			})
		}),
	)
}

// layoutTooltipBox renders the tooltip box
func (t *Tooltip) layoutTooltipBox(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	// Tooltip background color (dark with transparency)
	bgColor := color.NRGBA{R: 50, G: 50, B: 50, A: 230}

	return layout.Inset{
		Top: unit.Dp(8), Bottom: unit.Dp(8),
		Left: unit.Dp(12), Right: unit.Dp(12),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Measure text
		label := theme.Caption(t.text)
		label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White text

		// Draw background
		macro := op.Record(gtx.Ops)
		dims := label.Layout(gtx)
		call := macro.Stop()

		// Add padding to dimensions
		paddedSize := image.Point{
			X: dims.Size.X + gtx.Dp(unit.Dp(16)),
			Y: dims.Size.Y + gtx.Dp(unit.Dp(8)),
		}

		// Draw rounded rectangle background
		rect := image.Rectangle{Max: paddedSize}
		rr := clip.UniformRRect(rect, gtx.Dp(unit.Dp(4)))
		paint.FillShape(gtx.Ops, bgColor, rr.Op(gtx.Ops))

		// Draw text on top
		layout.Inset{
			Top: unit.Dp(4), Bottom: unit.Dp(4),
			Left: unit.Dp(8), Right: unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return dims
		})

		return layout.Dimensions{Size: paddedSize}
	})
}

// WithDelay sets custom delay time before showing tooltip
func (t *Tooltip) WithDelay(delay time.Duration) *Tooltip {
	t.delayTime = delay
	return t
}

// Helper to create layout.Widget from a button with tooltip
func WithTooltip(tooltip *Tooltip, theme *AppTheme, button layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return tooltip.Layout(gtx, theme, button)
	}
}
