//go:build !solution

package fileleak

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

func VerifyNone(t testingT) {
	before := parseOpenFiles()
	t.Cleanup(func() {
		after := parseOpenFiles()

		var leaked []string
		for fdAfter, pathAfter := range after {
			if pathBefore, ok := before[fdAfter]; !ok || pathBefore != pathAfter {
				leaked = append(leaked, pathAfter)
			}
		}

		assert.Len(t, leaked, 0, "leaked %d files: %v", len(leaked), leaked)
	})
}

func parseOpenFiles() map[string]string {
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return nil
	}

	openFiles := make(map[string]string, len(fds))
	for _, f := range fds {
		path, err := os.Readlink(filepath.Join("/proc/self/fd", f.Name()))
		if err != nil {
			// fd might have been closed between ReadDir and Readlink
			continue
		}
		openFiles[f.Name()] = path // "0" -> "/path/to/file.txt"
	}

	return openFiles
}
