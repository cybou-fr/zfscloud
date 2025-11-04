package gui

import (
	"fmt"

	"github.com/getlantern/systray"
)

// Placeholder for icon data
var iconData = []byte{}

func (a *App) onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("ZFSCloud")
	systray.SetTooltip("ZFSCloud Backup")

	// Navigation menu items
	mDashboard := systray.AddMenuItem("Dashboard (Ctrl+D)", "Switch to dashboard view")
	mFiles := systray.AddMenuItem("Files (Ctrl+F)", "Switch to files view")
	mSettings := systray.AddMenuItem("Settings (Ctrl+S)", "Switch to settings view")
	systray.AddSeparator()

	// Action menu items
	mRefresh := systray.AddMenuItem("Refresh (Ctrl+R)", "Refresh statistics")
	mStartStop := systray.AddMenuItem("Start/Stop Backup", "Toggle backup engine")
	systray.AddSeparator()

	// Window control
	mShow := systray.AddMenuItem("Show Window", "Show the main application window")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
			case <-mDashboard.ClickedCh:
				a.SetView(ViewDashboard)
				a.ShowToast(ToastTypeInfo, "Switched to Dashboard")
			case <-mFiles.ClickedCh:
				a.SetView(ViewFiles)
				a.ShowToast(ToastTypeInfo, "Switched to Files")
			case <-mSettings.ClickedCh:
				a.SetView(ViewSettings)
				a.ShowToast(ToastTypeInfo, "Switched to Settings")
			case <-mRefresh.ClickedCh:
				a.refreshStatistics()
				a.ShowToast(ToastTypeInfo, "Statistics refreshed")
			case <-mStartStop.ClickedCh:
				a.toggleBackup()
			case <-mShow.ClickedCh:
				a.showWindow()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Initial state
	a.updateSystray()
}

func (a *App) onExit() {
	// Clean up resources here
	if a.engine != nil {
		if err := a.engine.Stop(); err != nil {
			fmt.Printf("Error stopping engine: %v\n", err)
		}
	}
}

func (a *App) updateSystray() {
	// This function can be used to update the systray menu items dynamically
	// For example, to show connection status or last backup time
}
