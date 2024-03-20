package engine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Okami-Kato/gitfame/configs"
	"github.com/Okami-Kato/gitfame/internal/domain"
	"github.com/Okami-Kato/gitfame/internal/filter"
	"github.com/Okami-Kato/gitfame/internal/git"
	"github.com/Okami-Kato/gitfame/internal/sort"
)

type GitFameEngine struct {
	repository     string
	revision       string
	useCommitter   bool
	filterChain    *filter.Chain
	orderBy        sort.CompositeKey
	parallelFactor int
}

type CreationRequest struct {
	Repository     string
	Revision       string
	OrderBy        string
	UseCommitter   bool
	Extensions     []string
	Languages      []string
	Exclude        []string
	RestrictTo     []string
	ParallelFactor int
}

func New(req *CreationRequest) (*GitFameEngine, error) {
	filterChain, err := buildFilterChain(req.Extensions, req.Languages, req.Exclude, req.RestrictTo)
	if err != nil {
		return nil, err
	}
	compositeKey, err := sort.ToCompositeKey(sort.Key(req.OrderBy))
	if err != nil {
		return nil, err
	}
	return &GitFameEngine{
		repository:     req.Repository,
		revision:       req.Revision,
		useCommitter:   req.UseCommitter,
		filterChain:    filterChain,
		orderBy:        compositeKey,
		parallelFactor: req.ParallelFactor,
	}, nil
}

func (e *GitFameEngine) Run() ([]domain.FameEntry, error) {
	paths, err := git.ListFiles(e.repository, e.revision)
	if err != nil {
		return nil, fmt.Errorf("error listing git files: %w", err)
	}
	paths, err = e.filterChain.Filter(paths)
	if err != nil {
		return nil, fmt.Errorf("error filtering files: %w", err)
	}
	blamer := git.NewBlamer(e.repository, e.revision, e.useCommitter)
	stats := make(chan map[string]map[string]int, e.parallelFactor)
	errors := make(chan error, 1)
	sema := make(chan struct{}, e.parallelFactor)
	var wg sync.WaitGroup
	for _, path := range paths {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sema <- struct{}{}
			defer func() {
				<-sema
			}()
			var fileStats map[string]map[string]int
			fileStats, err = blamer.Blame(path)
			if err != nil {
				errors <- fmt.Errorf("error blaming %s: %w", path, err)
				return
			}
			if len(fileStats) == 0 {
				var commit, actor string
				commit, actor, err = git.GetLastCommit(e.repository, e.revision, path, e.useCommitter)
				if err != nil {
					errors <- fmt.Errorf("error retrieving last commit of %s: %w", path, err)
					return
				}
				fileStats[actor] = make(map[string]int)
				fileStats[actor][commit] = 0
			}
			stats <- fileStats
		}()
	}
	go func() {
		wg.Wait()
		close(stats)
	}()
	combinedStats := make(map[string]map[string]int)
	actorFileCount := make(map[string]int)
Loop:
	for {
		select {
		case err := <-errors:
			return nil, err
		default:
		}
		select {
		case fileStats, ok := <-stats:
			if !ok {
				break Loop
			}
			for actor := range fileStats {
				actorFileCount[actor]++
			}
			mergeStats(combinedStats, fileStats)
		default:
		}
	}
	fameEntries := toFameEntries(combinedStats, actorFileCount)
	sort.SortFameEntries(fameEntries, e.orderBy)
	return fameEntries, nil
}

func buildFilterChain(extensions, languages, exclude, restrictTo []string) (*filter.Chain, error) {
	var filterers []filter.Filterer
	if len(extensions) > 0 {
		f := filter.NewPathSuffixFilterer(extensions)
		filterers = append(filterers, f)
	}
	if len(languages) > 0 {
		langExtensions, err := configs.GetExtensions(languages...)
		if err != nil {
			var errUnsupportedLang *configs.ErrUnsupportedLanguages
			if !errors.As(err, &errUnsupportedLang) {
				return nil, fmt.Errorf("error retrieving extensions for %v: %w", languages, err)
			}
		}
		f := filter.NewPathSuffixFilterer(langExtensions)
		filterers = append(filterers, f)
	}
	if len(exclude) > 0 {
		f := filter.NewPathPatternFilterer(exclude, filter.BlackList)
		filterers = append(filterers, f)
	}
	if len(restrictTo) > 0 {
		f := filter.NewPathPatternFilterer(restrictTo, filter.WhiteList)
		filterers = append(filterers, f)
	}
	return filter.NewChain(filterers...), nil
}

func mergeStats(target, source map[string]map[string]int) {
	for actor, lineCountSource := range source {
		lineCountTarget, ok := target[actor]
		if !ok {
			lineCountTarget = make(map[string]int)
			target[actor] = lineCountTarget
		}
		for commit, linesSource := range lineCountSource {
			lineCountTarget[commit] += linesSource
		}
	}
}
