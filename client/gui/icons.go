package gui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Icon represents a displayable icon
type Icon struct {
	Symbol string
	Color  color.NRGBA
	Size   unit.Sp
}

// IconType represents different icon types
type IconType int

const (
	IconDashboard IconType = iota
	IconSettings
	IconFiles
	IconStart
	IconStop
	IconPause
	IconResume
	IconRefresh
	IconSuccess
	IconError
	IconWarning
	IconInfo
	IconConnected
	IconDisconnected
	IconUpload
	IconDownload
	IconFolder
	IconFile
	IconSave
	IconTest
	IconClock
	IconCheck
	IconClose
	IconSearch
)

// IconSymbols maps icon types to Unicode symbols
var IconSymbols = map[IconType]string{
	IconDashboard:    "📊",
	IconSettings:     "⚙",
	IconFiles:        "📁",
	IconStart:        "▶",
	IconStop:         "■",
	IconPause:        "⏸",
	IconResume:       "⏯",
	IconRefresh:      "🔄",
	IconSuccess:      "✓",
	IconError:        "✗",
	IconWarning:      "⚠",
	IconInfo:         "ℹ",
	IconConnected:    "●",
	IconDisconnected: "○",
	IconUpload:       "↑",
	IconDownload:     "↓",
	IconFolder:       "📂",
	IconFile:         "📄",
	IconSave:         "💾",
	IconTest:         "🔌",
	IconClock:        "🕐",
	IconCheck:        "✔",
	IconClose:        "✕",
	IconSearch:       "🔍",
}

// NewIcon creates a new icon with default size
func NewIcon(iconType IconType, theme *AppTheme) *Icon {
	return &Icon{
		Symbol: IconSymbols[iconType],
		Color:  theme.OnSurface,
		Size:   unit.Sp(20),
	}
}

// WithColor sets the icon color
func (i *Icon) WithColor(c color.NRGBA) *Icon {
	i.Color = c
	return i
}

// WithSize sets the icon size
func (i *Icon) WithSize(size unit.Sp) *Icon {
	i.Size = size
	return i
}

// Layout renders the icon
func (i *Icon) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	label := material.Label(theme.Theme, i.Size, i.Symbol)
	label.Color = i.Color
	return label.Layout(gtx)
}

// IconButton creates a button with an icon and text
type IconButton struct {
	Icon *Icon
	Text string
}

// NewIconButton creates a new icon button
func NewIconButton(iconType IconType, text string, theme *AppTheme) *IconButton {
	return &IconButton{
		Icon: NewIcon(iconType, theme),
		Text: text,
	}
}

// Layout renders the icon button content
func (ib *IconButton) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ib.Icon.Layout(gtx, theme)
		}),
		layout.Rigid(layout.Spacer{Width: theme.SpaceSmall()}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.Body1(ib.Text)
			return label.Layout(gtx)
		}),
	)
}

// StatusIcon creates an icon for connection status
func StatusIcon(connected bool, theme *AppTheme) *Icon {
	if connected {
		return NewIcon(IconConnected, theme).WithColor(theme.StatusConnected).WithSize(unit.Sp(16))
	}
	return NewIcon(IconDisconnected, theme).WithColor(theme.StatusDisconnected).WithSize(unit.Sp(16))
}

// ActionIcon creates a semantic action icon
func ActionIcon(iconType IconType, theme *AppTheme) *Icon {
	icon := NewIcon(iconType, theme)

	switch iconType {
	case IconStart, IconSuccess, IconCheck:
		icon.Color = theme.Success
	case IconStop, IconError, IconClose:
		icon.Color = theme.Error
	case IconPause, IconWarning:
		icon.Color = theme.Warning
	case IconRefresh, IconInfo, IconTest:
		icon.Color = theme.Info
	}

	return icon
}

// IconLabel combines an icon with a label
type IconLabel struct {
	Icon  *Icon
	Label string
}

// NewIconLabel creates a new icon label
func NewIconLabel(iconType IconType, label string, theme *AppTheme) *IconLabel {
	return &IconLabel{
		Icon:  NewIcon(iconType, theme),
		Label: label,
	}
}

// Layout renders the icon label
func (il *IconLabel) Layout(gtx layout.Context, theme *AppTheme) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return il.Icon.Layout(gtx, theme)
		}),
		layout.Rigid(layout.Spacer{Width: theme.SpaceSmall()}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := theme.Body1(il.Label)
			return label.Layout(gtx)
		}),
	)
}
