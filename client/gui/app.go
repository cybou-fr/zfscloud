package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/getlantern/systray"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	"github.com/ioustamora/zfscloud/shared/config"
)

// GUIPreferences holds GUI-specific user preferences
type GUIPreferences struct {
	ThemeMode ThemeMode `json:"theme_mode"`
}

// App represents the GUI application
type App struct {
	window      *app.Window
	theme       *AppTheme
	client      *client.Client
	engine      *backup.Engine
	config      *config.ClientConfig
	configPath  string
	currentView View
	dashboard   *DashboardView
	settings    *SettingsView
	files       *FilesView
	stats       *Statistics
	statsMutex  sync.RWMutex  // Protects stats from concurrent access
	stopUpdate  chan struct{} // Signal to stop statistics updater

	// Navigation buttons
	navDashboard widget.Clickable
	navSettings  widget.Clickable
	navFiles     widget.Clickable

	// Help button and dialog
	helpButton      widget.Clickable
	helpCloseButton widget.Clickable // Separate button for closing help dialog
	showHelpDialog  bool

	// Theme toggle
	themeToggleButton widget.Clickable
	themeTooltip      *Tooltip

	// Error dialog
	errorDialog     ErrorDialog
	showErrorDialog bool

	// Toast notifications
	toastManager *ToastManager

	// Responsive layout
	responsiveLayout *ResponsiveLayout

	// Tooltips for navigation
	tooltipDashboard *Tooltip
	tooltipFiles     *Tooltip
	tooltipSettings  *Tooltip
	tooltipHelp      *Tooltip

	// GUI preferences
	guiPrefs     *GUIPreferences
	guiPrefsPath string
}

// ErrorDialog represents an error dialog
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

	// Determine GUI preferences path
	guiPrefsPath := filepath.Join(filepath.Dir(configPath), "gui-prefs.json")

	// Create app instance first (without theme)
	app := &App{
		window:           window,
		client:           client,
		engine:           engine,
		config:           cfg,
		configPath:       configPath,
		currentView:      ViewDashboard,
		stats:            &Statistics{},
		stopUpdate:       make(chan struct{}),
		toastManager:     NewToastManager(),
		responsiveLayout: NewResponsiveLayout(),
		tooltipDashboard: NewTooltip("View dashboard and backup status"),
		tooltipFiles:     NewTooltip("Browse and manage backed up files"),
		tooltipSettings:  NewTooltip("Configure backup settings and connection"),
		tooltipHelp:      NewTooltip("Show keyboard shortcuts and help information"),
		themeTooltip:     NewTooltip("Toggle between light and dark theme"),
		guiPrefsPath:     guiPrefsPath,
	}

	// Load GUI preferences
	app.loadGUIPreferences()

	// Create custom theme based on preferences
	app.theme = NewAppTheme(app.guiPrefs.ThemeMode)

	// Initialize views
	app.dashboard = NewDashboardView(app)
	app.settings = NewSettingsView(app)
	app.files = NewFilesView(app)

	// Start system tray
	app.StartSystemTray()

	// Show welcome message
	app.toastManager.ShowInfo("Welcome to ZFSCloud Backup")

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
			// Stop the statistics updater goroutine
			close(a.stopUpdate)
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Handle keyboard shortcuts
			a.handleKeyboardShortcuts(gtx)

			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// handleKeyboardShortcuts processes keyboard shortcuts
// Note: Keyboard shortcuts are implemented via system tray for now
// Direct keyboard handling requires additional Gio event processing
func (a *App) handleKeyboardShortcuts(gtx layout.Context) {
	// Keyboard shortcuts are currently implemented via system tray menu
	// Direct keyboard event handling would require processing FrameEvent.Key events
	// For now, users can use the system tray menu for quick actions
}

// Layout renders the main application layout
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Show help dialog overlay if visible
	if a.showHelpDialog {
		return a.layoutHelpDialog(gtx)
	}

	// Show error dialog overlay if visible
	if a.showErrorDialog {
		return a.layoutErrorDialog(gtx)
	}

	// Render main UI with toast notifications overlay
	return layout.Stack{}.Layout(gtx,
		// Main content
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
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
		}),
		// Toast notifications overlay (top-right)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.toastManager.Layout(gtx, a.theme)
				}),
			)
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
				title := a.theme.H5("ZFSCloud Backup")
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

// layoutNavButtons renders the navigation buttons with icons (responsive)
func (a *App) layoutNavButtons(gtx layout.Context) layout.Dimensions {
	// Get responsive navigation axis
	axis := a.responsiveLayout.GetNavigationAxis(gtx)
	spacing := a.responsiveLayout.GetSpacing(gtx)

	spacer := layout.Spacer{Width: spacing}
	if axis == layout.Vertical {
		spacer = layout.Spacer{Height: spacing}
	}

	// Handle help button click
	if a.helpButton.Clicked(gtx) {
		a.showHelpDialog = true
	}

	// Handle theme toggle click
	if a.themeToggleButton.Clicked(gtx) {
		a.toggleTheme()
	}

	return layout.Flex{
		Axis:    axis,
		Spacing: layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tooltipDashboard.Layout(gtx, a.theme, func(gtx layout.Context) layout.Dimensions {
				return a.navButton(gtx, &a.navDashboard, IconDashboard, "Dashboard", ViewDashboard)
			})
		}),
		layout.Rigid(spacer.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tooltipFiles.Layout(gtx, a.theme, func(gtx layout.Context) layout.Dimensions {
				return a.navButton(gtx, &a.navFiles, IconFiles, "Files", ViewFiles)
			})
		}),
		layout.Rigid(spacer.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tooltipSettings.Layout(gtx, a.theme, func(gtx layout.Context) layout.Dimensions {
				return a.navButton(gtx, &a.navSettings, IconSettings, "Settings", ViewSettings)
			})
		}),
		layout.Rigid(spacer.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tooltipHelp.Layout(gtx, a.theme, func(gtx layout.Context) layout.Dimensions {
				// Help button (info icon)
				btn := material.Button(a.theme.Theme, &a.helpButton, IconSymbols[IconInfo]+" Help")
				btn.Background = a.theme.Info
				btn.Color = a.theme.OnPrimary
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(spacer.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.themeTooltip.Layout(gtx, a.theme, func(gtx layout.Context) layout.Dimensions {
				// Theme toggle button
				themeIcon := "☀" // Light mode icon
				if a.theme == NewAppTheme(ThemeDark) {
					themeIcon = "🌙" // Dark mode icon
				}
				btn := material.Button(a.theme.Theme, &a.themeToggleButton, themeIcon)
				btn.Background = a.theme.Secondary
				btn.Color = a.theme.OnPrimary
				return btn.Layout(gtx)
			})
		}),
	)
}

// navButton renders a single navigation button with icon
func (a *App) navButton(gtx layout.Context, btn *widget.Clickable, iconType IconType, text string, view View) layout.Dimensions {
	// Get icon symbol and create button text
	iconSymbol := IconSymbols[iconType]
	buttonText := iconSymbol + " " + text

	button := material.Button(a.theme.Theme, btn, buttonText)
	if a.currentView == view {
		button.Background = a.theme.Primary
		button.Color = a.theme.OnPrimary
	}

	return button.Layout(gtx)
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

// layoutStatusBar renders the bottom status bar with enhanced information
func (a *App) layoutStatusBar(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Build status text
		statusText := "Disconnected"
		if a.stats.Connected {
			statusText = "Connected"
		}
		if a.stats.EngineRunning {
			statusText = fmt.Sprintf("%s • Engine: %s", statusText, a.stats.EngineState)
		}

		// Add last update time if available
		if !a.stats.LastBackup.IsZero() {
			elapsed := time.Since(a.stats.LastBackup)
			if elapsed < time.Minute {
				statusText = fmt.Sprintf("%s • Last backup: %ds ago", statusText, int(elapsed.Seconds()))
			} else if elapsed < time.Hour {
				statusText = fmt.Sprintf("%s • Last backup: %dm ago", statusText, int(elapsed.Minutes()))
			} else if elapsed < 24*time.Hour {
				statusText = fmt.Sprintf("%s • Last backup: %dh ago", statusText, int(elapsed.Hours()))
			}
		}

		return layout.Flex{
			Axis:      layout.Horizontal,
			Spacing:   layout.SpaceBetween,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// Left side: Status text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := a.theme.Body2(statusText)
				return label.Layout(gtx)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			// Right side: Version info
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				versionText := "ZFSCloud v1.0"
				version := a.theme.Caption(versionText)
				version.Color = a.theme.TextSecondary
				return version.Layout(gtx)
			}),
		)
	})
}

// updateStatistics periodically updates statistics from the engine
func (a *App) updateStatistics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Lock stats for writing
			a.statsMutex.Lock()
			a.updateStats()
			a.statsMutex.Unlock()

			// Request UI redraw
			a.window.Invalidate()
		case <-a.stopUpdate:
			// Stop signal received, exit goroutine
			return
		}
	}
}

// updateStats updates the statistics from client and engine
func (a *App) updateStats() {
	if a.client != nil {
		// Check if connection is actually active
		a.stats.Connected = a.client.IsConnected()
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
func (a *App) GetTheme() *AppTheme {
	return a.theme
}

// GetStats returns a safe copy of the current statistics
func (a *App) GetStats() *Statistics {
	a.statsMutex.RLock()
	defer a.statsMutex.RUnlock()

	// Return a copy to avoid race conditions
	statsCopy := *a.stats
	return &statsCopy
}

// ShowError displays an error message
func (a *App) ShowError(title, message string) {
	log.Printf("Error: %s - %s", title, message)

	// Show error dialog for critical errors
	a.errorDialog = ErrorDialog{
		title:   title,
		message: message,
	}
	a.showErrorDialog = true

	// Also show error toast
	a.toastManager.ShowError(fmt.Sprintf("%s: %s", title, message))
	a.window.Invalidate()
}

// ShowToast shows a toast notification
func (a *App) ShowToast(toastType ToastType, message string) {
	switch toastType {
	case ToastTypeInfo:
		a.toastManager.ShowInfo(message)
	case ToastTypeSuccess:
		a.toastManager.ShowSuccess(message)
	case ToastTypeWarning:
		a.toastManager.ShowWarning(message)
	case ToastTypeError:
		a.toastManager.ShowError(message)
	}
	a.window.Invalidate()
}

// Keyboard shortcuts are available through system tray menu instead
// due to complexity of Gio keyboard event handling API.
// System tray provides:
// - Quick access to views (Dashboard, Settings, Files)
// - Backup controls (Start, Stop)
// layoutHelpDialog renders the help dialog overlay
func (a *App) layoutHelpDialog(gtx layout.Context) layout.Dimensions {
	// Handle close button click (using separate close button to avoid conflicts)
	for a.helpCloseButton.Clicked(gtx) {
		a.showHelpDialog = false
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
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(550))
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(500))

				return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						// Title
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							title := a.theme.H5("ℹ ZFSCloud Backup - Help")
							return title.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
						// Help content
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							helpText := GetShortcutHelp()
							msg := a.theme.Body2(helpText)
							return msg.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
						// Version and info
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							info := a.theme.Caption("Version: 1.0.0 | Built with Gio UI")
							info.Color = a.theme.TextSecondary
							return info.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
						// Close button
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:    layout.Horizontal,
								Spacing: layout.SpaceEnd,
							}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									btn := material.Button(a.theme.Theme, &a.helpCloseButton, "Close")
									btn.Background = a.theme.Primary
									btn.Color = a.theme.OnPrimary
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

// layoutErrorDialog renders the error dialog overlay
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
							title := a.theme.H6(a.errorDialog.title)
							return title.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						// Message
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							msg := a.theme.Body1(a.errorDialog.message)
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
									btn := material.Button(a.theme.Theme, &a.errorDialog.okButton, "OK")
									btn.Background = a.theme.Error
									btn.Color = a.theme.OnPrimary
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

// toggleTheme switches between light and dark themes
func (a *App) toggleTheme() {
	// Toggle theme
	newMode := ThemeLight
	if a.guiPrefs.ThemeMode == ThemeLight {
		newMode = ThemeDark
	}

	a.guiPrefs.ThemeMode = newMode
	a.theme = NewAppTheme(newMode)
	a.window.Invalidate()

	// Save theme preference
	a.saveGUIPreferences()

	// Show toast notification
	themeName := "Light"
	if newMode == ThemeDark {
		themeName = "Dark"
	}
	a.ShowToast(ToastTypeInfo, fmt.Sprintf("Switched to %s theme", themeName))
}

// loadGUIPreferences loads GUI preferences from file
func (a *App) loadGUIPreferences() {
	a.guiPrefs = &GUIPreferences{
		ThemeMode: ThemeLight, // Default to light theme
	}

	if _, err := os.Stat(a.guiPrefsPath); os.IsNotExist(err) {
		// File doesn't exist, use defaults
		return
	}

	data, err := os.ReadFile(a.guiPrefsPath)
	if err != nil {
		log.Printf("Failed to read GUI preferences: %v", err)
		return
	}

	if err := json.Unmarshal(data, a.guiPrefs); err != nil {
		log.Printf("Failed to parse GUI preferences: %v", err)
		// Reset to defaults on parse error
		a.guiPrefs = &GUIPreferences{ThemeMode: ThemeLight}
	}
}

// saveGUIPreferences saves GUI preferences to file
func (a *App) saveGUIPreferences() {
	data, err := json.MarshalIndent(a.guiPrefs, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal GUI preferences: %v", err)
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(a.guiPrefsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create GUI preferences directory: %v", err)
		return
	}

	if err := os.WriteFile(a.guiPrefsPath, data, 0644); err != nil {
		log.Printf("Failed to save GUI preferences: %v", err)
	}
}

func (a *App) StartSystemTray() {
	go systray.Run(a.onReady, a.onExit)
}

// refreshStatistics manually refreshes the statistics display
func (a *App) refreshStatistics() {
	a.statsMutex.Lock()
	a.updateStats()
	a.statsMutex.Unlock()
	a.window.Invalidate()
}

// toggleBackup toggles the backup engine start/stop state
func (a *App) toggleBackup() {
	if a.engine == nil {
		a.ShowToast(ToastTypeError, "Backup engine not available")
		return
	}

	stats := a.engine.GetStats()
	currentState := stats.CurrentState

	// Check if engine is currently running (not idle)
	isRunning := currentState != backup.StateIdle && currentState != backup.StateError

	if isRunning {
		// Stop backup
		if err := a.engine.Stop(); err != nil {
			a.ShowError("Failed to stop backup", err.Error())
			return
		}
		a.ShowToast(ToastTypeSuccess, "Backup stopped")
	} else {
		// Start backup
		ctx := context.Background()
		if err := a.engine.Start(ctx); err != nil {
			a.ShowError("Failed to start backup", err.Error())
			return
		}
		a.ShowToast(ToastTypeSuccess, "Backup started")
	}
}

// showWindow brings the application window to the front
func (a *App) showWindow() {
	// Request focus for the window
	a.window.Invalidate()
}
