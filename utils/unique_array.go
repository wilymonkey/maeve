package utils

type CircSlice[T any] struct {
	data []T
	head int
	tail int
	size int
	cap  int
}

func NewCircSlice[T any](capacity int) *CircSlice[T] {
	return &CircSlice[T]{
		data: make([]T, capacity),
		cap:  capacity,
	}
}

// Push to the back. Does not push if slice if full.
func (cs *CircSlice[T]) Push(item T) bool {
	if cs.size == cs.cap {
		return false
	}
	cs.data[cs.tail] = item
	cs.tail = (cs.tail + 1) % cs.cap
	cs.size++
	return true
}

// Pop from the front.
func (cs *CircSlice[T]) Pop() (T, bool) {
	var empty T
	if cs.size == 0 {
		return empty, false
	}
	item := cs.data[cs.head]
	cs.head = (cs.head + 1) % cs.cap
	cs.size--
	return item, true
}

func (cs *CircSlice[T]) Len() int {
	return cs.size
}

func (cs *CircSlice[T]) ForEach(fn func(item T)) {
	if cs.size == 0 {
		return
	}

	for i := range cs.size {
		index := (cs.head + i) % cs.cap
		fn(cs.data[index])
	}
}

type UniqueCircSlice[T comparable] struct {
	slice   *CircSlice[T]
	indices map[T]int
}

func NewUniqueCircSlice[T comparable](capacity int) *UniqueCircSlice[T] {
	return &UniqueCircSlice[T]{
		slice:   NewCircSlice[T](capacity),
		indices: make(map[T]int, capacity),
	}
}

// Updates in place or pushes to the slice.
func (usc *UniqueCircSlice[T]) Push(item T) {
	if i, exists := usc.indices[item]; exists {
		usc.slice.data[i] = item
		return
	}
	if usc.slice.size == usc.slice.cap {
		removed, _ := usc.slice.Pop()
		delete(usc.indices, removed)
	}
	usc.slice.Push(item)
	usc.indices[item] = usc.slice.tail
}

func (ucs *UniqueCircSlice[T]) ForEach(fn func(item T)) {
	ucs.slice.ForEach(fn)
}
