package filter

import (
	"path/filepath"
	"strings"
	"sync"

	"gitlab.com/slon/shad-go/gitfame/configs"
)

var (
	once       sync.Once
	langToExts map[string][]string
)

func MatchFlags(file string, extensions, languages, exclude, restrictTo []string) bool {
	if !matchExtensions(file, extensions) {
		return false
	}
	if !matchLanguages(file, languages) {
		return false
	}
	if len(exclude) > 0 && matchPatterns(file, exclude) {
		return false
	}
	if !matchPatterns(file, restrictTo) {
		return false
	}
	return true
}

func matchExtensions(file string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}

	for _, ext := range extensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}

	return false
}

func matchLanguages(file string, languages []string) bool {
	if len(languages) == 0 {
		return true
	}

	once.Do(func() {
		langToExts = configs.GetLanguageMapping()
	})
	if langToExts == nil {
		return true
	}

	extensions := make([]string, 0, len(languages))
	for _, lang := range languages {
		if exts, ok := langToExts[lang]; ok {
			extensions = append(extensions, exts...)
		}
	}

	return matchExtensions(file, extensions)
}

func matchPatterns(file string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	for _, pattern := range patterns {
		// ignore error since patterns are already handled in validateFlags
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
	}

	return false
}
