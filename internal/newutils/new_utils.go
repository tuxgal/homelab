package newutils

// Ptr returns a pointer to the provided value.
func Ptr[T any](v T) *T {
	return &v
}

// Nil returns a nil pointer of the specified type.
func Nil[T any]() *T {
	return nil
}
