package backup

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
)

func Launch(app fyne.App, exitOnDone bool) fyne.Window {
	w := app.NewWindow("Maeve - Backing Up")
	loadState(w, exitOnDone)
	w.SetContent(mainWindow())
	w.Resize(fyne.NewSize(500, 1000))
	return w
}

func mainWindow() fyne.CanvasObject {
	go runBackup()

	main := container.NewBorder(
		nil,
		cancelBtn(),
		nil, nil,
		container.NewGridWithRows(2,
			container.NewBorder(
				fynext.ColorWithin(
					fynext.H2("Updating My State"),
					global.currTask,
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
					global.currTask,
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

	return container.NewPadded(main)
}

func cancelBtn() fyne.CanvasObject {
	var cancelBtn *widget.Button

	onTap := func() {
		if cancelBtn.Importance == widget.MediumImportance {
			global.ctxCancel()
		} else {
			fyne.Do(global.window.Close)
		}
	}

	cancelBtn = widget.NewButton("Cancel", onTap)

	go func() {
		<-global.ctx.Done()
		fyne.Do(func() {
			cancelBtn.SetText("Okay")
			cancelBtn.Importance = widget.SuccessImportance
			cancelBtn.Refresh()
		})
	}()

	return cancelBtn
}

func showErrorDialog(err error) {
	var d *dialog.CustomDialog

	cancelBtn := func() fyne.CanvasObject {
		cancelBtn := widget.NewButton("OK", func() {
			d.Dismiss()
		})
		cancelBtn.Importance = widget.HighImportance
		return cancelBtn
	}
	lbl := widget.NewLabel(err.Error())
	lbl.Wrapping = fyne.TextWrapWord

	d = dialog.NewCustomWithoutButtons(
		"ERROR",
		container.NewVBox(
			fynext.StackOfSize(fyne.NewSize(250, 1)),
			lbl,
			cancelBtn(),
		),
		global.window,
	)

	d.Show()
}
