package git

import (
	"bufio"
	"io"
)

const (
	flagNamesOnly = "--name-only"
	flagRecursive = "-r"
)

func ListFiles(repository, revision string) (paths []string, err error) {
	readPipe := func(r io.Reader) error {
		s := bufio.NewScanner(r)
		for s.Scan() {
			line := s.Text()
			paths = append(paths, line)
		}
		return s.Err()
	}
	err = execute(readPipe, &executionRequest{
		directory: repository,
		program:   "git",
		flags:     []string{"ls-tree", revision, flagNamesOnly, flagRecursive},
	})
	return paths, err
}
