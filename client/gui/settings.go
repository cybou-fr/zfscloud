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

	// Validation
	validator        *Validator
	validationErrors map[string]string

	// Tooltips
	tooltipTest *Tooltip
	tooltipSave *Tooltip
}

// NewSettingsView creates a new settings view
func NewSettingsView(app *App) *SettingsView {
	s := &SettingsView{
		app:              app,
		validator:        NewValidator(),
		validationErrors: make(map[string]string),
		tooltipTest:      NewTooltip("Test connection to server with current settings"),
		tooltipSave:      NewTooltip("Save configuration changes to file"),
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
			return s.validatedTextField(gtx, &s.serverAddr, "Server Address", "localhost:9090", "serverAddr")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.textField(gtx, &s.serverPort, "Server Port", "9090")
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
			return s.validatedTextAreaField(gtx, &s.backupFolders, "Backup Folders (one per line)", "C:\\Users\\Documents\nC:\\Users\\Pictures", "backupFolders")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.validatedTextAreaField(gtx, &s.excludePaths, "Exclude Patterns (one per line)", "*.tmp\n*.log\n.git", "excludePaths")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.validatedTextField(gtx, &s.scanInterval, "Scan Interval (minutes)", "5", "scanInterval")
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
			return s.tooltipTest.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(theme.Theme, &s.testButton, IconSymbols[IconTest]+" Test Connection")
				btn.Background = theme.Info
				btn.Color = theme.OnPrimary
				return btn.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.tooltipSave.Layout(gtx, theme, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(theme.Theme, &s.saveButton, IconSymbols[IconSave]+" Save Settings")
				btn.Background = theme.Success
				btn.Color = theme.OnPrimary
				return btn.Layout(gtx)
			})
		}),
	)
}

// textField renders a single-line text input field with validation
func (s *SettingsView) textField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	return s.validatedTextField(gtx, editor, label, hint, "")
}

// validatedTextField renders a text field with validation error display
func (s *SettingsView) validatedTextField(gtx layout.Context, editor *widget.Editor, label, hint, fieldKey string) layout.Dimensions {
	editor.SingleLine = true
	editor.Submit = false
	theme := s.app.GetTheme()

	// Check if there's a validation error for this field
	hasError := false
	errorMsg := ""
	if fieldKey != "" {
		if msg, exists := s.validationErrors[fieldKey]; exists {
			hasError = true
			errorMsg = msg
		}
	}

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
			// Highlight with error color if validation failed
			if hasError {
				e.Color = theme.Error
			}
			return e.Layout(gtx)
		}),
		// Show error message if validation failed
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !hasError {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				errorLabel := material.Body2(theme.Theme, errorMsg)
				errorLabel.Color = theme.Error
				errorLabel.TextSize = unit.Sp(12)
				return errorLabel.Layout(gtx)
			})
		}),
	)
}

// textAreaField renders a multi-line text input field with validation
func (s *SettingsView) textAreaField(gtx layout.Context, editor *widget.Editor, label, hint string) layout.Dimensions {
	return s.validatedTextAreaField(gtx, editor, label, hint, "")
}

// validatedTextAreaField renders a text area with validation error display
func (s *SettingsView) validatedTextAreaField(gtx layout.Context, editor *widget.Editor, label, hint, fieldKey string) layout.Dimensions {
	editor.SingleLine = false
	editor.Submit = false
	theme := s.app.GetTheme()

	// Check if there's a validation error for this field
	hasError := false
	errorMsg := ""
	if fieldKey != "" {
		if msg, exists := s.validationErrors[fieldKey]; exists {
			hasError = true
			errorMsg = msg
		}
	}

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
			if hasError {
				e.Color = theme.Error
			}
			// Set minimum height for text area
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(80))
			return e.Layout(gtx)
		}),
		// Show error message if validation failed
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !hasError {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				errorLabel := material.Body2(theme.Theme, errorMsg)
				errorLabel.Color = theme.Error
				errorLabel.TextSize = unit.Sp(12)
				return errorLabel.Layout(gtx)
			})
		}),
	)
}

// saveSettings saves the current settings to config file
func (s *SettingsView) saveSettings() {
	if s.app.config == nil {
		log.Println("Error: No config loaded")
		return
	}

	// Clear previous validation errors
	s.validationErrors = make(map[string]string)

	// Validate all fields
	if result := s.validator.ValidateServerAddress(s.serverAddr.Text()); !result.Valid {
		s.validationErrors["serverAddr"] = result.Message
		s.app.ShowToast(ToastTypeError, result.Message)
		return
	}

	if result := s.validator.ValidateBackupFolders(s.backupFolders.Text()); !result.Valid {
		s.validationErrors["backupFolders"] = result.Message
		s.app.ShowToast(ToastTypeError, result.Message)
		return
	}

	if result := s.validator.ValidateExcludePatterns(s.excludePaths.Text()); !result.Valid {
		s.validationErrors["excludePaths"] = result.Message
		s.app.ShowToast(ToastTypeError, result.Message)
		return
	}

	if s.scanInterval.Text() != "" {
		if result := s.validator.ValidatePositiveInt(s.scanInterval.Text(), "Scan interval"); !result.Valid {
			s.validationErrors["scanInterval"] = result.Message
			s.app.ShowToast(ToastTypeError, result.Message)
			return
		}
	}

	// Validate TLS paths if TLS is enabled
	if s.tlsEnabled.Value {
		if result := s.validator.ValidateFilePath(s.tlsCertPath.Text(), true); !result.Valid {
			s.validationErrors["tlsCertPath"] = result.Message
			s.app.ShowToast(ToastTypeError, "TLS Certificate: "+result.Message)
			return
		}
		if result := s.validator.ValidateFilePath(s.tlsKeyPath.Text(), true); !result.Valid {
			s.validationErrors["tlsKeyPath"] = result.Message
			s.app.ShowToast(ToastTypeError, "TLS Key: "+result.Message)
			return
		}
		if result := s.validator.ValidateFilePath(s.tlsCACertPath.Text(), true); !result.Valid {
			s.validationErrors["tlsCACertPath"] = result.Message
			s.app.ShowToast(ToastTypeError, "TLS CA Certificate: "+result.Message)
			return
		}
	}

	cfg := s.app.config

	// Update config from form fields (all validated now)
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
		s.app.ShowToast(ToastTypeError, fmt.Sprintf("Failed to save settings: %v", err))
		return
	}

	log.Println("Settings saved successfully!")
	s.app.ShowToast(ToastTypeSuccess, "Settings saved successfully!")
}

// testConnection tests the connection to the server
func (s *SettingsView) testConnection() {
	log.Println("Testing connection...")
	s.app.ShowToast(ToastTypeInfo, "Testing connection...")

	// Try to connect in a goroutine so we don't block UI
	go func() {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.app.client == nil {
			log.Println("Error: No client available")
			s.app.ShowToast(ToastTypeError, "No client configured")
			return
		}

		// Try to connect
		if err := s.app.client.Connect(); err != nil {
			log.Printf("Connection test failed: %v", err)
			s.app.ShowToast(ToastTypeError, fmt.Sprintf("Connection failed: %v", err))
			return
		}

		log.Println("Connection test successful!")
		s.app.ShowToast(ToastTypeSuccess, "Connection test successful!")

		// Close the test connection
		s.app.client.Close()
	}()
}
