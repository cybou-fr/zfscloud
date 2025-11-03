package gui

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/ioustamora/zfscloud/shared/config"
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
	s := &SettingsView{
		app: app,
	}
	s.LoadSettings()
	return s
}

// LoadSettings loads current config into the form fields
func (s *SettingsView) LoadSettings() {
	if s.app.config == nil {
		return
	}

	cfg := s.app.config

	// Server settings
	s.serverAddr.SetText(cfg.Client.ServerAddress)

	// TLS settings
	s.tlsEnabled.Value = cfg.Client.TLS.Enabled
	s.tlsCertPath.SetText(cfg.Client.TLS.CertFile)
	s.tlsKeyPath.SetText(cfg.Client.TLS.KeyFile)
	s.tlsCACertPath.SetText(cfg.Client.TLS.CAFile)

	// Backup settings
	folders := ""
	for i, folder := range cfg.Backup.Folders {
		if i > 0 {
			folders += "\n"
		}
		folders += folder
	}
	s.backupFolders.SetText(folders)

	excludes := ""
	for i, pattern := range cfg.Backup.ExcludePatterns {
		if i > 0 {
			excludes += "\n"
		}
		excludes += pattern
	}
	s.excludePaths.SetText(excludes)

	// Convert scan interval to string
	s.scanInterval.SetText(fmt.Sprintf("%d", cfg.Backup.Schedule.CheckIntervalMinutes))

	// Advanced settings - use default values if not set
	s.maxWorkers.SetText("4")
	s.chunkSize.SetText("4")
	s.maxRetries.SetText("3")
}

// Layout renders the settings view
func (s *SettingsView) Layout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				theme := s.app.GetTheme()
				title := theme.H5("Settings")
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
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.H6("Server Connection")
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
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.H6("TLS/Security")
			return label.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(theme.Theme, &s.tlsEnabled, "Enable TLS/mTLS").Layout(gtx)
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
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.H6("Backup Configuration")
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
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.H6("Advanced Settings")
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

	theme := s.app.GetTheme()

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(theme.Theme, &s.testButton, "Test Connection")
			btn.Background = theme.Info
			btn.Color = theme.OnPrimary
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(theme.Theme, &s.saveButton, "Save Settings")
			btn.Background = theme.Success
			btn.Color = theme.OnPrimary
			return btn.Layout(gtx)
		}),
	)
}

// textField renders a single-line text input field
func (s *SettingsView) textField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	editor.SingleLine = true
	editor.Submit = false
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := theme.Body2(label)
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(theme.Theme, editor, hint)
			return e.Layout(gtx)
		}),
	)
}

// textAreaField renders a multi-line text input field
func (s *SettingsView) textAreaField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	editor.SingleLine = false
	editor.Submit = false
	theme := s.app.GetTheme()

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := theme.Body2(label)
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(theme.Theme, editor, hint)
			// Set minimum height for text area
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(80))
			return e.Layout(gtx)
		}),
	)
}

// saveSettings saves the current settings to config file
func (s *SettingsView) saveSettings() {
	if s.app.config == nil {
		log.Println("Error: No config loaded")
		return
	}

	cfg := s.app.config

	// Update config from form fields
	cfg.Client.ServerAddress = s.serverAddr.Text()
	cfg.Client.TLS.Enabled = s.tlsEnabled.Value
	cfg.Client.TLS.CertFile = s.tlsCertPath.Text()
	cfg.Client.TLS.KeyFile = s.tlsKeyPath.Text()
	cfg.Client.TLS.CAFile = s.tlsCACertPath.Text()

	// Parse backup folders
	foldersText := s.backupFolders.Text()
	if foldersText != "" {
		cfg.Backup.Folders = strings.Split(foldersText, "\n")
		// Trim whitespace
		for i := range cfg.Backup.Folders {
			cfg.Backup.Folders[i] = strings.TrimSpace(cfg.Backup.Folders[i])
		}
	}

	// Parse exclude patterns
	excludeText := s.excludePaths.Text()
	if excludeText != "" {
		cfg.Backup.ExcludePatterns = strings.Split(excludeText, "\n")
		// Trim whitespace
		for i := range cfg.Backup.ExcludePatterns {
			cfg.Backup.ExcludePatterns[i] = strings.TrimSpace(cfg.Backup.ExcludePatterns[i])
		}
	}

	// Parse scan interval
	if intervalStr := s.scanInterval.Text(); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			cfg.Backup.Schedule.CheckIntervalMinutes = interval
		}
	}

	// Save to file
	if err := config.SaveClientConfig(cfg, s.app.configPath); err != nil {
		log.Printf("Failed to save config: %v", err)
		return
	}

	log.Println("Settings saved successfully!")
}

// testConnection tests the connection to the server
func (s *SettingsView) testConnection() {
	log.Println("Testing connection...")

	// Try to connect in a goroutine so we don't block UI
	go func() {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.app.client == nil {
			log.Println("Error: No client available")
			return
		}

		// Try to connect
		if err := s.app.client.Connect(); err != nil {
			log.Printf("Connection test failed: %v", err)
			return
		}

		log.Println("Connection test successful!")

		// Close the test connection
		s.app.client.Close()
	}()
}
