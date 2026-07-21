//go:build !solution

package genericsum

import (
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func SortSlice[T constraints.Ordered](a []T) {
	slices.Sort(a)
}

func MapsEqual[K, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for aKey, aVal := range a {
		if bVal, ok := b[aKey]; !ok || bVal != aVal {
			return false
		}
	}
	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	return slices.Contains(s, v)
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup
	for _, ch := range chs {
		wg.Go(func() {
			for v := range ch {
				out <- v
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

type Number interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

func IsHermitianMatrix[T Number](m [][]T) bool {
	dim := len(m)
	if dim == 0 {
		return true
	}

	for i := range dim {
		if len(m[i]) != dim {
			return false
		}

		for j := i; j < dim; j++ {
			if m[i][j] != complexConjugate(m[j][i]) {
				return false
			}
		}
	}

	return true
}

func complexConjugate[T Number](z T) T {
	switch v := any(z).(type) {
	case complex64:
		return any(complex(real(v), -imag(v))).(T)
	case complex128:
		return any(complex(real(v), -imag(v))).(T)
	default:
		return z
	}
}
