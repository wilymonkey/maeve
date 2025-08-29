package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/fynext"
	"github.com/wilymonkey/maeve/fynext/icons"
)

func backupPCs() fyne.CanvasObject {
	openDialog := func() {
		var dlg dialog.Dialog
		dlg = dialog.NewCustomWithoutButtons(
			"Backup PCs",
			remoteNodeDlg(dlg),
			global.Window,
		)
		dlg.Show()
	}
	info := widget.NewLabel("PCs that have accepted files from this PC before.")
	info.Wrapping = fyne.TextWrapWord
	btn := widget.NewButtonWithIcon("", icons.PencilSvg, openDialog)

	return container.NewBorder(
		nil, nil, nil,
		btn,
		fynext.VBox(
			fynext.SmallTxt("Trusted PCs"),
			fynext.StackOfSize(
				fyne.NewSize(200, 1),
				info,
			),
		),
	)
}

func remoteNodeDlg(dlg dialog.Dialog) fyne.CanvasObject {
	return container.NewVScroll(
		container.NewVBox(
			remoteNodeList(),
			addRemoteNote(),
		),
	)
}

func remoteNodeList() fyne.CanvasObject {
	emptyRows := func() fyne.CanvasObject {
		label := widget.NewLabel("")
		deleteBtn := fynext.DeleteBtn(func() {})
		return container.NewBorder(
			nil, nil, nil,
			deleteBtn,
			label,
		)
	}
	updateRows := func(i binding.DataItem, o fyne.CanvasObject) {
		row := o.(*fyne.Container)
		label := row.Objects[0].(*widget.Label)
		deleteBtn := row.Objects[1].(*widget.Button)
		label.Bind(i.(binding.String))
		deleteBtn.OnTapped = func() {
			global.RemoteNodes.Remove(label.Text)
		}
	}
	w := container.NewVScroll(
		fynext.GreyBox(
			container.NewPadded(
				widget.NewListWithData(
					global.RemoteNodes,
					emptyRows,
					updateRows,
				),
			),
		),
	)
	return w
}

func addRemoteNote() fyne.CanvasObject {
	inputEntry := widget.NewEntry()
	addButton := fynext.HighBtn("  +  ", func() {
		global.RemoteNodes.Append(inputEntry.Text)
		inputEntry.SetText("")
	})
	addButton.Disable()
	inputEntry.OnChanged = func(s string) {
		if sshNodeReg.MatchString(s) {
			addButton.Enable()
		} else {
			addButton.Disable()
		}
	}

	return container.NewBorder(
		nil, nil, nil,
		addButton,
		inputEntry,
	)
}
