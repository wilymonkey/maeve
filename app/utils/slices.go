package utils

func SliceToSet[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		set[v] = struct{}{}
	}
	return set
}

func SetToSlice[T comparable](set map[T]struct{}) []T {
	slice := make([]T, 0, len(set))
	for k := range set {
		slice = append(slice, k)
	}
	return slice
}
