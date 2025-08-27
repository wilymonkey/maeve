package gui

import (
	"regexp"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/app/clipboard"
	"github.com/wilymonkey/maeve/app/conf"
	"github.com/wilymonkey/maeve/fynext"
	"golang.org/x/crypto/ssh"
)

func thisPC() fyne.CanvasObject {
	return container.NewGridWithColumns(2,
		editableLabel(
			"Name",
			global.Name,
			"^[A-Za-z0-9_-]+$",
			25,
		),
		editableLabel(
			"Max Backups to Keep",
			binding.IntToString(global.MaxBackups),
			`^\d+$`,
			5,
		),
		backupDir(),
		editableLabel(
			"Max Upload Speed",
			global.MaxUpload,
			`^\s*\d+(\.\d+)?\s*(k|K|m|M|g|G)?(b|B)\s*$`,
			15,
		),
		maeveKey(),
		trustedPCs(),
	)
}

func editableLabel(
	name string,
	data binding.String,
	isValidReg string,
	maxChars int,
) *fyne.Container {
	label := widget.NewLabelWithData(data)
	entry := widget.NewEntry()
	entry.Hide()
	entry.Validator = nil

	var editBtn *widget.Button
	editBtn = fynext.EditBtn(
		label.Visible,
		func() {
			label.Hide()
			entry.SetText(label.Text)
			entry.Show()
		},
		func() {
			data.Set(entry.Text)
			entry.Hide()
			label.Show()
		},
	)

	reg := regexp.MustCompile(isValidReg)
	entry.OnChanged = func(s string) {
		if !reg.MatchString(s) || utf8.RuneCountInString(s) > maxChars {
			editBtn.Disable()
		} else {
			editBtn.Enable()
		}
	}

	return container.NewBorder(
		nil, nil, nil,
		editBtn,
		fynext.VBox(
			fynext.SmallTxt(name),
			container.NewHScroll(
				container.NewStack(label, entry),
			),
		),
	)
}

func backupDir() fyne.CanvasObject {
	finderDlg := func() {
		dialog.ShowFolderOpen(
			func(list fyne.ListableURI, err error) {
				if err != nil {
					global.ShowError(err)
				}
				if list != nil {
					global.MaeveDir.Set(list.Path())
				}
			},
			global.Window,
		)
	}
	return container.NewBorder(
		nil, nil, nil,
		fynext.PencilBtn(finderDlg),
		fynext.VBox(
			fynext.SmallTxt("Backup Folder"),
			container.NewHScroll(
				widget.NewLabelWithData(global.MaeveDir),
			),
		),
	)
}

func maeveKey() fyne.CanvasObject {
	copyKey := func() {
		p := conf.GetConf().SSHPrivateKey.Public()
		pubKey, err := ssh.NewPublicKey(p)
		if err != nil {
			global.ShowError(err)
			return
		}
		key := ssh.MarshalAuthorizedKey(pubKey)
		if err := clipboard.WriteAll(string(key)); err != nil {
			global.ShowError(err)
			return
		}
	}
	info := widget.NewLabel("Copy this key to the backup PC so that it will accept files.")

	return container.NewBorder(
		nil, nil, nil,
		fynext.DupliBtn(copyKey),
		fynext.VBox(
			fynext.SmallTxt("Maeve Key"),
			container.NewHScroll(info),
		),
	)
}

func trustedPCs() fyne.CanvasObject {
	openDialog := func() {}
	info := widget.NewLabel("PCs that have accepted files from this PC before.")
	btn := fynext.PencilBtn(openDialog)

	return container.NewBorder(
		nil, nil, nil,
		btn,
		fynext.VBox(
			fynext.SmallTxt("Trusted PCs"),
			container.NewHScroll(info),
		),
	)
}
