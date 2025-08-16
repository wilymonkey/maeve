package gui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

//go:embed icon.svg
var faviconBytes []byte
var faviconSvg = &fyne.StaticResource{
	StaticName:    "favicon.svg",
	StaticContent: faviconBytes,
}

func favicon(size float32) *canvas.Image {
	img := canvas.NewImageFromResource(faviconSvg)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSquareSize(size))
	return img
}
