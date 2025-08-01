package gui

import (
	"fmt"
	"regexp"

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
			editableLabel("Name", Global.Name, "^[A-Za-z0-9_-]{1,25}$"),
			backupDir(),
		),
		container.New(
			layout.NewCustomPaddedVBoxLayout(0),
			editableLabel("Max Backups to Keep", binding.IntToString(Global.MaxBackups), "^[0-9]{1,5}$"),
			editableLabel("Max Upload Speed", Global.MaxUpload, "^[A-Za-z0-9 .]{1,15}$"),
		),
	)
}

func editableLabel(
	name string,
	data binding.String,
	isValidReg string,
) *fyne.Container {
	label := widget.NewLabelWithData(data)
	entry := widget.NewEntryWithData(data)
	entry.Hide()
	entry.Validator = func(s string) error {
		match, err := regexp.MatchString(isValidReg, s)
		if err != nil {
			panic(err)
		}
		if !match {
			return fmt.Errorf("Invalid")
		}
		return nil
	}
	textStack := container.NewStack(label, entry)

	var editBtn *widget.Button
	editBtn = widget.NewButton("Edit", func() {
		if label.Visible() {
			label.Hide()
			entry.Show()
			editBtn.SetText("Save")
		} else {
			entry.Hide()
			label.Show()
			editBtn.SetText("Edit")
		}
	})
	entry.SetOnValidationChanged(func(err error) {
		if err != nil {
			editBtn.Disable()
		} else {
			editBtn.Enable()
		}
	})

	return container.NewBorder(
		nil,
		nil,
		widget.NewLabel(name),
		editBtn,
		textStack,
	)
}

func backupDir() fyne.CanvasObject {
	btn := theme.HighBtn("Edit", func() {
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
