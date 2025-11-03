package gui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// SettingsView displays the settings panel
type SettingsView struct {
	app        *App
	saveButton widget.Clickable
	testButton widget.Clickable

	// Server settings
	serverAddr widget.Editor
	serverPort widget.Editor

	// TLS settings
	tlsEnabled    widget.Bool
	tlsCertPath   widget.Editor
	tlsKeyPath    widget.Editor
	tlsCACertPath widget.Editor

	// Backup settings
	backupFolders widget.Editor
	excludePaths  widget.Editor
	scanInterval  widget.Editor

	// Advanced settings
	maxWorkers widget.Editor
	chunkSize  widget.Editor
	maxRetries widget.Editor
}

// NewSettingsView creates a new settings view
func NewSettingsView(app *App) *SettingsView {
	return &SettingsView{
		app: app,
	}
}

// Layout renders the settings view
func (s *SettingsView) Layout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H5(s.app.theme, "Settings")
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			// Server connection settings
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutServerSettings(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

			// TLS settings
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutTLSSettings(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

			// Backup settings
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutBackupSettings(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

			// Advanced settings
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutAdvancedSettings(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),

			// Action buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutActions(gtx)
			}),
		)
	})
}

// layoutServerSettings renders server connection settings
func (s *SettingsView) layoutServerSettings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H6(s.app.theme, "Server Connection")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.serverAddr, "Server Address", "localhost")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.serverPort, "Server Port", "50051")
		}),
	)
}

// layoutTLSSettings renders TLS security settings
func (s *SettingsView) layoutTLSSettings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H6(s.app.theme, "TLS/Security")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(s.app.theme, &s.tlsEnabled, "Enable TLS/mTLS").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.tlsEnabled.Value {
				gtx = gtx.Disabled()
			}
			return s.textField(gtx, &s.tlsCertPath, "Client Certificate", "./certs/client.crt")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.tlsEnabled.Value {
				gtx = gtx.Disabled()
			}
			return s.textField(gtx, &s.tlsKeyPath, "Client Key", "./certs/client.key")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.tlsEnabled.Value {
				gtx = gtx.Disabled()
			}
			return s.textField(gtx, &s.tlsCACertPath, "CA Certificate", "./certs/ca.crt")
		}),
	)
}

// layoutBackupSettings renders backup configuration settings
func (s *SettingsView) layoutBackupSettings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H6(s.app.theme, "Backup Configuration")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textAreaField(gtx, &s.backupFolders, "Backup Folders (one per line)", "/home/user/Documents\n/home/user/Pictures")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textAreaField(gtx, &s.excludePaths, "Exclude Patterns (one per line)", "*.tmp\n*.log\n.git")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.scanInterval, "Scan Interval (seconds)", "300")
		}),
	)
}

// layoutAdvancedSettings renders advanced configuration settings
func (s *SettingsView) layoutAdvancedSettings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H6(s.app.theme, "Advanced Settings")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.maxWorkers, "Max Upload Workers", "4")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.chunkSize, "Chunk Size (MB)", "4")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.maxRetries, "Max Retries", "3")
		}),
	)
}

// layoutActions renders save and test buttons
func (s *SettingsView) layoutActions(gtx layout.Context) layout.Dimensions {
	// Handle button clicks
	if s.saveButton.Clicked(gtx) {
		s.saveSettings()
	}
	if s.testButton.Clicked(gtx) {
		s.testConnection()
	}

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(s.app.theme, &s.testButton, "Test Connection")
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(s.app.theme, &s.saveButton, "Save Settings")
			btn.Background = s.app.theme.Palette.ContrastBg
			return btn.Layout(gtx)
		}),
	)
}

// textField renders a single-line text input field
func (s *SettingsView) textField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	editor.SingleLine = true
	editor.Submit = false

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(s.app.theme, label)
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(s.app.theme, editor, hint)
			return e.Layout(gtx)
		}),
	)
}

// textAreaField renders a multi-line text input field
func (s *SettingsView) textAreaField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	editor.SingleLine = false
	editor.Submit = false

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(s.app.theme, label)
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(s.app.theme, editor, hint)
			// Set minimum height for text area
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(80))
			return e.Layout(gtx)
		}),
	)
}

// saveSettings saves the current settings to config file
func (s *SettingsView) saveSettings() {
	// TODO: Implement config saving
	// 1. Validate all input fields
	// 2. Update client config
	// 3. Save to config file
	// 4. Show success/error message
}

// testConnection tests the connection to the server
func (s *SettingsView) testConnection() {
	// TODO: Implement connection test
	// 1. Apply current settings temporarily
	// 2. Try to connect to server
	// 3. Show success/error message
}
