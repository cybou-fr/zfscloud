package gui

import (
	"fmt"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	"github.com/ioustamora/zfscloud/shared/config"
)

// App represents the GUI application
type App struct {
	window      *app.Window
	theme       *material.Theme
	client      *client.Client
	engine      *backup.Engine
	config      *config.ClientConfig
	configPath  string
	currentView View
	dashboard   *DashboardView
	settings    *SettingsView
	files       *FilesView
	stats       *Statistics

	// Navigation buttons
	navDashboard widget.Clickable
	navSettings  widget.Clickable
	navFiles     widget.Clickable

	// Error dialog
	errorDialog     ErrorDialog
	showErrorDialog bool
} // ErrorDialog represents an error dialog
type ErrorDialog struct {
	title    string
	message  string
	okButton widget.Clickable
}

// Statistics holds real-time backup statistics
type Statistics struct {
	Connected     bool
	EngineRunning bool
	EngineState   string
	FilesScanned  int64
	FilesUploaded int64
	FilesFailed   int64
	BytesUploaded int64
	LastBackup    time.Time
	CurrentFile   string
}

// View represents different views in the application
type View int

const (
	ViewDashboard View = iota
	ViewSettings
	ViewFiles
)

// NewApp creates a new GUI application
func NewApp(client *client.Client, engine *backup.Engine, cfg *config.ClientConfig, configPath string) *App {
	window := new(app.Window)
	window.Option(app.Title("ZFSCloud Backup"))
	window.Option(app.Size(unit.Dp(900), unit.Dp(600)))

	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	app := &App{
		window:      window,
		theme:       theme,
		client:      client,
		engine:      engine,
		config:      cfg,
		configPath:  configPath,
		currentView: ViewDashboard,
		stats:       &Statistics{},
	}

	// Initialize views
	app.dashboard = NewDashboardView(app)
	app.settings = NewSettingsView(app)
	app.files = NewFilesView(app)

	// Start system tray
	app.StartSystemTray()

	return app
}

// Run starts the GUI event loop
func (a *App) Run() error {
	// Start statistics updater
	go a.updateStatistics()

	var ops op.Ops
	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
} // Layout renders the main application layout
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Show error dialog overlay if visible
	if a.showErrorDialog {
		return a.layoutErrorDialog(gtx)
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// Top navigation bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutNavBar(gtx)
		}),
		// Main content area
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.layoutContent(gtx)
		}),
		// Status bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutStatusBar(gtx)
		}),
	)
}

// layoutNavBar renders the navigation bar
func (a *App) layoutNavBar(gtx layout.Context) layout.Dimensions {
	// Handle navigation button clicks
	if a.navDashboard.Clicked(gtx) {
		a.SetView(ViewDashboard)
	}
	if a.navSettings.Clicked(gtx) {
		a.SetView(ViewSettings)
	}
	if a.navFiles.Clicked(gtx) {
		a.SetView(ViewFiles)
	}

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
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Flexed(1, layout.Spacer{}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutNavButtons(gtx)
			}),
		)
	})
}

// layoutNavButtons renders the navigation buttons
func (a *App) layoutNavButtons(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(a.theme, &a.navDashboard, "Dashboard")
			if a.currentView == ViewDashboard {
				btn.Background = a.theme.Palette.ContrastBg
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(a.theme, &a.navSettings, "Settings")
			if a.currentView == ViewSettings {
				btn.Background = a.theme.Palette.ContrastBg
			}
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(a.theme, &a.navFiles, "Files")
			if a.currentView == ViewFiles {
				btn.Background = a.theme.Palette.ContrastBg
			}
			return btn.Layout(gtx)
		}),
	)
}

// layoutContent renders the main content based on current view
func (a *App) layoutContent(gtx layout.Context) layout.Dimensions {
	switch a.currentView {
	case ViewDashboard:
		return a.dashboard.Layout(gtx)
	case ViewSettings:
		return a.settings.Layout(gtx)
	case ViewFiles:
		return a.files.Layout(gtx)
	default:
		return a.dashboard.Layout(gtx)
	}
}

// layoutStatusBar renders the bottom status bar
func (a *App) layoutStatusBar(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		statusText := "Disconnected"
		if a.stats.Connected {
			statusText = "Connected"
		}
		if a.stats.EngineRunning {
			statusText = fmt.Sprintf("%s • Engine: %s", statusText, a.stats.EngineState)
		}

		label := material.Body2(a.theme, statusText)
		return label.Layout(gtx)
	})
}

// updateStatistics periodically updates statistics from the engine
func (a *App) updateStatistics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.updateStats()
		a.window.Invalidate()
	}
}

// updateStats updates the statistics from client and engine
func (a *App) updateStats() {
	if a.client != nil {
		// Check if connection is active (simplified check)
		a.stats.Connected = (a.client != nil)
	}

	if a.engine != nil {
		stats := a.engine.GetStats()
		a.stats.EngineRunning = stats.CurrentState != backup.StateIdle
		a.stats.EngineState = string(stats.CurrentState)
		a.stats.FilesScanned = stats.FilesScanned
		a.stats.FilesUploaded = stats.FilesUploaded
		a.stats.FilesFailed = stats.FilesFailed
		a.stats.BytesUploaded = stats.BytesUploaded
		a.stats.LastBackup = stats.LastBackupTime
		a.stats.CurrentFile = "" // Not tracked yet
	}
}

// SetView changes the current view
func (a *App) SetView(view View) {
	a.currentView = view
	a.window.Invalidate()
}

// GetTheme returns the application theme
func (a *App) GetTheme() *material.Theme {
	return a.theme
}

// GetStats returns the current statistics
func (a *App) GetStats() *Statistics {
	return a.stats
}

// ShowError displays an error message
func (a *App) ShowError(title, message string) {
	log.Printf("Error: %s - %s", title, message)
	a.errorDialog = ErrorDialog{
		title:   title,
		message: message,
	}
	a.showErrorDialog = true
	a.window.Invalidate()
}

// Keyboard shortcuts are available through system tray menu instead
// due to complexity of Gio keyboard event handling API.
// System tray provides:
// - Quick access to views (Dashboard, Settings, Files)
// - Backup controls (Start, Stop)
// - Quit option// layoutErrorDialog renders the error dialog overlay
func (a *App) layoutErrorDialog(gtx layout.Context) layout.Dimensions {
	// Handle OK button click
	if a.errorDialog.okButton.Clicked(gtx) {
		a.showErrorDialog = false
		a.window.Invalidate()
	}

	return layout.Stack{}.Layout(gtx,
		// Dimmed background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutNavBar(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.layoutContent(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.layoutStatusBar(gtx)
				}),
			)
		}),
		// Dialog overlay
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				// Dialog size
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(450))
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(400))

				return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						// Title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							title := material.H6(a.theme, a.errorDialog.title)
							return title.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						// Message
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							msg := material.Body1(a.theme, a.errorDialog.message)
							return msg.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
						// OK button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:    layout.Horizontal,
								Spacing: layout.SpaceEnd,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(a.theme, &a.errorDialog.okButton, "OK")
									btn.Background = a.theme.Palette.ContrastBg
									return btn.Layout(gtx)
								}),
							)
						}),
					)
				})
			})
		}),
	)
}

// Quit closes the application
func (a *App) Quit() {
	os.Exit(0)
}
