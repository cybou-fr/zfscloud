package gui

import (
	"image/color"

	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// AppTheme extends material.Theme with custom colors and styling
type AppTheme struct {
	*material.Theme

	// Semantic colors
	Primary   color.NRGBA
	Secondary color.NRGBA
	Success   color.NRGBA
	Warning   color.NRGBA
	Error     color.NRGBA
	Info      color.NRGBA

	// Surface colors
	Background color.NRGBA
	Surface    color.NRGBA
	SurfaceAlt color.NRGBA

	// Text colors
	OnPrimary     color.NRGBA
	OnSurface     color.NRGBA
	OnBackground  color.NRGBA
	TextSecondary color.NRGBA
	TextDisabled  color.NRGBA

	// State colors
	Disabled color.NRGBA
	Divider  color.NRGBA
	Overlay  color.NRGBA

	// Status indicator colors
	StatusConnected    color.NRGBA
	StatusDisconnected color.NRGBA
	StatusRunning      color.NRGBA
	StatusPaused       color.NRGBA
	StatusError        color.NRGBA

	// Spacing
	Unit unit.Dp
}

// ThemeMode represents the theme mode
type ThemeMode int

const (
	ThemeLight ThemeMode = iota
	ThemeDark
)

// NewAppTheme creates a new application theme
func NewAppTheme(mode ThemeMode) *AppTheme {
	baseTheme := material.NewTheme()
	baseTheme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	theme := &AppTheme{
		Theme: baseTheme,
		Unit:  unit.Dp(8),
	}

	if mode == ThemeLight {
		theme.applyLightTheme()
	} else {
		theme.applyDarkTheme()
	}

	return theme
}

// applyLightTheme applies light theme colors
func (t *AppTheme) applyLightTheme() {
	// Semantic colors
	t.Primary = rgb(0x1976D2)   // Blue 700
	t.Secondary = rgb(0x424242) // Grey 800
	t.Success = rgb(0x388E3C)   // Green 700
	t.Warning = rgb(0xF57C00)   // Orange 700
	t.Error = rgb(0xD32F2F)     // Red 700
	t.Info = rgb(0x0288D1)      // Light Blue 700

	// Surface colors
	t.Background = rgb(0xFAFAFA) // Grey 50
	t.Surface = rgb(0xFFFFFF)    // White
	t.SurfaceAlt = rgb(0xF5F5F5) // Grey 100

	// Text colors
	t.OnPrimary = rgb(0xFFFFFF) // White
	t.OnSurface = rgba(0, 0, 0, 0.87)
	t.OnBackground = rgba(0, 0, 0, 0.87)
	t.TextSecondary = rgba(0, 0, 0, 0.60)
	t.TextDisabled = rgba(0, 0, 0, 0.38)

	// State colors
	t.Disabled = rgba(0, 0, 0, 0.12)
	t.Divider = rgba(0, 0, 0, 0.12)
	t.Overlay = rgba(0, 0, 0, 0.50)

	// Status colors
	t.StatusConnected = rgb(0x4CAF50)    // Green 500
	t.StatusDisconnected = rgb(0xF44336) // Red 500
	t.StatusRunning = rgb(0x2196F3)      // Blue 500
	t.StatusPaused = rgb(0xFF9800)       // Orange 500
	t.StatusError = rgb(0xD32F2F)        // Red 700

	// Update base theme colors
	t.Theme.Palette.Bg = t.Background
	t.Theme.Palette.Fg = t.OnBackground
	t.Theme.Palette.ContrastBg = t.Primary
	t.Theme.Palette.ContrastFg = t.OnPrimary
}

// applyDarkTheme applies dark theme colors
func (t *AppTheme) applyDarkTheme() {
	// Semantic colors
	t.Primary = rgb(0x90CAF9)   // Blue 200
	t.Secondary = rgb(0xBDBDBD) // Grey 400
	t.Success = rgb(0x66BB6A)   // Green 400
	t.Warning = rgb(0xFFB74D)   // Orange 300
	t.Error = rgb(0xEF5350)     // Red 400
	t.Info = rgb(0x4FC3F7)      // Light Blue 300

	// Surface colors
	t.Background = rgb(0x121212) // Material Dark
	t.Surface = rgb(0x1E1E1E)    // Elevated surface
	t.SurfaceAlt = rgb(0x2C2C2C) // Higher elevation

	// Text colors
	t.OnPrimary = rgba(0, 0, 0, 0.87)
	t.OnSurface = rgb(0xFFFFFF)
	t.OnBackground = rgb(0xFFFFFF)
	t.TextSecondary = rgba(255, 255, 255, 0.70)
	t.TextDisabled = rgba(255, 255, 255, 0.50)

	// State colors
	t.Disabled = rgba(255, 255, 255, 0.12)
	t.Divider = rgba(255, 255, 255, 0.12)
	t.Overlay = rgba(0, 0, 0, 0.70)

	// Status colors
	t.StatusConnected = rgb(0x66BB6A)    // Green 400
	t.StatusDisconnected = rgb(0xEF5350) // Red 400
	t.StatusRunning = rgb(0x42A5F5)      // Blue 400
	t.StatusPaused = rgb(0xFFA726)       // Orange 400
	t.StatusError = rgb(0xEF5350)        // Red 400

	// Update base theme colors
	t.Theme.Palette.Bg = t.Background
	t.Theme.Palette.Fg = t.OnBackground
	t.Theme.Palette.ContrastBg = t.Primary
	t.Theme.Palette.ContrastFg = t.OnPrimary
}

// Spacing helpers
func (t *AppTheme) Space(multiplier float32) unit.Dp {
	return unit.Dp(float32(t.Unit) * multiplier)
}

// Common spacing values
func (t *AppTheme) SpaceTiny() unit.Dp   { return t.Space(0.5) } // 4dp
func (t *AppTheme) SpaceSmall() unit.Dp  { return t.Space(1) }   // 8dp
func (t *AppTheme) SpaceMedium() unit.Dp { return t.Space(2) }   // 16dp
func (t *AppTheme) SpaceLarge() unit.Dp  { return t.Space(3) }   // 24dp
func (t *AppTheme) SpaceXLarge() unit.Dp { return t.Space(4) }   // 32dp

// Typography helpers
func (t *AppTheme) H1(text string) material.LabelStyle {
	l := material.H1(t.Theme, text)
	return l
}

func (t *AppTheme) H2(text string) material.LabelStyle {
	l := material.H2(t.Theme, text)
	return l
}

func (t *AppTheme) H3(text string) material.LabelStyle {
	l := material.H3(t.Theme, text)
	return l
}

func (t *AppTheme) H4(text string) material.LabelStyle {
	l := material.H4(t.Theme, text)
	return l
}

func (t *AppTheme) H5(text string) material.LabelStyle {
	l := material.H5(t.Theme, text)
	return l
}

func (t *AppTheme) H6(text string) material.LabelStyle {
	l := material.H6(t.Theme, text)
	return l
}

func (t *AppTheme) Body1(text string) material.LabelStyle {
	l := material.Body1(t.Theme, text)
	return l
}

func (t *AppTheme) Body2(text string) material.LabelStyle {
	l := material.Body2(t.Theme, text)
	return l
}

func (t *AppTheme) Caption(text string) material.LabelStyle {
	l := material.Caption(t.Theme, text)
	return l
}

// Helper to create buttons with semantic colors
func (t *AppTheme) PrimaryButton(btn material.ButtonStyle) material.ButtonStyle {
	btn.Background = t.Primary
	btn.Color = t.OnPrimary
	return btn
}

func (t *AppTheme) SecondaryButton(btn material.ButtonStyle) material.ButtonStyle {
	btn.Background = t.Secondary
	btn.Color = t.OnPrimary
	return btn
}

func (t *AppTheme) SuccessButton(btn material.ButtonStyle) material.ButtonStyle {
	btn.Background = t.Success
	btn.Color = t.OnPrimary
	return btn
}

func (t *AppTheme) ErrorButton(btn material.ButtonStyle) material.ButtonStyle {
	btn.Background = t.Error
	btn.Color = t.OnPrimary
	return btn
}

// Helper functions for color creation
func rgb(c uint32) color.NRGBA {
	return argb((0xff << 24) | c)
}

func rgba(r, g, b uint8, a float32) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: uint8(a * 255)}
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{
		A: uint8(c >> 24),
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
	}
}
