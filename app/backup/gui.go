package backup

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/help"
	"github.com/wilymonkey/maeve/fynext"
)

func Launch(app fyne.App, exitOnDone bool) fyne.Window {
	window := app.NewWindow("Maeve - Backing Up")
	loadState(window, exitOnDone)
	window.SetContent(mainWindow())
	return window
}

func mainWindow() fyne.CanvasObject {
	go runBackup()

	return container.NewBorder(
		nil,
		cancelBtn(),
		nil, nil,
		container.NewGridWithRows(2,
			container.NewBorder(
				fynext.ColorWithin(
					fynext.H2("Updating My State"),
					guiState.currTask,
					repairingDB,
					pushing,
				),
				nil, nil, nil,
				container.NewBorder(
					repairDBUI(),
					updateDBUI(),
					nil, nil,
					pullStateTable(),
				),
			),
			container.NewBorder(
				fynext.ColorWithin(
					fynext.H2("Sending Files to PCs"),
					guiState.currTask,
					pushing,
					done,
				),
				nil, nil, nil,
				container.NewGridWithRows(2,
					nodeStateGrid(),
					pushStateCard(),
				),
			),
		),
	)
}

func cancelBtn() fyne.CanvasObject {
	var cancelBtn *widget.Button

	onTap := func() {
		if cancelBtn.Importance == widget.MediumImportance {
			guiState.ctxCancel()
		} else {
			fyne.Do(guiState.window.Close)
		}
	}

	cancelBtn = widget.NewButton("Cancel", onTap)

	go func() {
		<-guiState.ctx.Done()
		fyne.Do(func() {
			cancelBtn.SetText("Okay")
			cancelBtn.Importance = widget.SuccessImportance
			cancelBtn.Refresh()
		})
	}()

	return cancelBtn
}

func showErrorDialog(err error) {
	dlg := func(w fyne.Window) dialog.Dialog {
		var d *dialog.CustomDialog

		cancelBtn := func() fyne.CanvasObject {
			cancelBtn := widget.NewButton("OK", func() { d.Dismiss() })
			cancelBtn.Importance = widget.HighImportance
			return cancelBtn
		}

		d = dialog.NewCustomWithoutButtons(
			"ERROR",
			container.NewVBox(
				help.Render(err),
				cancelBtn(),
			),
			w,
		)
		return d
	}

	fyne.Do(func() {
		fynext.ShowWindowDialog(guiState.window, "ERROR", dlg)
	})
}
