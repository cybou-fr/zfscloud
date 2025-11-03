package gui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// DashboardView displays the main dashboard
type DashboardView struct {
	app           *App
	startButton   widget.Clickable
	stopButton    widget.Clickable
	pauseButton   widget.Clickable
	resumeButton  widget.Clickable
	refreshButton widget.Clickable
}

// NewDashboardView creates a new dashboard view
func NewDashboardView(app *App) *DashboardView {
	return &DashboardView{
		app: app,
	}
}

// Layout renders the dashboard
func (d *DashboardView) Layout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceBetween,
		}.Layout(gtx,
			// Connection status card
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutConnectionCard(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			// Engine controls
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutEngineControls(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			// Statistics cards
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutStatistics(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			// Current activity
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutCurrentActivity(gtx)
			}),
		)
	})
}

// layoutConnectionCard renders the connection status card with colored indicator
func (d *DashboardView) layoutConnectionCard(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()

	statusText := "Disconnected from server"
	indicatorColor := color.NRGBA{R: 220, G: 53, B: 69, A: 255} // Red
	if stats.Connected {
		statusText = "Connected to server"
		indicatorColor = color.NRGBA{R: 40, G: 167, B: 69, A: 255} // Green
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Status indicator circle
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Dp(unit.Dp(12))
				rect := image.Rectangle{Max: image.Point{X: size, Y: size}}
				circle := clip.UniformRRect(rect, size/2)
				paint.FillShape(gtx.Ops, indicatorColor, circle.Op(gtx.Ops))
				return layout.Dimensions{Size: image.Point{X: size, Y: size}}
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.H6(d.app.theme, statusText)
				return label.Layout(gtx)
			}),
		)
	})
}

// layoutEngineControls renders the backup engine control buttons
func (d *DashboardView) layoutEngineControls(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()

	// Handle button clicks
	if d.startButton.Clicked(gtx) {
		if d.app.engine != nil {
			// Start engine in background
			go func() {
				ctx := context.Background()
				if err := d.app.engine.Start(ctx); err != nil {
					log.Printf("Failed to start engine: %v", err)
				}
			}()
		}
	}
	if d.stopButton.Clicked(gtx) {
		if d.app.engine != nil {
			if err := d.app.engine.Stop(); err != nil {
				log.Printf("Failed to stop engine: %v", err)
			}
		}
	}
	if d.pauseButton.Clicked(gtx) {
		if d.app.engine != nil {
			d.app.engine.Pause()
		}
	}
	if d.resumeButton.Clicked(gtx) {
		if d.app.engine != nil {
			d.app.engine.Resume()
		}
	}
	if d.refreshButton.Clicked(gtx) {
		if d.app.engine != nil {
			// Perform initial scan
			go func() {
				ctx := context.Background()
				if err := d.app.engine.PerformInitialScan(ctx); err != nil {
					log.Printf("Failed to perform scan: %v", err)
				}
			}()
		}
	}

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(d.app.theme, &d.startButton, "Start Backup")
			btn.Background = d.app.theme.Palette.ContrastBg
			if stats.EngineRunning {
				// Disable if already running
				gtx = gtx.Disabled()
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(d.app.theme, &d.stopButton, "Stop")
			if !stats.EngineRunning {
				gtx = gtx.Disabled()
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			text := "Pause"
			if stats.EngineState == "paused" {
				text = "Resume"
			}
			btn := material.Button(d.app.theme, &d.pauseButton, text)
			if !stats.EngineRunning || stats.EngineState == "paused" {
				gtx = gtx.Disabled()
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(d.app.theme, &d.refreshButton, "Scan Now")
			if !stats.Connected {
				gtx = gtx.Disabled()
			}
			return btn.Layout(gtx)
		}),
	)
}

// layoutStatistics renders the statistics cards
func (d *DashboardView) layoutStatistics(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return d.statCard(gtx, "Files Scanned", fmt.Sprintf("%d", stats.FilesScanned))
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return d.statCard(gtx, "Files Uploaded", fmt.Sprintf("%d", stats.FilesUploaded))
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return d.statCard(gtx, "Files Failed", fmt.Sprintf("%d", stats.FilesFailed))
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			bytes := formatBytes(stats.BytesUploaded)
			return d.statCard(gtx, "Data Uploaded", bytes)
		}),
	)
}

// statCard renders a single statistics card with enhanced styling
func (d *DashboardView) statCard(gtx layout.Context, label, value string) layout.Dimensions {
	// Card background color (light gray)
	bgColor := color.NRGBA{R: 245, G: 245, B: 245, A: 255}

	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Draw rounded rectangle background
		rect := image.Rectangle{Max: gtx.Constraints.Min}
		rr := clip.UniformRRect(rect, gtx.Dp(unit.Dp(8)))
		paint.FillShape(gtx.Ops, bgColor, rr.Op(gtx.Ops))

		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(d.app.theme, label)
					lbl.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					val := material.H3(d.app.theme, value)
					val.Color = d.app.theme.Palette.ContrastBg
					val.Alignment = text.Middle
					return val.Layout(gtx)
				}),
			)
		})
	})
}

// layoutCurrentActivity renders the current backup activity
func (d *DashboardView) layoutCurrentActivity(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()

	var activityText string
	if stats.EngineRunning && stats.CurrentFile != "" {
		activityText = fmt.Sprintf("Currently backing up: %s", stats.CurrentFile)
	} else if stats.EngineRunning {
		activityText = "Backup engine running..."
	} else if !stats.LastBackup.IsZero() {
		activityText = fmt.Sprintf("Last backup: %s", stats.LastBackup.Format("2006-01-02 15:04:05"))
	} else {
		activityText = "No backup activity"
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H6(d.app.theme, "Current Activity")
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := material.Body1(d.app.theme, activityText)
				return text.Layout(gtx)
			}),
		)
	})
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
