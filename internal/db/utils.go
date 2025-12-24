package db

import "slices"

func SlicesDiff[A comparable](a, b []A) []A {
	set := make(map[A]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}

	return slices.DeleteFunc(a, func(v A) bool {
		_, found := set[v]
		return found
	})
}
