package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"github.com/wilymonkey/maeve/installer/resource"
)

func NewH1(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  24.0,
		TextStyle: fyne.TextStyle{Bold: true},
	}
}

func NewH2(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  20.0,
		TextStyle: fyne.TextStyle{Bold: true},
	}
}

func BoolImg(data binding.Bool) *canvas.Image {
	img := canvas.NewImageFromResource(resource.CheckSvg)
	img.FillMode = canvas.ImageFillOriginal
	img.SetMinSize(fyne.NewSize(20, 25))

	data.AddListener(binding.NewDataListener(func() {
		val, err := data.Get()
		if err != nil {
			panic(err)
		}
		if val {
			img.Resource = resource.CheckSvg
		} else {
			img.Resource = resource.CancelSvg
		}
		img.Refresh()
	}))

	return img
}
