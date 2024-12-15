package pointer

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns v if p is nil, otherwise it returns *p.
// `v` means default value for T.
func Deref[T any](p *T, v T) T {
	if p == nil {
		return v
	}
	return *p
}
