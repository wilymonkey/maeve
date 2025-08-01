package utils

import "fyne.io/fyne/v2/data/binding"

func GetOrPanic[T any](b binding.Item[T]) T {
	t, err := b.Get()
	if err != nil {
		panic(err)
	}
	return t
}
