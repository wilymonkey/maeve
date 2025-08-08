package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/theme/internal/icons"
	"github.com/wilymonkey/maeve/utils"
)

func NewH1(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  24.0,
		TextStyle: fyne.TextStyle{Bold: true},
	}
}

func NewH2(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  20.0,
		TextStyle: fyne.TextStyle{Bold: true},
	}
}
func HighBtn(label string, tapped func()) *widget.Button {
	btn := widget.NewButton(label, tapped)
	btn.Importance = widget.HighImportance
	return btn
}

func DeleteBtn(tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", icons.TrashSvg, tapped)
	btn.Importance = widget.DangerImportance
	return btn
}

func EditBtn(onPred func() bool, onEdit func(), onConfirm func()) *widget.Button {
	var btn *widget.Button
	btn = widget.NewButtonWithIcon("", icons.PencilSvg, func() {
		if onPred() {
			onEdit()
			btn.SetIcon(icons.CheckSvg)
		} else {
			onConfirm()
			btn.SetIcon(icons.PencilSvg)
		}
	})
	return btn
}

func PrimaryBox(objects fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	return container.NewStack(background, objects)
}

func GreyBox(objects fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(zinc100)
	background.CornerRadius = 8
	background.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	background.StrokeWidth = 1
	return container.NewStack(background, objects)
}

func DangerBox(msg binding.String) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameError))
	bg.CornerRadius = 8
	bg.StrokeWidth = 1
	iconRes := theme.NewColoredResource(icons.DangerSvg, theme.ColorNameWarning)
	icon := canvas.NewImageFromResource(iconRes)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSquareSize(theme.IconInlineSize()))
	text := canvas.NewText("", theme.Color(theme.ColorNameForegroundOnError))
	text.TextStyle = fyne.TextStyle{Bold: true}
	msg.AddListener(binding.NewDataListener(func() {
		text.Text = utils.GetOrPanic(msg)
	}))
	return container.NewStack(
		bg,
		container.NewPadded(container.NewHBox(icon, text)),
	)
}
