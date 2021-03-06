package goscriptify

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

type BuildError struct {
	Exit    int
	Message string
}

func (e *BuildError) Error() string {
	return fmt.Sprintf("Go build error:\n\n%s", e.Message)
}

// BuildDir builds the directory to the destination.
//
// Currently just using the go runtime to build, for simplicity.
func BuildDir(dst string, dir string) error {
	// If the dst is not absolute, make it relative to the cwd.
	// This is needed because setting `cmd.Dir = dir` will cause the output
	// to be relative to the cmd.Dir, not this process Cwd.
	dst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", dst, ".")
	cmd.Dir = dir

	// Go returns build error output on the stderr, so we're storing it
	// in case we need it. If needed, it will be returned inside of the
	// BuildError
	var stderr bytes.Buffer
	defer stderr.Reset()
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return &BuildError{
					Exit:    status.ExitStatus(),
					Message: stderr.String(),
				}
			}
		}
		// If it's not an execerr or we can't get the status, return err
		return err
	}

	return nil
}

// BuildFiles builds the given sources to the destination
//
// Currently just using the go runtime to build, for simplicity.
func BuildFiles(dst string, srcs []string) error {
	// Becuase Go's builder can return some vague errors, lets do some
	// simple sanity checks.
	for _, s := range srcs {
		if e := filepath.Ext(s); e != ".go" {
			return errors.New("source must have go extension")
		}
	}

	args := append([]string{"build", "-o", dst}, srcs...)
	cmd := exec.Command("go", args...)

	// Go returns build error output on the stderr, so we're storing it
	// in case we need it. If needed, it will be returned inside of the
	// BuildError
	var stderr bytes.Buffer
	defer stderr.Reset()
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return &BuildError{
					Exit:    status.ExitStatus(),
					Message: stderr.String(),
				}
			}
		}
		// If it's not an execerr or we can't get the status, return err
		return err
	}

	return nil
}
