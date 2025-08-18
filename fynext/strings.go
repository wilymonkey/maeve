package fynext

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TruncLabel(label *widget.Label, width float32) {
	getWidth := func(s string) float32 {
		return fyne.MeasureText(s, theme.TextSize(), label.TextStyle).Width
	}

	txt := label.Text
	txtWidth := getWidth(txt)
	if txtWidth <= width {
		return
	}

	ellipsis := "…"
	ellipsisWidth := getWidth(ellipsis)
	availWidth := width - ellipsisWidth
	if availWidth <= 0 {
		label.SetText(ellipsis)
		return
	}

	avgCharWidth := txtWidth / float32(len(txt))
	estChars := int(availWidth / avgCharWidth)
	if estChars <= 0 {
		label.SetText(ellipsis)
		return
	}
	if estChars > len(txt) {
		return
	}
	start := len(txt) - estChars
	label.SetText(ellipsis + txt[start:])
}
