package gui

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// ProgressBar renders a determinate progress bar
type ProgressBar struct {
	Progress float32 // 0.0 to 1.0
	Color    color.NRGBA
	Height   unit.Dp
}

// Layout renders the progress bar
func (p *ProgressBar) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	if p.Color == (color.NRGBA{}) {
		p.Color = theme.Primary
	}
	if p.Height == 0 {
		p.Height = unit.Dp(4)
	}

	// Clamp progress between 0 and 1
	progress := p.Progress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	return layout.Stack{}.Layout(gtx,
		// Background track
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			height := gtx.Dp(p.Height)
			size := image.Point{X: gtx.Constraints.Max.X, Y: height}

			// Draw background
			rect := clip.Rect{Max: size}
			paint.FillShape(gtx.Ops, theme.Disabled, rect.Op())

			return layout.Dimensions{Size: size}
		}),
		// Progress fill
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			height := gtx.Dp(p.Height)
			width := int(float32(gtx.Constraints.Max.X) * progress)
			size := image.Point{X: width, Y: height}

			// Draw progress
			rect := clip.Rect{Max: size}
			paint.FillShape(gtx.Ops, p.Color, rect.Op())

			return layout.Dimensions{Size: size}
		}),
	)
}

// CircularProgress renders an indeterminate circular progress indicator
type CircularProgress struct {
	startAngle float32
	lastUpdate time.Time
	Color      color.NRGBA
	Size       unit.Dp
}

// NewCircularProgress creates a new circular progress indicator
func NewCircularProgress() *CircularProgress {
	return &CircularProgress{
		lastUpdate: time.Now(),
	}
}

// Layout renders the circular progress indicator
func (c *CircularProgress) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	if c.Color == (color.NRGBA{}) {
		c.Color = theme.Primary
	}
	if c.Size == 0 {
		c.Size = unit.Dp(24)
	}

	// Update rotation
	now := time.Now()
	if !c.lastUpdate.IsZero() {
		elapsed := now.Sub(c.lastUpdate).Seconds()
		// Rotate 360 degrees per 1 second
		c.startAngle += float32(elapsed * 360)
		if c.startAngle >= 360 {
			c.startAngle -= 360
		}
	}
	c.lastUpdate = now

	// Request redraw for continuous animation
	gtx.Execute(op.InvalidateCmd{})

	size := gtx.Dp(c.Size)
	gtx.Constraints = layout.Exact(image.Point{X: size, Y: size})

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			// Draw circular arc
			defer op.Offset(image.Point{}).Push(gtx.Ops).Pop()

			// Calculate arc parameters
			radius := float32(size) / 2
			center := f32.Point{X: radius, Y: radius}
			strokeWidth := float32(size) / 8

			// Create arc path (270 degrees)
			const arcLength = 270.0
			startAngle := c.startAngle
			endAngle := startAngle + arcLength

			path := clip.Path{}
			path.Begin(gtx.Ops)

			// Draw arc
			steps := 50
			for i := 0; i <= steps; i++ {
				t := float32(i) / float32(steps)
				angle := startAngle + (endAngle-startAngle)*t
				rad := angle * math.Pi / 180
				x := center.X + radius*float32(math.Cos(float64(rad)))
				y := center.Y + radius*float32(math.Sin(float64(rad)))

				if i == 0 {
					path.MoveTo(f32.Point{X: x, Y: y})
				} else {
					path.LineTo(f32.Point{X: x, Y: y})
				}
			}

			// Stroke the path
			paint.FillShape(gtx.Ops, c.Color, clip.Stroke{
				Path:  path.End(),
				Width: strokeWidth,
			}.Op())

			return layout.Dimensions{Size: image.Point{X: size, Y: size}}
		}),
	)
}

// LoadingOverlay renders a loading overlay with spinner and text
type LoadingOverlay struct {
	spinner *CircularProgress
	Message string
}

// NewLoadingOverlay creates a new loading overlay
func NewLoadingOverlay(message string) *LoadingOverlay {
	return &LoadingOverlay{
		spinner: NewCircularProgress(),
		Message: message,
	}
}

// Layout renders the loading overlay
func (l *LoadingOverlay) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		// Semi-transparent background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := gtx.Constraints.Max
			paint.FillShape(gtx.Ops, theme.Overlay, clip.Rect{Max: size}.Op())
			return layout.Dimensions{Size: size}
		}),
		// Centered content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Vertical,
					Alignment: layout.Middle,
				}.Layout(gtx,
					// Spinner
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						l.spinner.Size = unit.Dp(48)
						return l.spinner.Layout(gtx, theme)
					}),
					layout.Rigid(layout.Spacer{Height: theme.SpaceMedium()}.Layout),
					// Message
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if l.Message != "" {
							label := theme.Body1(l.Message)
							label.Color = theme.Surface
							return label.Layout(gtx)
						}
						return layout.Dimensions{}
					}),
				)
			})
		}),
	)
}

// ButtonWithLoading wraps a button with a loading state
type ButtonWithLoading struct {
	Button  *widget.Clickable
	Loading bool
	spinner *CircularProgress
}

// NewButtonWithLoading creates a new button with loading state
func NewButtonWithLoading() *ButtonWithLoading {
	return &ButtonWithLoading{
		Button:  &widget.Clickable{},
		spinner: NewCircularProgress(),
	}
}

// Layout renders the button with optional loading state
func (b *ButtonWithLoading) Layout(gtx layout.Context, theme *AppTheme, label string) layout.Dimensions {
	btn := material.Button(theme.Theme, b.Button, label)

	if b.Loading {
		// Disable button and show spinner
		gtx = gtx.Disabled()

		return btn.Layout(gtx)
	}

	return btn.Layout(gtx)
}

// SkeletonLoader renders a skeleton loading placeholder
type SkeletonLoader struct {
	Width     unit.Dp
	Height    unit.Dp
	animated  bool
	startTime time.Time
}

// NewSkeletonLoader creates a new skeleton loader
func NewSkeletonLoader(width, height unit.Dp) *SkeletonLoader {
	return &SkeletonLoader{
		Width:     width,
		Height:    height,
		animated:  true,
		startTime: time.Now(),
	}
}

// Layout renders the skeleton loader
func (s *SkeletonLoader) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	width := gtx.Dp(s.Width)
	height := gtx.Dp(s.Height)
	size := image.Point{X: width, Y: height}

	// Calculate shimmer effect
	var alpha uint8 = 255
	if s.animated {
		elapsed := time.Since(s.startTime).Seconds()
		// Pulse between 0.3 and 1.0
		pulse := 0.3 + 0.7*float32(math.Sin(elapsed*2))*0.5
		alpha = uint8(pulse * 255)

		// Request redraw for continuous animation
		gtx.Execute(op.InvalidateCmd{})
	}

	baseColor := theme.Disabled
	shimmerColor := color.NRGBA{
		R: baseColor.R,
		G: baseColor.G,
		B: baseColor.B,
		A: alpha,
	}

	// Draw rounded rectangle
	rect := clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(unit.Dp(4)))
	paint.FillShape(gtx.Ops, shimmerColor, rect.Op(gtx.Ops))

	return layout.Dimensions{Size: size}
}
