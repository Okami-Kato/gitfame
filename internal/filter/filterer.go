package filter

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Filterer interface {
	Filter([]string) ([]string, error)
}

type FiltererMode int

const (
	BlackList FiltererMode = iota
	WhiteList
)

type PathPatternFilterer struct {
	patterns []string
	mode     FiltererMode
}

func (f *PathPatternFilterer) Filter(paths []string) (filtered []string, err error) {
Outer:
	for _, path := range paths {
		for _, pattern := range f.patterns {
			if ok, err := filepath.Match(pattern, path); err != nil {
				return filtered, fmt.Errorf("error while applying exclude pattern [%s] to [%s]: %w", pattern, path, err)
			} else if ok {
				if f.mode == WhiteList {
					filtered = append(filtered, path)
				}
				continue Outer
			}
		}
		if f.mode == BlackList {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

func NewPathPatternFilterer(patterns []string, mode FiltererMode) Filterer {
	return &PathPatternFilterer{patterns, mode}
}

type PathSuffixFilterer struct {
	suffix []string
}

func (f *PathSuffixFilterer) Filter(paths []string) (filtered []string, err error) {
Outer:
	for _, path := range paths {
		for _, suffix := range f.suffix {
			if strings.HasSuffix(path, suffix) {
				filtered = append(filtered, path)
				continue Outer
			}
		}
	}
	return filtered, nil
}

func NewPathSuffixFilterer(suffexes []string) Filterer {
	return &PathSuffixFilterer{suffexes}
}
