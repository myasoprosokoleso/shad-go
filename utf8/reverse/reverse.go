//go:build !solution

package reverse

import (
	"strings"
	"unicode/utf8"
)

func Reverse(input string) string {
	invalidBytes := 0

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		// special output for invalid decoding
		if r == utf8.RuneError && size == 1 {
			invalidBytes++
		}

		i += size
	}

	var b strings.Builder
	// 1 invalid byte -> 3 bytes because of utf8.RuneError (U+FFFD)
	b.Grow(len(input) + 2*invalidBytes)

	for i := len(input); i > 0; {
		r, size := utf8.DecodeLastRuneInString(input[:i])
		// same as in utf8.DecodeRuneInString
		if r == utf8.RuneError && size == 1 {
			b.WriteRune(utf8.RuneError)
		} else {
			b.WriteRune(r)
		}

		i -= size
	}

	return b.String()
}
