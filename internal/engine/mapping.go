package engine

import "github.com/Okami-Kato/gitfame/internal/domain"

func toFameEntries(stats map[string]map[string]int, actorFileCount map[string]int) []domain.FameEntry {
	result := make([]domain.FameEntry, 0, len(stats))

	for actor, m := range stats {
		fileCount := actorFileCount[actor]
		commitCount := len(m)
		lineCount := 0
		for _, v := range m {
			lineCount += v
		}
		result = append(result, domain.FameEntry{
			Name:    actor,
			Files:   fileCount,
			Commits: commitCount,
			Lines:   lineCount,
		})
	}
	return result
}
