package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*topLevel)(nil)

type topLevel struct {
	left, right fyne.CanvasObject
}

func NewTopLevel(left, right fyne.CanvasObject) *fyne.Container {
	all := []fyne.CanvasObject{left, right}

	return container.New(newTopLayout(left, right), all...)
}

// NewBorderLayout creates a new BorderLayout instance with top, bottom, left
// and right objects set. All other items in the container will fill the remaining space in the middle.
// Multiple extra items will be stacked in the specified order as a Stack container.
func newTopLayout(left, right fyne.CanvasObject) fyne.Layout {
	return &topLevel{left, right}
}

// Layout is called to pack all child objects into a specified size.
// For BorderLayout this arranges the top, bottom, left and right widgets at
// the sides and any remaining widgets are maximised in the middle space.
func (b *topLevel) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	var topSize, bottomSize, leftSize, rightSize fyne.Size
	if b.left != nil && b.left.Visible() {
		leftWidth := b.left.MinSize().Width
		b.left.Resize(fyne.NewSize(leftWidth, size.Height-topSize.Height-bottomSize.Height))
		b.left.Move(fyne.NewPos(0, topSize.Height))
		leftSize = fyne.NewSize(leftWidth, size.Height-topSize.Height-bottomSize.Height)
	}
	if b.right != nil && b.right.Visible() {
		rightWidth := b.right.MinSize().Width
		b.right.Resize(fyne.NewSize(rightWidth, size.Height-topSize.Height-bottomSize.Height))
		b.right.Move(fyne.NewPos(size.Width-rightWidth, topSize.Height))
		rightSize = fyne.NewSize(rightWidth, size.Height-topSize.Height-bottomSize.Height)
	}

	middleSize := fyne.NewSize(size.Width-leftSize.Width-rightSize.Width, size.Height-topSize.Height-bottomSize.Height)
	middlePos := fyne.NewPos(leftSize.Width, topSize.Height)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if child != b.left && child != b.right {
			child.Resize(middleSize)
			child.Move(middlePos)
		}
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For BorderLayout this is determined by the MinSize height of the top and
// plus the MinSize width of the left and right, plus any padding needed.
// This is then added to the union of the MinSize for any remaining content.
func (b *topLevel) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if child != b.left && child != b.right {
			minSize = minSize.Max(child.MinSize())
		}
	}

	padding := theme.Padding()

	if b.left != nil && b.left.Visible() {
		leftMin := b.left.MinSize()
		minHeight := fyne.Max(minSize.Height, leftMin.Height)
		minSize = fyne.NewSize(minSize.Width+leftMin.Width+padding, minHeight)
	}
	if b.right != nil && b.right.Visible() {
		rightMin := b.right.MinSize()
		minHeight := fyne.Max(minSize.Height, rightMin.Height)
		minSize = fyne.NewSize(minSize.Width+rightMin.Width+padding, minHeight)
	}

	return minSize
}
