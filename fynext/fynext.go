package fynext

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/wilymonkey/maeve/fynext/internal/icons"
)

type Theme struct{}

func (m *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// -----------------------------------
	// BASICS
	// -----------------------------------
	case theme.ColorNameOverlayBackground:
		return color.White
	case theme.ColorNameBackground:
		return color.White
	case theme.ColorNameForeground:
		return color.Black
	// -----------------------------------
	// BUTTONS, INPUT & TEXT
	// -----------------------------------
	case theme.ColorNameButton:
		return zinc400
	case theme.ColorNameDisabledButton:
		return zinc200
	case theme.ColorNameDisabled:
		return zinc400
	case theme.ColorNamePrimary:
		return green400
	case theme.ColorNameForegroundOnPrimary:
		return color.Black
	case theme.ColorNameInputBackground: // Takes colour from ColorNameForeground.
		return zinc200
	case theme.ColorNameInputBorder:
		return color.Black
	case theme.ColorNameError:
		return red600
	case theme.ColorNameForegroundOnError:
		return color.White
	case theme.ColorNameWarning:
		return yellow400
	case theme.ColorNameSuccess:
		return green600
	case theme.ColorNameForegroundOnSuccess:
		return color.White
	// -----------------------------------
	// OTHER
	// -----------------------------------
	case theme.ColorNameScrollBar:
		return color.Black
	case theme.ColorNameScrollBarBackground:
		return zinc200
	case theme.ColorNameSeparator:
		return zinc200
	case theme.ColorNameShadow:
		return zinc200
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	switch name {
	case theme.IconNameCheckButtonFill:
		return icons.CheckboxSvg
	case theme.IconNameCheckButtonChecked:
		return icons.CheckboxCheckedSvg
	case theme.IconNameCheckButton:
		return icons.CheckboxSvg
	case theme.IconNameConfirm:
		return icons.CheckSvg
	case theme.IconNameContentClear:
		return icons.CancelSvg
	default:
		return theme.DefaultTheme().Icon(name)
	}
}

func (m *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 16.0
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// -----------------------------------
// COLORS
// -----------------------------------

var (
	zinc100 = color.NRGBA{R: 0xF4, G: 0xF4, B: 0xF5, A: 0xFF}
	zinc200 = color.NRGBA{R: 0xE4, G: 0xE4, B: 0xE7, A: 0xFF}
	zinc400 = color.NRGBA{R: 0xA1, G: 0xA1, B: 0xAA, A: 0xFF}
	zinc700 = color.NRGBA{R: 0x3F, G: 0x3F, B: 0x46, A: 0xFF}

	green400 = color.NRGBA{R: 0x4A, G: 0xDE, B: 0x80, A: 0xFF}

	red600 = color.NRGBA{R: 0xDC, G: 0x26, B: 0x26, A: 0xFF}

	yellow400 = color.NRGBA{R: 0xFB, G: 0xB9, B: 0x51, A: 0xFF}

	green600 = color.NRGBA{R: 0x16, G: 0xA3, B: 0x4A, A: 0xFF}
)
