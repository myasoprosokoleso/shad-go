package configs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed language_extensions.json
var LanguageExtensions []byte

func GetLanguageMapping() map[string][]string {
	type language struct {
		Name       string
		Extensions []string
	}

	var languages []language
	if err := json.Unmarshal(LanguageExtensions, &languages); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal language_extensions.json: %v\n", err)
		return nil
	}

	langToExts := make(map[string][]string)
	for _, lang := range languages {
		langToExts[strings.ToLower(lang.Name)] = lang.Extensions
	}

	if len(langToExts) <= 1 {
		fmt.Fprintf(os.Stderr, "not enough languages parsed from language_extensions.json: %v\n", langToExts)
		return nil
	}

	return langToExts
}
