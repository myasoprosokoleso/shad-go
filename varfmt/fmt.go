//go:build !solution

package varfmt

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func Sprintf(format string, args ...interface{}) string {
	var (
		b              strings.Builder
		placeholderIdx int
	)
	strArgs := make(map[int]string, len(args))

	n := len(format)
	b.Grow(n)
	for i := 0; i < n; {
		if format[i] != '{' {
			b.WriteByte(format[i])
			i++
			continue
		}

		numStart := i + 1
		for format[i+1] != '}' {
			i++
		}
		numEnd := i + 1

		var argsIdx int
		if numStart == numEnd {
			argsIdx = placeholderIdx
		} else {
			num, err := strconv.Atoi(format[numStart:numEnd])
			if err != nil {
				log.Fatal(err)
			}
			argsIdx = num
		}

		arg, ok := strArgs[argsIdx]
		if !ok {
			arg = fmt.Sprint(args[argsIdx])
			strArgs[argsIdx] = arg
		}

		if _, err := b.WriteString(arg); err != nil {
			log.Fatal(err)
		}
		placeholderIdx++
		i = numEnd + 1
	}

	return b.String()
}
