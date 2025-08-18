package fynext

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PercStatus is a custom widget that displays a percentage,
// or a check/cross mark for 100%/0%.
type PercStatus struct {
	widget.BaseWidget
	perc binding.Float
}

func NewPercStatus(perc binding.Float) *PercStatus {
	p := &PercStatus{perc: perc}
	p.ExtendBaseWidget(p)

	perc.AddListener(binding.NewDataListener(p.Refresh))

	return p
}

func (p *PercStatus) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel("")
	icon := canvas.NewImageFromResource(nil)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))

	content := container.NewStack(label, icon)

	renderer := &percStatusRenderer{
		status:  p,
		label:   label,
		icon:    icon,
		objects: []fyne.CanvasObject{content},
	}
	renderer.Refresh()
	return renderer
}

type percStatusRenderer struct {
	status  *PercStatus
	label   *widget.Label
	icon    *canvas.Image
	objects []fyne.CanvasObject
}

func (r *percStatusRenderer) MinSize() fyne.Size {
	return r.label.MinSize().Max(r.icon.MinSize())
}

func (r *percStatusRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *percStatusRenderer) Refresh() {
	perc, err := r.status.perc.Get()
	if err != nil {
		panic(err)
	}
	if perc == 1.0 {
		r.label.Hide()
		r.icon.Resource = theme.CheckButtonCheckedIcon()
		r.icon.Show()
	} else if perc == 0.0 {
		r.label.Hide()
		r.icon.Resource = theme.ContentClearIcon()
		r.icon.Show()
	} else {
		r.icon.Hide()
		r.label.SetText(fmt.Sprintf("%.0f%%", perc*100))
		r.label.Show()
	}
	r.label.Refresh()
	r.icon.Refresh()
	canvas.Refresh(r.status)
}

func (r *percStatusRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *percStatusRenderer) Destroy() {}
