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

	// Menu items
	mShow := systray.AddMenuItem("Show Window", "Show the main application window")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
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
