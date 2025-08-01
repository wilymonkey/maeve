package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/wilymonkey/maeve/theme"
)

func Render() fyne.CanvasObject {
	return container.NewScroll(
		container.NewPadded(
			container.NewBorder(
				container.NewHBox(
					favicon(64),
					theme.NewH1("Maeve"),
				),
				theme.HighBtn("Backup Now", func() {}),
				nil,
				nil,
				body(),
			),
		),
	)
}

func body() *fyne.Container {
	info := container.NewScroll(
		container.NewCenter(theme.NewH2("Info")),
	)
	info.SetMinSize(fyne.NewSquareSize(200))

	nodes := container.NewScroll(
		container.NewCenter(theme.NewH2("Backup Locations")),
	)
	nodes.SetMinSize(fyne.NewSquareSize(200))

	dirs := container.NewScroll(
		container.NewCenter(theme.NewH2("Folders")),
	)
	dirs.SetMinSize(fyne.NewSquareSize(200))

	return container.NewVBox(
		theme.NewH2("This PC Info"),
		theme.LowPriorBox(info),
		theme.NewH2("Backup PCs"),
		theme.LowPriorBox(nodes),
		theme.NewH2("Folders"),
		theme.LowPriorBox(dirs),
	)
}
