package gui

import (
	"fmt"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
)

// App represents the GUI application
type App struct {
	window       *app.Window
	theme        *material.Theme
	client       *client.Client
	engine       *backup.Engine
	currentView  View
	dashboard    *DashboardView
	settings     *SettingsView
	files        *FilesView
	stats        *Statistics
}

// Statistics holds real-time backup statistics
type Statistics struct {
	Connected      bool
	EngineRunning  bool
	EngineState    string
	FilesScanned   int64
	FilesUploaded  int64
	FilesFailed    int64
	BytesUploaded  int64
	LastBackup     time.Time
	CurrentFile    string
}

// View represents different views in the application
type View int

const (
	ViewDashboard View = iota
	ViewSettings
	ViewFiles
)

// NewApp creates a new GUI application
func NewApp(client *client.Client, engine *backup.Engine) *App {
	window := app.NewWindow(
		app.Title("ZFSCloud Backup"),
		app.Size(unit.Dp(900), unit.Dp(600)),
	)

	theme := material.NewTheme()
	theme.Shaper = gofont.Collection()

	app := &App{
		window:      window,
		theme:       theme,
		client:      client,
		engine:      engine,
		currentView: ViewDashboard,
		stats:       &Statistics{},
	}

	// Initialize views
	app.dashboard = NewDashboardView(app)
	app.settings = NewSettingsView(app)
	app.files = NewFilesView(app)

	return app
}

// Run starts the GUI event loop
func (a *App) Run() error {
	// Start statistics updater
	go a.updateStatistics()

	var ops op.Ops
	for {
		switch e := a.window.Event().(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// Layout renders the main application layout
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
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
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Navigation buttons will go here
				return layout.Dimensions{}
			}),
		)
	})
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
		a.stats.Connected = a.client.IsConnected()
	}

	if a.engine != nil {
		stats := a.engine.GetStatistics()
		a.stats.EngineRunning = stats.State != "idle"
		a.stats.EngineState = stats.State
		a.stats.FilesScanned = stats.FilesScanned
		a.stats.FilesUploaded = stats.FilesUploaded
		a.stats.FilesFailed = stats.FilesFailed
		a.stats.BytesUploaded = stats.BytesUploaded
		a.stats.LastBackup = stats.LastBackup
		a.stats.CurrentFile = stats.CurrentFile
	}
}

// SetView changes the current view
func (a *App) SetView(view View) {
	a.currentView = view
	a.window.Invalidate()
}

// GetTheme returns the application theme
func (a *GetTheme() *material.Theme {
	return a.theme
}

// GetStats returns the current statistics
func (a *App) GetStats() *Statistics {
	return a.stats
}

// ShowError displays an error message
func (a *App) ShowError(err error) {
	log.Printf("Error: %v", err)
	// TODO: Implement error dialog
}

// Quit closes the application
func (a *App) Quit() {
	os.Exit(0)
}
