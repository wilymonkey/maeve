package gui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

//go:embed icon.ico
var faviconBytes []byte
var FaviconIco = &fyne.StaticResource{
	StaticName:    "favicon.ico",
	StaticContent: faviconBytes,
}

//go:embed icon.svg
var faviconSvgBytes []byte
var faviconSvg = &fyne.StaticResource{
	StaticName:    "favicon.svg",
	StaticContent: faviconSvgBytes,
}

func favicon(size float32) *canvas.Image {
	img := canvas.NewImageFromResource(faviconSvg)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSquareSize(size))
	return img
}
