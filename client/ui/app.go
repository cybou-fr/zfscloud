package ui

import (
	"context"
	"fmt"
	"image"
	"image/color"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	"github.com/ioustamora/zfscloud/shared/config"
)

// App represents the main GUI application
type App struct {
	window *app.Window
	engine *backup.Engine
	client *client.Client
	config *config.ClientConfig
	theme  *material.Theme

	// UI state
	stats        backup.BackupStats
	notification string

	// Widgets
	startBtn  widget.Clickable
	stopBtn   widget.Clickable
	pauseBtn  widget.Clickable
	resumeBtn widget.Clickable
	scanBtn   widget.Clickable
	exitBtn   widget.Clickable
}

// NewApp creates a new GUI application instance
func NewApp(ctx context.Context, c *client.Client, cfg *config.ClientConfig) (*App, error) {
	w := app.NewWindow(
		app.Title("ZFSCloud Backup Client"),
		app.Size(unit.Dp(800), unit.Dp(600)),
	)

	a := &App{
		window: w,
		client: c,
		config: cfg,
		theme:  material.NewTheme(),
	}

	return a, nil
}

// Run starts the GUI application event loop
func (a *App) Run(ctx context.Context) error {
	var ops op.Ops

	for e := range a.window.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err

		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			// Handle button clicks
			for a.exitBtn.Clicked() {
				a.window.Perform(system.ActionClose)
			}

			for a.startBtn.Clicked() {
				if err := a.startBackup(ctx); err != nil {
					a.notification = fmt.Sprintf("Error: %v", err)
				}
			}

			for a.stopBtn.Clicked() {
				if err := a.stopBackup(); err != nil {
					a.notification = fmt.Sprintf("Error: %v", err)
				}
			}

			for a.pauseBtn.Clicked() {
				if err := a.pauseBackup(); err != nil {
					a.notification = fmt.Sprintf("Error: %v", err)
				}
			}

			for a.resumeBtn.Clicked() {
				if err := a.resumeBackup(); err != nil {
					a.notification = fmt.Sprintf("Error: %v", err)
				}
			}

			for a.scanBtn.Clicked() {
				if err := a.performScan(ctx); err != nil {
					a.notification = fmt.Sprintf("Error: %v", err)
				}
			}

			// Update stats if engine is running
			if a.engine != nil {
				a.stats = a.engine.GetStats()
			}

			// Draw the UI
			a.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}

	return nil
}

// layout draws the main UI
func (a *App) layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.drawHeader(gtx)
		}),
		// Main content
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.drawDashboard(gtx)
		}),
	)
}

// drawHeader draws the header with status
func (a *App) drawHeader(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Spacing:   layout.SpaceBetween,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H5(a.theme, "ZFSCloud Backup")
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.drawStatusIndicator(gtx)
			}),
		)
	})
}

// drawStatusIndicator shows the current backup state
func (a *App) drawStatusIndicator(gtx layout.Context) layout.Dimensions {
	state := "Idle"
	statusColor := color.NRGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF} // Gray

	if a.engine != nil {
		state = string(a.stats.CurrentState)
		switch a.stats.CurrentState {
		case "backing_up", "scanning":
			statusColor = color.NRGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // Green
		case "error":
			statusColor = color.NRGBA{R: 0xF4, G: 0x43, B: 0x36, A: 0xFF} // Red
		case "paused":
			statusColor = color.NRGBA{R: 0xFF, G: 0x98, B: 0x00, A: 0xFF} // Orange
		}
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Status dot
			size := gtx.Dp(unit.Dp(12))
			rect := image.Rect(0, 0, size, size)
			defer clip.Ellipse(rect).Push(gtx.Ops).Pop()
			paint.ColorOp{Color: statusColor}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: image.Pt(size, size)}
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, state)
			return label.Layout(gtx)
		}),
	)
}

// drawDashboard shows backup statistics and controls
func (a *App) drawDashboard(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H6(a.theme, "Dashboard")
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Stats cards
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.drawStatsCards(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

			// Control buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.drawControlButtons(gtx)
			}),

			// Notification area
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if a.notification != "" {
					return layout.Inset{Top: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(a.theme, a.notification)
						label.Color = color.NRGBA{R: 0xF4, G: 0x43, B: 0x36, A: 0xFF}
						return label.Layout(gtx)
					})
				}
				return layout.Dimensions{}
			}),
		)
	})
}

// drawStatsCards displays backup statistics
func (a *App) drawStatsCards(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.drawStatCard(gtx, "Files Scanned", fmt.Sprintf("%d", a.stats.FilesScanned))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.drawStatCard(gtx, "Files Uploaded", fmt.Sprintf("%d", a.stats.FilesUploaded))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.drawStatCard(gtx, "Bytes Uploaded", formatBytes(a.stats.BytesUploaded))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.drawStatCard(gtx, "Files Failed", fmt.Sprintf("%d", a.stats.FilesFailed))
		}),
	)
}

// drawStatCard draws a single stat card
func (a *App) drawStatCard(gtx layout.Context, label, value string) layout.Dimensions {
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.Caption(a.theme, label)
				title.Color = color.NRGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF}
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				val := material.H4(a.theme, value)
				val.Color = color.NRGBA{R: 0x3F, G: 0x51, B: 0xB5, A: 0xFF}
				return val.Layout(gtx)
			}),
		)
	})
}

// drawControlButtons shows the backup control buttons
func (a *App) drawControlButtons(gtx layout.Context) layout.Dimensions {
	engineRunning := a.engine != nil

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !engineRunning {
				btn := material.Button(a.theme, &a.startBtn, "Start Auto-Backup")
				btn.Background = color.NRGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if engineRunning {
				btn := material.Button(a.theme, &a.stopBtn, "Stop Auto-Backup")
				btn.Background = color.NRGBA{R: 0xF4, G: 0x43, B: 0x36, A: 0xFF}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if engineRunning && a.stats.CurrentState != "paused" {
				btn := material.Button(a.theme, &a.pauseBtn, "Pause")
				btn.Background = color.NRGBA{R: 0xFF, G: 0x98, B: 0x00, A: 0xFF}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if engineRunning && a.stats.CurrentState == "paused" {
				btn := material.Button(a.theme, &a.resumeBtn, "Resume")
				btn.Background = color.NRGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(a.theme, &a.scanBtn, "Scan Now")
			return btn.Layout(gtx)
		}),
	)
}

// Helper functions for backup operations
func (a *App) startBackup(ctx context.Context) error {
	if a.engine != nil {
		return fmt.Errorf("backup engine already running")
	}

	engine, err := backup.NewEngine(a.client, a.config)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	if err := engine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}

	a.engine = engine
	a.notification = "Auto-backup started"
	return nil
}

func (a *App) stopBackup() error {
	if a.engine == nil {
		return fmt.Errorf("backup engine not running")
	}

	if err := a.engine.Stop(); err != nil {
		return err
	}

	a.engine = nil
	a.notification = "Auto-backup stopped"
	return nil
}

func (a *App) pauseBackup() error {
	if a.engine == nil {
		return fmt.Errorf("backup engine not running")
	}

	a.engine.Pause()
	a.notification = "Auto-backup paused"
	return nil
}

func (a *App) resumeBackup() error {
	if a.engine == nil {
		return fmt.Errorf("backup engine not running")
	}

	a.engine.Resume()
	a.notification = "Auto-backup resumed"
	return nil
}

func (a *App) performScan(ctx context.Context) error {
	if a.engine == nil {
		return fmt.Errorf("backup engine not running")
	}

	go func() {
		if err := a.engine.PerformInitialScan(ctx); err != nil {
			a.notification = fmt.Sprintf("Scan error: %v", err)
		} else {
			a.notification = "Scan completed"
		}
	}()

	return nil
}

// formatBytes formats byte count to human-readable string
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
