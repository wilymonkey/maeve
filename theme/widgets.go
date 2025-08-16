package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/wilymonkey/maeve/theme/internal/icons"
	"github.com/wilymonkey/maeve/utils"
)

func H1(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  theme.Size(theme.SizeNameHeadingText) * 1.5,
		TextStyle: fyne.TextStyle{Bold: true},
	}
}

func H2(text string) *canvas.Text {
	return &canvas.Text{
		Color:     theme.Color(theme.ColorNameForeground),
		Text:      text,
		TextSize:  theme.Size(theme.SizeNameHeadingText),
		TextStyle: fyne.TextStyle{Bold: true},
	}
}
func SmallText(text string) *canvas.Text {
	return &canvas.Text{
		Color:    zinc700,
		Text:     text,
		TextSize: theme.Size(theme.SizeNameCaptionText),
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

func PencilBtn(tapped func()) *widget.Button {
	return widget.NewButtonWithIcon("", icons.PencilSvg, tapped)
}

func DupliBtn(tapped func()) *widget.Button {
	return widget.NewButtonWithIcon("", icons.DuplicateSvg, tapped)
}

func EditBtn(onPred func() bool, onEdit func(), onConfirm func()) *widget.Button {
	var btn *widget.Button
	btn = PencilBtn(func() {
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

func GreyBox(objects fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(zinc100)
	background.CornerRadius = 8
	background.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	background.StrokeWidth = 1
	return container.NewStack(background, objects)
}

func ErrorBox(ErrMsg binding.String) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameError))
	bg.CornerRadius = 8
	bg.StrokeWidth = 1
	iconRes := theme.NewColoredResource(icons.DangerSvg, theme.ColorNameWarning)
	icon := canvas.NewImageFromResource(iconRes)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSquareSize(theme.IconInlineSize()))

	title := canvas.NewText(
		"ERROR",
		theme.Color(theme.ColorNameForegroundOnError),
	)
	title.TextSize = 20
	title.TextStyle.Bold = true

	richtext := widget.NewRichTextWithText("")
	text := richtext.Segments[0].(*widget.TextSegment)
	text.Style.ColorName = theme.ColorNameForegroundOnError
	text.Style.TextStyle.Bold = true
	richtext.Wrapping = fyne.TextWrapWord

	ErrMsg.AddListener(binding.NewDataListener(func() {
		text.Text = utils.GetOrPanic(ErrMsg)
	}))
	return container.NewStack(
		bg,
		container.NewPadded(
			container.NewBorder(
				container.NewHBox(icon, title),
				nil, nil, nil,
				richtext,
			),
		),
	)
}

// Vbox with no padding.
func VBox(objects ...fyne.CanvasObject) fyne.CanvasObject {
	return container.New(
		layout.NewCustomPaddedVBoxLayout(0),
		objects...,
	)
}

// Show a dialog on a separate window.
func ShowWindowDialog(
	currWindow fyne.Window,
	title string,
	dlg func(w fyne.Window) dialog.Dialog,
) {
	lock := dialog.NewCustomWithoutButtons(
		"Dialog Open",
		widget.NewLabel("A dialog is open, close it to unlock this window."),
		currWindow,
	)

	w := fyne.CurrentApp().NewWindow(title)
	w.SetContent(canvas.NewRectangle(color.Black))
	w.SetFixedSize(true)
	w.SetPadded(false)
	w.SetOnClosed(func() { lock.Dismiss() })

	d := dlg(w)

	lock.Show()
	w.Show()
	d.Show()
	d.SetOnClosed(func() { w.Close() })
}
