package fynext

import "fyne.io/fyne/v2/data/binding"

// Gets the given field or panics if not available.
func Unwrap[T any](b binding.Item[T]) T {
	t, err := b.Get()
	if err != nil {
		panic(err)
	}
	return t
}

func BindNewErr() binding.Item[error] {
	return binding.NewItem(func(err1, err2 error) bool {
		return err1 == err2
	})
}

func BindNewInt64() binding.Item[int64] {
	return binding.NewItem(func(i1, i2 int64) bool {
		return i1 == i2
	})
}

func BindNewString(init string) binding.String {
	b := binding.NewString()
	b.Set(init)
	return b
}
