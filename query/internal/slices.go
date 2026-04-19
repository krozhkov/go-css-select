package internal

func FilterFunc[S ~[]E, E any](s S, f func(E) bool) S {
	r := make(S, 0, len(s))

	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}

	return r
}

func MapFunc[S ~[]E, E any, R any](s S, f func(E) R) []R {
	r := make([]R, len(s))

	for i, v := range s {
		r[i] = f(v)
	}

	return r
}

func Every[S ~[]E, E any](s S, f func(E) bool) bool {
	for _, v := range s {
		if !f(v) {
			return false
		}
	}
	return true
}
