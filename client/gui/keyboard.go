package gui

// KeyboardShortcuts provides keyboard shortcut functionality
// Note: Full keyboard shortcut support requires Gio UI input focus management
// This is a simplified implementation documenting the intended shortcuts

// Keyboard shortcut documentation
const (
	ShortcutDashboard   = "Ctrl+D"
	ShortcutFiles       = "Ctrl+F"
	ShortcutSettings    = "Ctrl+S"
	ShortcutRefresh     = "Ctrl+R or F5"
	ShortcutCloseDialog = "Escape"
)

// GetShortcutHelp returns a help string describing available shortcuts
func GetShortcutHelp() string {
	return `Keyboard Shortcuts:
-------------------
Navigation:
  Ctrl+D     - Dashboard
  Ctrl+F     - Files
  Ctrl+S     - Settings

Actions:
  Ctrl+R     - Refresh statistics
  F5         - Refresh statistics
  Escape     - Close dialog/overlay

Backup Engine:
  Space      - Start/Stop backup (when on dashboard)
  P          - Pause/Resume backup (when running)

Note: Keyboard shortcuts will be fully implemented in a future update.
      Currently, please use the mouse/touch interface.
`
}

// Keyboard shortcut documentation
const (
	ShortcutHelp = `
Keyboard Shortcuts:
-------------------
Navigation:
  Ctrl+D     - Dashboard
  Ctrl+F     - Files
  Ctrl+S     - Settings

Actions:
  Ctrl+R     - Refresh statistics
  F5         - Refresh statistics
  Escape     - Close dialog/overlay

General:
  Ctrl+Q     - Quit application (future)
  Ctrl+,     - Open settings (future)
`
)
