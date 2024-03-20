package sort

import (
	"errors"
	"sort"

	"github.com/Okami-Kato/gitfame/internal/domain"
)

type Key string

const (
	Lines   Key = "lines"
	Commits Key = "commits"
	Files   Key = "files"
)

var keySelectors = map[Key]func(*domain.FameEntry) int{
	Lines:   func(e *domain.FameEntry) int { return e.Lines },
	Commits: func(e *domain.FameEntry) int { return e.Commits },
	Files:   func(e *domain.FameEntry) int { return e.Files },
}

type CompositeKey []Key

var defaultCompositeKey = CompositeKey{Lines, Commits, Files}

var ErrUnsupportedKey = errors.New("unsupported key")

func ToCompositeKey(orderBy Key) (CompositeKey, error) {
	compositeKey := make(CompositeKey, 0, len(defaultCompositeKey))
	compositeKey = append(compositeKey, orderBy)
	for _, key := range defaultCompositeKey {
		if key == orderBy {
			continue
		}
		compositeKey = append(compositeKey, key)
	}
	if len(compositeKey) > len(defaultCompositeKey) {
		return nil, ErrUnsupportedKey
	}
	return compositeKey, nil
}

func SortFameEntries(arr []domain.FameEntry, orderBy CompositeKey) {
	sort.Slice(arr, func(i, j int) bool {
		for _, key := range orderBy {
			valueI := keySelectors[key](&arr[i])
			valueJ := keySelectors[key](&arr[j])
			if valueI == valueJ {
				continue
			}
			return valueI > valueJ
		}
		return arr[i].Name < arr[j].Name
	})
}
