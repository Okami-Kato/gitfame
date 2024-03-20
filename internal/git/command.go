package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type executionRequest struct {
	directory string
	program   string
	flags     []string
}

func execute(callback func(io.Reader) error, req *executionRequest) error {
	cmd := exec.Command(req.program, req.flags...)
	cmd.Dir = req.directory

	var stdout io.ReadCloser
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error when acquiring stdout pipe: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error when starting the cmd: %w", err)
	}

	err = callback(bufio.NewReader(stdout))
	if err != nil {
		err = fmt.Errorf("error when parsing stdout: %w", err)
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return fmt.Errorf("%s exited unsuccessfully: %w", req.program, exitErr)
		}
		return fmt.Errorf("error while waiting on the %s: %w", req.program, waitErr)
	}

	return err
}
