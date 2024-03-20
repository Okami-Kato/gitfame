package configs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed language_extensions.json
var languageExtensionsRaw []byte

type languageInfo struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

type ErrUnsupportedLanguages struct {
	languages []string
}

func (e *ErrUnsupportedLanguages) Error() string {
	return fmt.Sprintf("unsupported languages: %v", e.languages)
}

func GetExtensions(languages ...string) (extensions []string, err error) {
	langSet := make(map[string]struct{})
	for _, lang := range languages {
		langSet[strings.ToLower(lang)] = struct{}{}
	}

	var arr []languageInfo
	err = json.Unmarshal(languageExtensionsRaw, &arr)
	if err != nil {
		return extensions, fmt.Errorf("error unmarshalling language extensions: %w", err)
	}

	for _, langInfo := range arr {
		if _, ok := langSet[strings.ToLower(langInfo.Name)]; ok {
			extensions = append(extensions, langInfo.Extensions...)
			delete(langSet, strings.ToLower(langInfo.Name))
		}
	}
	if len(langSet) > 0 {
		leftLanguages := make([]string, 0, len(langSet))
		for lang := range langSet {
			leftLanguages = append(leftLanguages, lang)
		}
		return extensions, &ErrUnsupportedLanguages{leftLanguages}
	}
	return extensions, err
}
