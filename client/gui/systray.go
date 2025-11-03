package gui

import (
	"log"

	"github.com/getlantern/systray"
)

// StartSystemTray initializes the system tray icon and menu
func (a *App) StartSystemTray() {
	go func() {
		systray.Run(a.onTrayReady, a.onTrayExit)
	}()
}

// onTrayReady is called when the system tray is ready
func (a *App) onTrayReady() {
	// Set tray icon (you would need an .ico file for Windows)
	// For now, use a simple icon placeholder
	systray.SetTitle("ZFSCloud")
	systray.SetTooltip("ZFSCloud Backup Client")

	// Create menu items
	mShow := systray.AddMenuItem("Show Window", "Show the main window")
	mDashboard := systray.AddMenuItem("Dashboard", "Open dashboard")
	mSettings := systray.AddMenuItem("Settings", "Open settings")
	mFiles := systray.AddMenuItem("Files", "Browse files")
	systray.AddSeparator()
	mStartBackup := systray.AddMenuItem("Start Backup", "Start backup operation")
	mStopBackup := systray.AddMenuItem("Stop Backup", "Stop backup operation")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu item clicks
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				log.Println("System tray: Show window")
				// Note: Gio doesn't have a built-in way to show/hide window from code
				// This would require platform-specific implementation

			case <-mDashboard.ClickedCh:
				log.Println("System tray: Dashboard")
				a.SetView(ViewDashboard)

			case <-mSettings.ClickedCh:
				log.Println("System tray: Settings")
				a.SetView(ViewSettings)

			case <-mFiles.ClickedCh:
				log.Println("System tray: Files")
				a.SetView(ViewFiles)

			case <-mStartBackup.ClickedCh:
				log.Println("System tray: Start backup")
				if a.engine != nil {
					go a.engine.Start()
				}

			case <-mStopBackup.ClickedCh:
				log.Println("System tray: Stop backup")
				if a.engine != nil {
					a.engine.Stop()
				}

			case <-mQuit.ClickedCh:
				log.Println("System tray: Quit")
				systray.Quit()
				a.Quit()
				return
			}
		}
	}()
}

// onTrayExit is called when the system tray exits
func (a *App) onTrayExit() {
	log.Println("System tray exited")
}

// StopSystemTray stops the system tray
func (a *App) StopSystemTray() {
	systray.Quit()
}
