package git

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Blamer interface {
	Blame(string) (map[string]map[string]int, error)
}

type FileBlamer struct {
	repository   string
	revision     string
	useCommitter bool
}

func NewBlamer(repository string, revision string, useCommitter bool) Blamer {
	return &FileBlamer{
		revision:     revision,
		repository:   repository,
		useCommitter: useCommitter,
	}
}

const (
	tokenAuthor    = "author"
	tokenCommitter = "committer"
)

const flagIncremantal = "--incremental"

func (b *FileBlamer) Blame(path string) (map[string]map[string]int, error) {
	stats := make(map[string]map[string]int)
	readPipe := func(r io.Reader) error {
		s := bufio.NewScanner(r)

		var tokenActor string
		if b.useCommitter {
			tokenActor = tokenCommitter
		} else {
			tokenActor = tokenAuthor
		}

		var currentCommit string
		var currentActor string

		for {
			if ok := seekToCommit(s); !ok {
				break
			}
			line := s.Text()
			tokens := strings.Split(line, " ")

			lineCount, err := strconv.Atoi(tokens[3])
			if err != nil {
				return fmt.Errorf("expected integer - instead recieved %s", tokens[3])
			}

			if currentCommit == tokens[0] {
				stats[currentActor][currentCommit] += lineCount
				continue
			}

			currentCommit = tokens[0]

			if ok := seekToToken(s, tokenActor); !ok {
				break
			}
			currentActor = strings.TrimPrefix(s.Text(), tokenActor+" ")

			if stats[currentActor] == nil {
				stats[currentActor] = make(map[string]int)
			}

			stats[currentActor][currentCommit] += lineCount
		}
		return s.Err()
	}
	err := execute(readPipe, &executionRequest{
		directory: b.repository,
		program:   "git",
		flags:     []string{"blame", b.revision, flagIncremantal, path},
	})
	return stats, err
}

func seekToToken(s *bufio.Scanner, token string) bool {
	for s.Scan() {
		line := s.Text()
		tokens := strings.Split(line, " ")
		if len(tokens) > 0 && token == tokens[0] {
			return true
		}
	}
	return false
}

func seekToCommit(s *bufio.Scanner) bool {
	for s.Scan() {
		line := s.Text()
		tokens := strings.Split(line, " ")
		if len(tokens) > 0 && isSHA1(tokens[0]) {
			return true
		}
	}
	return false
}

var sha1Regex = regexp.MustCompile(`^[0-9a-f]{40}$`)

func isSHA1(s string) bool {
	return sha1Regex.MatchString(s)
}
