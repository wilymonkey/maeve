package style

func BoolView(b bool) string {
	if b {
		return ITick
	}
	return ICross
}
