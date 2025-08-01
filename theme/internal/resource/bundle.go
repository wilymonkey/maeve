package resource

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed icons/cancel.svg
var cancelIcon []byte
var CancelSvg = &fyne.StaticResource{
	StaticName:    "cancel.svg",
	StaticContent: cancelIcon,
}

//go:embed icons/check.svg
var checkIcon []byte
var CheckSvg = &fyne.StaticResource{
	StaticName:    "check.svg",
	StaticContent: checkIcon,
}

//go:embed icons/checkbox-checked.svg
var checkboxCheckedIcon []byte
var CheckboxCheckedSvg = &fyne.StaticResource{
	StaticName:    "checkbox-checked.svg",
	StaticContent: checkboxCheckedIcon,
}

//go:embed icons/checkbox.svg
var checkboxIcon []byte
var CheckboxSvg = &fyne.StaticResource{
	StaticName:    "checkbox.svg",
	StaticContent: checkboxIcon,
}

//go:embed icons/favicon.svg
var faviconIcon []byte
var FaviconSvg = &fyne.StaticResource{
	StaticName:    "favicon.svg",
	StaticContent: faviconIcon,
}

//go:embed icons/trash.svg
var trashIcon []byte
var TrashSvg = &fyne.StaticResource{
	StaticName:    "trash.svg",
	StaticContent: trashIcon,
}
