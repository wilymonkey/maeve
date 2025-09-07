package fynext

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
	"github.com/wilymonkey/maeve/icons"
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
func SmallTxt(text string) *canvas.Text {
	return &canvas.Text{
		Color:    zinc700,
		Text:     text,
		TextSize: theme.Size(theme.SizeNameCaptionText),
	}
}
func RedBoldTxt(text string) *canvas.Text {
	return &canvas.Text{
		Color:     red600,
		Text:      text,
		TextSize:  theme.Size(theme.SizeNameCaptionText),
		TextStyle: fyne.TextStyle{Bold: true},
	}
}

func LabelBold(text string) *widget.Label {
	return widget.NewLabelWithStyle(
		text,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
}

func LabelDisableUntil[M ~int](label *widget.Label, b binding.Int, match M) *widget.Label {
	update := func() {
		if M(Unwrap(b)) == match {
			label.Importance = widget.MediumImportance
		} else {
			label.Importance = widget.LowImportance
		}
		label.Refresh()
	}
	b.AddListener(binding.NewDataListener(update))
	return label
}

func ColorWhen[M ~int](text *canvas.Text, b binding.Int, match M) *canvas.Text {
	baseColor := text.Color
	update := func() {
		if M(Unwrap(b)) == match {
			text.Color = baseColor
		} else {
			text.Color = zinc400
		}
		canvas.Refresh(text)
	}
	update()
	b.AddListener(binding.NewDataListener(update))
	return text
}

func ColorWithin[M ~int](text *canvas.Text, b binding.Int, from, to M) *canvas.Text {
	baseColor := text.Color
	update := func() {
		val := Unwrap(b)
		if val >= int(from) && val < int(to) {
			text.Color = baseColor
		} else {
			text.Color = zinc400
		}
		canvas.Refresh(text)
	}
	update()
	b.AddListener(binding.NewDataListener(update))
	return text
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

func GreyBox(objects fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(zinc100)
	background.CornerRadius = 8
	background.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	background.StrokeWidth = 1
	return container.NewStack(background, objects)
}

func StackOfSize(size fyne.Size, objects ...fyne.CanvasObject) fyne.CanvasObject {
	box := canvas.NewRectangle(color.Transparent)
	box.SetMinSize(size)
	objects = append(objects, box)
	return container.NewStack(objects...)
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
		txt, err := ErrMsg.Get()
		if err != nil {
			panic(err)
		}
		text.Text = txt
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
func DepShowWindowDialog(
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
