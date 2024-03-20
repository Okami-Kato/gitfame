package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	placeholderAuthor    = "%an"
	placeholderCommitter = "%cn"
)

const (
	flagPrettyFormat    = "--pretty=%%H|%s"
	flagNoPager         = "--no-pager"
	flagOneCommit       = "-n 1"
	revAndPathSeparator = "--"
)

func GetLastCommit(repository, revision, path string, useCommitter bool) (commit, actor string, err error) {
	readPipe := func(r io.Reader) error {
		s := bufio.NewScanner(r)
		if !s.Scan() {
			return errors.New("empty stdout")
		}
		line := s.Text()
		tokens := strings.SplitN(line, "|", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("unexpected amount of tokens: %s", line)
		}
		commit = tokens[0]
		actor = tokens[1]
		return s.Err()
	}
	actorPlaceholder := placeholderAuthor
	if useCommitter {
		actorPlaceholder = placeholderCommitter
	}
	flagPretty := fmt.Sprintf(flagPrettyFormat, actorPlaceholder)
	err = execute(readPipe, &executionRequest{
		directory: repository,
		program:   "git",
		flags:     []string{flagNoPager, "log", flagOneCommit, flagPretty, revision, revAndPathSeparator, path},
	})
	return commit, actor, err
}
