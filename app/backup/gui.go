package backup

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/conf"
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
				nodeStatusTable(guiState.nodeStates),
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

func nodeStatusTable(nodeStatus map[string]*nodeStatus) fyne.CanvasObject {
	cfg := conf.GetConf()
	objects := make([]fyne.CanvasObject, 0, len(cfg.RemoteNodes)*2)
	for _, node := range cfg.RemoteNodes {
		status := nodeStatus[node]
		errStatus := help.NewWidget(status.err)
		border := container.NewBorder(
			nil, nil,
			fynext.LabelDisableUntil(
				widget.NewLabel(node),
				guiState.currTask,
				pushing,
			),
			nil,
			errStatus,
		)
		sep := widget.NewSeparator()
		objects = append(objects, border, sep)
	}
	vbox := container.NewVBox()
	vbox.Objects = objects[:len(objects)-1] // Remove trailing separator.
	return container.NewVScroll(vbox)
}

func showErrorDialog() {
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
				help.NewWidget(guiState.err),
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
