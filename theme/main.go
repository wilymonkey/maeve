package theme

import (
	_ "embed"
	imgColor "image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/wilymonkey/maeve/theme/internal/color"
	"github.com/wilymonkey/maeve/theme/internal/resource"
)

type Theme struct{}

func (m *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) imgColor.Color {
	switch name {
	case theme.ColorNameBackground:
		return imgColor.White
	case theme.ColorNameForeground:
		return color.Zinc950
	case theme.ColorNameDisabled:
		return color.Zinc400
	case theme.ColorNameButton:
		return color.Green400
	case theme.ColorNamePrimary:
		return color.Green400
	case theme.ColorNameForegroundOnPrimary:
		return color.Zinc100
	case theme.ColorNameScrollBar:
		return color.Zinc950
	case theme.ColorNameSeparator:
		return color.Zinc200
	case theme.ColorNameShadow:
		return color.Zinc300
	case theme.ColorNameInputBackground:
		return color.Zinc300
	case theme.ColorNameInputBorder:
		return color.Zinc950
	case theme.ColorNameDisabledButton:
		return color.Zinc300
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
		return resource.CheckboxSvg
	case theme.IconNameCheckButtonChecked:
		return resource.CheckboxCheckedSvg
	case theme.IconNameCheckButton:
		return resource.CheckboxSvg
	case theme.IconNameConfirm:
		return resource.CheckSvg
	case theme.IconNameContentClear:
		return resource.CancelSvg
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
