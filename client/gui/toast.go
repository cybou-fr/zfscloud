package gui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// ToastType represents the type/severity of a toast notification
type ToastType int

const (
	ToastTypeInfo ToastType = iota
	ToastTypeSuccess
	ToastTypeWarning
	ToastTypeError
)

// Toast represents a temporary notification message
type Toast struct {
	Type      ToastType
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
	Visible   bool

	dismissBtn widget.Clickable
}

// NewToast creates a new toast notification
func NewToast(toastType ToastType, message string, duration time.Duration) *Toast {
	return &Toast{
		Type:      toastType,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  duration,
		Visible:   true,
	}
}

// IsExpired returns true if the toast has exceeded its display duration
func (t *Toast) IsExpired() bool {
	return time.Since(t.CreatedAt) > t.Duration
}

// Dismiss hides the toast
func (t *Toast) Dismiss() {
	t.Visible = false
}

// Layout renders the toast notification
func (t *Toast) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	if !t.Visible {
		return layout.Dimensions{}
	}

	// Auto-dismiss after duration
	if t.IsExpired() {
		t.Dismiss()
		return layout.Dimensions{}
	}

	// Handle dismiss button click
	if t.dismissBtn.Clicked(gtx) {
		t.Dismiss()
		return layout.Dimensions{}
	}

	// Choose color based on toast type
	var bgColor, textColor color.NRGBA
	switch t.Type {
	case ToastTypeSuccess:
		bgColor = theme.Success
		textColor = theme.OnPrimary
	case ToastTypeWarning:
		bgColor = theme.Warning
		textColor = theme.OnPrimary
	case ToastTypeError:
		bgColor = theme.Error
		textColor = theme.OnPrimary
	default: // ToastTypeInfo
		bgColor = theme.Info
		textColor = theme.OnPrimary
	}

	// Draw toast with rounded corners and shadow
	return layout.Inset{
		Top:    theme.SpaceMedium(),
		Bottom: theme.SpaceSmall(),
		Left:   theme.SpaceLarge(),
		Right:  theme.SpaceLarge(),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Limit width
		gtx.Constraints.Max.X = gtx.Dp(unit.Dp(400))

		return widget.Border{
			Color:        theme.TextSecondary,
			Width:        unit.Dp(1),
			CornerRadius: unit.Dp(4),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Background
			defer clip.UniformRRect(image.Rectangle{
				Max: gtx.Constraints.Min,
			}, gtx.Dp(unit.Dp(4))).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, bgColor)

			// Content
			return layout.UniformInset(theme.SpaceSmall()).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceBetween,
				}.Layout(gtx,
					// Message
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						label := material.Label(theme.Theme, theme.TextSize, t.Message)
						label.Color = theme.OnPrimary
						return label.Layout(gtx)
					}),
					// Dismiss button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Left: theme.SpaceSmall()}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(theme.Theme, &t.dismissBtn, "✕")
							btn.Background = bgColor
							btn.Color = textColor
							btn.Inset = layout.Inset{
								Top: theme.SpaceTiny(), Bottom: theme.SpaceTiny(),
								Left: theme.SpaceSmall(), Right: theme.SpaceSmall(),
							}
							return btn.Layout(gtx)
						})
					}),
				)
			})
		})
	})
}

// ToastManager manages multiple toast notifications
type ToastManager struct {
	toasts    []*Toast
	maxToasts int
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:    make([]*Toast, 0),
		maxToasts: 3, // Show max 3 toasts at once
	}
}

// Show adds a new toast notification
func (tm *ToastManager) Show(toastType ToastType, message string, duration time.Duration) {
	// Remove expired toasts
	tm.cleanup()

	// Limit number of visible toasts
	if len(tm.toasts) >= tm.maxToasts {
		// Remove oldest toast
		tm.toasts = tm.toasts[1:]
	}

	// Add new toast
	toast := NewToast(toastType, message, duration)
	tm.toasts = append(tm.toasts, toast)
}

// ShowInfo shows an info toast
func (tm *ToastManager) ShowInfo(message string) {
	tm.Show(ToastTypeInfo, message, 3*time.Second)
}

// ShowSuccess shows a success toast
func (tm *ToastManager) ShowSuccess(message string) {
	tm.Show(ToastTypeSuccess, message, 3*time.Second)
}

// ShowWarning shows a warning toast
func (tm *ToastManager) ShowWarning(message string) {
	tm.Show(ToastTypeWarning, message, 4*time.Second)
}

// ShowError shows an error toast
func (tm *ToastManager) ShowError(message string) {
	tm.Show(ToastTypeError, message, 5*time.Second)
}

// cleanup removes expired and dismissed toasts
func (tm *ToastManager) cleanup() {
	filtered := make([]*Toast, 0, len(tm.toasts))
	for _, toast := range tm.toasts {
		if toast.Visible && !toast.IsExpired() {
			filtered = append(filtered, toast)
		}
	}
	tm.toasts = filtered
}

// Layout renders all active toasts
func (tm *ToastManager) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	// Cleanup expired toasts
	tm.cleanup()

	if len(tm.toasts) == 0 {
		return layout.Dimensions{}
	}

	// Stack toasts vertically from top
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					// Render all toasts
					dims := layout.Dimensions{}
					for _, toast := range tm.toasts {
						dims = toast.Layout(gtx, theme)
						if dims.Size.Y > 0 {
							// Add spacing between toasts
							gtx.Constraints.Min.Y -= dims.Size.Y + gtx.Dp(theme.SpaceSmall())
						}
					}
					return dims
				}),
			)
		}),
	)
}

// HasActive returns true if there are any active toasts
func (tm *ToastManager) HasActive() bool {
	tm.cleanup()
	return len(tm.toasts) > 0
}
