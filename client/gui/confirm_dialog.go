package gui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// ConfirmDialog represents a confirmation dialog with Yes/No buttons
type ConfirmDialog struct {
	Title         string
	Message       string
	ConfirmText   string // Default: "Yes"
	CancelText    string // Default: "No"
	Visible       bool
	confirmButton widget.Clickable
	cancelButton  widget.Clickable
	onConfirm     func()
	onCancel      func()
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		Title:       title,
		Message:     message,
		ConfirmText: "Yes",
		CancelText:  "No",
		Visible:     false,
	}
}

// Show displays the confirmation dialog
func (cd *ConfirmDialog) Show(onConfirm, onCancel func()) {
	cd.Visible = true
	cd.onConfirm = onConfirm
	cd.onCancel = onCancel
}

// Hide closes the confirmation dialog
func (cd *ConfirmDialog) Hide() {
	cd.Visible = false
	cd.onConfirm = nil
	cd.onCancel = nil
}

// Layout renders the confirmation dialog if visible
func (cd *ConfirmDialog) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	if !cd.Visible {
		return layout.Dimensions{}
	}

	// Handle button clicks
	if cd.confirmButton.Clicked(gtx) {
		cd.Visible = false
		if cd.onConfirm != nil {
			cd.onConfirm()
		}
	}
	if cd.cancelButton.Clicked(gtx) {
		cd.Visible = false
		if cd.onCancel != nil {
			cd.onCancel()
		}
	}

	// Render dialog overlay
	return layout.Stack{}.Layout(gtx,
		// Semi-transparent background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Draw dimmed background
			dims := layout.Dimensions{Size: gtx.Constraints.Max}
			// Note: In production, add a semi-transparent rect here
			return dims
		}),
		// Dialog box
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Dialog constraints
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(450))
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(400))

				return cd.layoutDialog(gtx, theme)
			})
		}),
	)
}

// layoutDialog renders the dialog content
func (cd *ConfirmDialog) layoutDialog(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	// Dialog dimensions
	dialogWidth := gtx.Dp(unit.Dp(400))
	dialogHeight := gtx.Dp(unit.Dp(200))
	dialogSize := image.Point{X: dialogWidth, Y: dialogHeight}

	// Draw background
	paint.FillShape(gtx.Ops, theme.Surface, clip.UniformRRect(image.Rectangle{Max: dialogSize}, gtx.Dp(unit.Dp(12))).Op(gtx.Ops))

	return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := theme.H6(cd.Title)
				title.Color = theme.Error // Red for warning dialogs
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			// Message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				msg := theme.Body1(cd.Message)
				return msg.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					// Cancel button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(theme.Theme, &cd.cancelButton, cd.CancelText)
						btn.Background = theme.Surface
						btn.Color = theme.OnSurface
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					// Confirm button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(theme.Theme, &cd.confirmButton, cd.ConfirmText)
						btn.Background = theme.Error
						btn.Color = theme.OnPrimary
						return btn.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// WithText sets custom button text
func (cd *ConfirmDialog) WithText(confirm, cancel string) *ConfirmDialog {
	cd.ConfirmText = confirm
	cd.CancelText = cancel
	return cd
}
