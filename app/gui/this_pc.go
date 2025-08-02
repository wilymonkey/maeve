package gui

import (
	"regexp"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/theme"
)

func thisPC() fyne.CanvasObject {
	return container.NewHBox(
		container.New(
			layout.NewCustomPaddedVBoxLayout(0),
			editableLabel("Name", Global.Name, "^[A-Za-z0-9_-]+$", 25),
			backupDir(),
		),
		container.New(
			layout.NewCustomPaddedVBoxLayout(0),
			editableLabel("Max Backups to Keep", binding.IntToString(Global.MaxBackups), `^\d+$`, 5),
			editableLabel("Max Upload Speed", Global.MaxUpload, `^\s*\d+(\.\d+)?\s*(k|K|m|M|g|G)?(b|B)\s*$`, 15),
		),
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
		nil,
		nil,
		widget.NewLabel(name),
		editBtn,
		textStack,
	)
}

func backupDir() fyne.CanvasObject {
	btn := theme.EditBtn(
		func() bool { return false }, // Always in edit mode.
		func() {},
		func() {
			dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
				if err != nil {
					Global.ShowError(err)
					return
				}
				if list == nil {
					return // User canceled
				}
				Global.BackupDir.Set(list.Path())
			}, Global.Window).Show()
		})
	return container.NewHBox(
		widget.NewLabel("Backup Folder"),
		widget.NewLabelWithData(Global.BackupDir),
		btn,
	)
}
