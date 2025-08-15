package gui

import (
	"regexp"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/clipboard"
	"github.com/wilymonkey/maeve/conf"
	"github.com/wilymonkey/maeve/theme"
	"golang.org/x/crypto/ssh"
)

func thisPC() fyne.CanvasObject {
	return container.NewGridWithColumns(2,
		editableLabel(
			"Name",
			Global.Name,
			"^[A-Za-z0-9_-]+$",
			25,
		),
		editableLabel(
			"Max Backups to Keep",
			binding.IntToString(Global.MaxBackups),
			`^\d+$`,
			5,
		),
		backupDir(),
		editableLabel(
			"Max Upload Speed",
			Global.MaxUpload,
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
	textStack := container.NewStack(label, entry)

	var editBtn *widget.Button
	editBtn = theme.EditBtn(
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
	entry.OnChanged = func(s string) {
		reg := regexp.MustCompile(isValidReg)
		if !reg.MatchString(s) || utf8.RuneCountInString(s) > maxChars {
			editBtn.Disable()
		} else {
			editBtn.Enable()
		}
	}

	return container.NewBorder(
		nil, nil,
		NewFixedWidthLabel(name, 150),
		editBtn,
		textStack,
	)
}

func backupDir() fyne.CanvasObject {
	openFinder := func() {
		dialog.NewFolderOpen(
			func(list fyne.ListableURI, err error) {
				if err != nil {
					Global.ShowError(err)
				}
				if list != nil {
					Global.BackupDir.Set(list.Path())
				}
			},
			Global.Window,
		).Show()
	}
	return container.NewBorder(
		nil, nil,
		widget.NewLabel("Backup Folder"),
		theme.PencilBtn(openFinder),
		widget.NewLabelWithData(Global.BackupDir),
	)
}

func maeveKey() fyne.CanvasObject {
	copyKey := func() {
		p := conf.GetConf().SSHPrivateKey.Public()
		pubKey, err := ssh.NewPublicKey(p)
		if err != nil {
			Global.ShowError(err)
			return
		}
		key := ssh.MarshalAuthorizedKey(pubKey)
		if err := clipboard.WriteAll(string(key)); err != nil {
			Global.ShowError(err)
			return
		}
	}
	return container.NewBorder(
		nil, nil, nil,
		theme.DupliBtn(copyKey),
		widget.NewLabel("Maeve Key"),
	)
}

func trustedPCs() fyne.CanvasObject {
	openDialog := func() {}
	return container.NewBorder(
		nil, nil, nil,
		theme.PencilBtn(openDialog),
		widget.NewLabel("Trusted PCs"),
	)
}

type FixedWidthLabel struct {
	widget.Label
	width float32
}

func NewFixedWidthLabel(text string, width float32) *FixedWidthLabel {
	fl := &FixedWidthLabel{width: width}
	fl.ExtendBaseWidget(fl)
	fl.SetText(text)
	fl.Wrapping = fyne.TextWrapWord
	return fl
}

func (f *FixedWidthLabel) MinSize() fyne.Size {
	orig := f.Label.Size()
	f.Label.Resize(fyne.NewSize(f.width, orig.Height))
	min := f.Label.MinSize()
	min.Width = f.width
	return min
}
