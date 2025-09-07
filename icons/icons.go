package icons

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed cancel.svg
var cancelIcon []byte
var CancelSvg = &fyne.StaticResource{
	StaticName:    "cancel.svg",
	StaticContent: cancelIcon,
}

//go:embed check.svg
var checkIcon []byte
var CheckSvg = &fyne.StaticResource{
	StaticName:    "check.svg",
	StaticContent: checkIcon,
}

//go:embed checkbox-checked.svg
var checkboxCheckedIcon []byte
var CheckboxCheckedSvg = &fyne.StaticResource{
	StaticName:    "checkbox-checked.svg",
	StaticContent: checkboxCheckedIcon,
}

//go:embed checkbox.svg
var checkboxIcon []byte
var CheckboxSvg = &fyne.StaticResource{
	StaticName:    "checkbox.svg",
	StaticContent: checkboxIcon,
}

//go:embed trash.svg
var trashIcon []byte
var TrashSvg = &fyne.StaticResource{
	StaticName:    "trash.svg",
	StaticContent: trashIcon,
}

//go:embed pencil.svg
var pencilIcon []byte
var PencilSvg = &fyne.StaticResource{
	StaticName:    "pencil.svg",
	StaticContent: pencilIcon,
}

//go:embed danger.svg
var dangerIcon []byte
var DangerSvg = &fyne.StaticResource{
	StaticName:    "danger.svg",
	StaticContent: dangerIcon,
}

//go:embed clipboard.svg
var clipboardIcon []byte
var ClipboardSvg = &fyne.StaticResource{
	StaticName:    "clipboard.svg",
	StaticContent: clipboardIcon,
}

//go:embed link.svg
var linkIcon []byte
var LinkSvg = &fyne.StaticResource{
	StaticName:    "link.svg",
	StaticContent: linkIcon,
}

//go:embed favicon.ico
var faviconBytes []byte
var FaviconIco = &fyne.StaticResource{
	StaticName:    "favicon.ico",
	StaticContent: faviconBytes,
}

//go:embed favicon.svg
var faviconSvgBytes []byte
var FaviconSvg = &fyne.StaticResource{
	StaticName:    "favicon.svg",
	StaticContent: faviconSvgBytes,
}
