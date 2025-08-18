package fynext

import "fyne.io/fyne/v2/data/binding"

func GetOrPanic[T any](b binding.Item[T]) T {
	t, err := b.Get()
	if err != nil {
		panic(err)
	}
	return t
}

func NewErrBinding() binding.Item[error] {
	return binding.NewItem(func(err1, err2 error) bool {
		return err1 == err2
	})
}
