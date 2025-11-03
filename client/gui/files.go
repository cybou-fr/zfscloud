package gui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// FilesView displays backed up files and restore functionality
type FilesView struct {
	app           *App
	refreshButton widget.Clickable
	restoreButton widget.Clickable

	// File list (simplified for now)
	fileList widget.List
}

// NewFilesView creates a new files view
func NewFilesView(app *App) *FilesView {
	return &FilesView{
		app: app,
		fileList: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
	}
}

// Layout renders the files view
func (f *FilesView) Layout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title and actions
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return f.layoutHeader(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			// File browser/list
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return f.layoutFileList(gtx)
			}),
		)
	})
}

// layoutHeader renders the header with title and action buttons
func (f *FilesView) layoutHeader(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if f.refreshButton.Clicked(gtx) {
		f.refreshFileList()
	}
	if f.restoreButton.Clicked(gtx) {
		f.showRestoreDialog()
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.H5(f.app.theme, "Backup Files")
			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(f.app.theme, &f.refreshButton, "Refresh")
					return btn.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(f.app.theme, &f.restoreButton, "Restore...")
					btn.Background = f.app.theme.Palette.ContrastBg
					return btn.Layout(gtx)
				}),
			)
		}),
	)
}

// layoutFileList renders the list of backed up files
func (f *FilesView) layoutFileList(gtx layout.Context) layout.Dimensions {
	// TODO: Implement actual file list from server
	// For now, show placeholder message

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := material.H6(f.app.theme, "No files to display")
				return text.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := material.Body1(f.app.theme, "Start a backup to see files here")
				return text.Layout(gtx)
			}),
		)
	})
}

// refreshFileList fetches the latest file list from server
func (f *FilesView) refreshFileList() {
	// TODO: Implement file list refresh
	// 1. Call ListFiles RPC
	// 2. Update file list data
	// 3. Invalidate window to redraw
}

// showRestoreDialog shows the file restore dialog
func (f *FilesView) showRestoreDialog() {
	// TODO: Implement restore dialog
	// 1. Show file picker
	// 2. Select restore destination
	// 3. Call RestoreFile RPC
	// 4. Show progress
}
