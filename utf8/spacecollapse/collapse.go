//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
)

func CollapseSpaces(input string) string {
	var b strings.Builder
	b.Grow(len(input))

	var hadSpace bool
	for _, r := range input {
		if !hadSpace && unicode.IsSpace(r) {
			hadSpace = true
			b.WriteRune(' ')
		} else if !unicode.IsSpace(r) {
			b.WriteRune(r)
			hadSpace = false
		}
	}

	return b.String()
}
