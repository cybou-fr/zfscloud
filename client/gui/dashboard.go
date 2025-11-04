package gui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

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

	// Tooltips for buttons
	tooltipStart   *Tooltip
	tooltipStop    *Tooltip
	tooltipPause   *Tooltip
	tooltipRefresh *Tooltip

	// Dialogs
	stopConfirmDialog *ConfirmDialog

	// Progress tracking
	scanProgress      *ProgressBar
	scanProgressLabel string
	showScanProgress  bool
	scanInProgress    bool
}

// NewDashboardView creates a new dashboard view
func NewDashboardView(app *App) *DashboardView {
	return &DashboardView{
		app:               app,
		tooltipStart:      NewTooltip("Start the backup engine to begin syncing files"),
		tooltipStop:       NewTooltip("Stop the backup engine and close all connections"),
		tooltipPause:      NewTooltip("Pause or resume the backup process"),
		tooltipRefresh:    NewTooltip("Scan directories for new or changed files"),
		stopConfirmDialog: NewConfirmDialog("Stop Backup Engine?", "This will stop all backup operations in progress. Are you sure?"),
		scanProgress:      &ProgressBar{Height: unit.Dp(8)},
	}
}

// Layout renders the dashboard
func (d *DashboardView) Layout(gtx layout.Context) layout.Dimensions {
	// Show confirmation dialog if active
	if d.stopConfirmDialog.Visible {
		return layout.Stack{}.Layout(gtx,
			// Main dashboard (dimmed)
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return d.layoutMain(gtx)
			}),
			// Confirmation dialog overlay
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return d.stopConfirmDialog.Layout(gtx, d.app.GetTheme())
			}),
		)
	}

	return d.layoutMain(gtx)
}

// layoutMain renders the main dashboard content
func (d *DashboardView) layoutMain(gtx layout.Context) layout.Dimensions {
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
	theme := d.app.GetTheme()

	statusText := "Disconnected from server"
	statusIcon := IconSymbols[IconDisconnected]
	indicatorColor := theme.StatusDisconnected
	if stats.Connected {
		statusText = "Connected to server"
		statusIcon = IconSymbols[IconConnected]
		indicatorColor = theme.StatusConnected
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
				label := theme.H6(statusIcon + " " + statusText)
				return label.Layout(gtx)
			}),
		)
	})
}

// layoutEngineControls renders the backup engine control buttons
func (d *DashboardView) layoutEngineControls(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()

	// Handle button clicks with proper nil checking and context timeouts
	if d.startButton.Clicked(gtx) {
		engine := d.app.engine
		if engine == nil {
			d.app.ShowToast(ToastTypeError, "Backup engine not initialized")
			return layout.Dimensions{}
		}

		d.app.ShowToast(ToastTypeInfo, "Starting backup engine...")
		// Start engine in background with timeout
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := engine.Start(ctx); err != nil {
				log.Printf("Failed to start engine: %v", err)
				d.app.ShowToast(ToastTypeError, "Failed to start backup engine. Please check connection.")
			} else {
				d.app.ShowToast(ToastTypeSuccess, "Backup engine started successfully")
			}
		}()
	}

	if d.stopButton.Clicked(gtx) {
		engine := d.app.engine
		if engine == nil {
			d.app.ShowToast(ToastTypeError, "Backup engine not initialized")
			return layout.Dimensions{}
		}

		// Show confirmation dialog
		d.stopConfirmDialog.Show(
			func() { // onConfirm - user clicked "Yes"
				if err := engine.Stop(); err != nil {
					log.Printf("Failed to stop engine: %v", err)
					d.app.ShowToast(ToastTypeError, "Failed to stop backup engine")
				} else {
					d.app.ShowToast(ToastTypeInfo, "Backup engine stopped")
				}
			},
			func() { // onCancel - user clicked "No"
				d.app.ShowToast(ToastTypeInfo, "Stop operation cancelled")
			},
		)
	}

	if d.pauseButton.Clicked(gtx) {
		engine := d.app.engine
		if engine == nil {
			d.app.ShowToast(ToastTypeError, "Backup engine not initialized")
			return layout.Dimensions{}
		}

		engine.Pause()
		d.app.ShowToast(ToastTypeWarning, "Backup paused")
	}

	if d.resumeButton.Clicked(gtx) {
		engine := d.app.engine
		if engine == nil {
			d.app.ShowToast(ToastTypeError, "Backup engine not initialized")
			return layout.Dimensions{}
		}

		engine.Resume()
		d.app.ShowToast(ToastTypeSuccess, "Backup resumed")
	}

	if d.refreshButton.Clicked(gtx) {
		// Prevent multiple simultaneous scans
		if d.scanInProgress {
			d.app.ShowToast(ToastTypeWarning, "Scan already in progress")
			return layout.Dimensions{}
		}

		engine := d.app.engine
		if engine == nil {
			d.app.ShowToast(ToastTypeError, "Backup engine not initialized")
			return layout.Dimensions{}
		}

		// Start scan with progress tracking
		d.scanInProgress = true
		d.showScanProgress = true
		d.scanProgress.Progress = 0.0
		d.scanProgressLabel = "Starting scan..."
		d.app.ShowToast(ToastTypeInfo, "Scanning files...")

		// Perform initial scan with timeout
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			defer func() {
				d.scanInProgress = false
				d.showScanProgress = false
			}()

			// Simulate progress updates (in real implementation, engine would report progress)
			progressSteps := []struct {
				progress float32
				label    string
			}{
				{0.2, "Scanning directories..."},
				{0.5, "Analyzing files..."},
				{0.8, "Calculating checksums..."},
			}

			for i, step := range progressSteps {
				select {
				case <-ctx.Done():
					d.app.ShowToast(ToastTypeError, "Scan cancelled")
					return
				case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
					d.scanProgress.Progress = step.progress
					d.scanProgressLabel = step.label
					// UI will update on next frame
				}
			}

			// Perform actual scan
			d.scanProgress.Progress = 0.9
			d.scanProgressLabel = "Finalizing..."
			if err := engine.PerformInitialScan(ctx); err != nil {
				log.Printf("Failed to perform scan: %v", err)
				d.app.ShowToast(ToastTypeError, "File scan failed. Please check your configuration.")
			} else {
				d.scanProgress.Progress = 1.0
				d.scanProgressLabel = "Scan complete!"
				time.Sleep(500 * time.Millisecond) // Show completion briefly
				d.app.ShowToast(ToastTypeSuccess, "File scan completed successfully")
			}
		}()
	}

	theme := d.app.GetTheme()

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.tooltipStart.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(theme.Theme, &d.startButton, IconSymbols[IconStart]+" Start Backup")
				btn.Background = theme.Success
				btn.Color = theme.OnPrimary
				if stats.EngineRunning {
					// Disable if already running
					gtx = gtx.Disabled()
				}
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.tooltipStop.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(theme.Theme, &d.stopButton, IconSymbols[IconStop]+" Stop")
				btn.Background = theme.Error
				btn.Color = theme.OnPrimary
				if !stats.EngineRunning {
					gtx = gtx.Disabled()
				}
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.tooltipPause.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				text := IconSymbols[IconPause] + " Pause"
				// Change button logic: enable Resume when paused, enable Pause when running
				if stats.EngineState == "paused" {
					text = IconSymbols[IconResume] + " Resume"
					// Enable Resume button when paused
				} else if !stats.EngineRunning {
					// Disable button when engine not running
					gtx = gtx.Disabled()
				}
				btn := material.Button(theme.Theme, &d.pauseButton, text)
				btn.Background = theme.Warning
				btn.Color = theme.OnPrimary
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.tooltipRefresh.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				// Show loading state if scan is in progress
				if d.scanInProgress {
					return d.layoutButtonWithSpinner(gtx, "Scanning...", theme.Info)
				}
				btn := material.Button(theme.Theme, &d.refreshButton, IconSymbols[IconRefresh]+" Scan Now")
				btn.Background = theme.Info
				btn.Color = theme.OnPrimary
				if !stats.Connected {
					gtx = gtx.Disabled()
				}
				return btn.Layout(gtx)
			})
		}),
	)
}

// layoutStatistics renders the statistics cards with enhanced styling and responsive layout
func (d *DashboardView) layoutStatistics(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()
	theme := d.app.GetTheme()

	// Create enhanced stat cards with icons and colors
	cards := []*StatCard{
		NewStatCard("Files Scanned", fmt.Sprintf("%d", stats.FilesScanned), IconSearch, theme).
			WithColor(theme.Info),
		NewStatCard("Files Uploaded", fmt.Sprintf("%d", stats.FilesUploaded), IconUpload, theme).
			WithColor(theme.Success),
		NewStatCard("Files Failed", fmt.Sprintf("%d", stats.FilesFailed), IconError, theme).
			WithColor(theme.Error),
		NewStatCard("Data Uploaded", formatBytes(stats.BytesUploaded), IconFolder, theme).
			WithColor(theme.Primary),
	}

	// Get responsive columns based on screen size
	columns := d.app.responsiveLayout.GetCardColumns(gtx)

	// Create grid and render with responsive layout
	grid := NewStatCardGrid(cards...)
	return grid.LayoutResponsive(gtx, theme, columns)
}

// statCard renders a single statistics card with enhanced styling
func (d *DashboardView) statCard(gtx layout.Context, label, value string) layout.Dimensions {
	theme := d.app.GetTheme()

	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Draw rounded rectangle background
		rect := image.Rectangle{Max: gtx.Constraints.Min}
		rr := clip.UniformRRect(rect, gtx.Dp(unit.Dp(8)))
		paint.FillShape(gtx.Ops, theme.Surface, rr.Op(gtx.Ops))

		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := theme.Body2(label)
					lbl.Color = theme.TextSecondary
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					val := theme.H3(value)
					val.Color = theme.Primary
					val.Alignment = text.Middle
					return val.Layout(gtx)
				}),
			)
		})
	})
}

// layoutCurrentActivity renders the current backup activity with helpful empty states
func (d *DashboardView) layoutCurrentActivity(gtx layout.Context) layout.Dimensions {
	stats := d.app.GetStats()
	theme := d.app.GetTheme()

	var activityText string
	var isEmptyState bool

	if stats.EngineRunning && stats.CurrentFile != "" {
		activityText = fmt.Sprintf("📄 Currently backing up: %s", stats.CurrentFile)
	} else if stats.EngineRunning {
		activityText = "⚙️ Backup engine is running and monitoring for changes..."
	} else if !stats.LastBackup.IsZero() {
		activityText = fmt.Sprintf("✓ Last backup: %s", stats.LastBackup.Format("2006-01-02 15:04:05"))
	} else if !stats.Connected {
		activityText = "⚠️ Not connected to server. Please check Settings and test your connection."
		isEmptyState = true
	} else if stats.FilesScanned == 0 {
		activityText = "👋 Welcome! Click 'Scan Now' to begin scanning your files for backup."
		isEmptyState = true
	} else {
		activityText = "Ready to backup. Click 'Start Backup' to begin."
		isEmptyState = true
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := theme.H6("Current Activity")
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := theme.Body1(activityText)
				if isEmptyState {
					text.Color = theme.TextSecondary
				}
				return text.Layout(gtx)
			}),
		}

		// Show progress bar if scan is in progress
		if d.showScanProgress {
			children = append(children,
				layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						// Progress label
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if d.scanProgressLabel != "" {
								label := theme.Caption(d.scanProgressLabel)
								label.Color = theme.TextSecondary
								return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, label.Layout)
							}
							return layout.Dimensions{}
						}),
						// Progress bar
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return d.scanProgress.Layout(gtx, theme)
						}),
					)
				}),
			)
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx, children...)
	})
}

// layoutButtonWithSpinner creates a button with a loading spinner
func (d *DashboardView) layoutButtonWithSpinner(gtx layout.Context, text string, color color.NRGBA) layout.Dimensions {
	theme := d.app.GetTheme()

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Circular progress indicator
			return widget.Border{
				Color:        color,
				CornerRadius: unit.Dp(4),
				Width:        unit.Dp(1),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				size := gtx.Dp(unit.Dp(20))
				return layout.Dimensions{Size: image.Pt(size, size)}
			})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(theme.Theme, unit.Sp(14), text)
			label.Color = color
			return label.Layout(gtx)
		}),
	)
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
