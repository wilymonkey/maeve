package theme

import (
	imgColor "image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/installer/theme/internal/color"
)

type Theme struct{}

func (m *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) imgColor.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.Zinc100
	case theme.ColorNameForeground:
		return color.Zinc950
	case theme.ColorNameButton:
		return color.Zinc300
	case theme.ColorNamePrimary:
		return color.Green400
	case theme.ColorNameForegroundOnPrimary:
		return color.Zinc100
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style) // replace with your own fonts
}

func (m *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name) // replace icons if needed
}

func (m *Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 16.0
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func NewH1(text string) *fyne.Container {
	return container.NewPadded(
		&canvas.Text{
			Color:     theme.Color(theme.ColorNameForeground),
			Text:      text,
			TextSize:  24.0,
			TextStyle: fyne.TextStyle{Bold: true},
		},
	)
}

func HighButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance
	return btn
}

func PrimaryBox(objects fyne.CanvasObject) *fyne.Container {
	background := canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	return container.NewStack(background, objects)
}

func Icon(size float32) *canvas.Image {
	img := canvas.NewImageFromFile("icon.svg")
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return img
}
