//
// # GoScriptify
//
package goscriptify

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/leeola/goscriptify/utils"
)

type ScriptOptions struct {
	Temp   string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Build the given source to the destination
//
// Currently just using the go runtime to build, for simplicity.
func Build(dst, src string) error {
	if filepath.Ext(src) != ".go" {
		return errors.New("source must have go extension")
	}

	_, err := os.Stat(src)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", dst, src)
	return cmd.Run()
}

// Find a single file from a list of multiple files, and returning
// the first found filename. This is to support multiple name types,
// or cases.
//
// Example:
//
//    FindScript([]string{"Builder", "builder"})
//
// Will search for both "Builder" and "builder", and return the first
// found file.
//
// NOTE: Due to OSX's case insensitivity, it's hard (maybe possible?)
// to know the *actual* filename of the found file. Tests, then
// have to ignore the string output of this function, as it will
// fail on OSX. I'd love to see a workaround for this issue.
func FindScript(ss [][]string) (string, error) {
	return "", errors.New("Not implemented")
}

// When given a script path and a base destination directory,
// return the formatted temporary paths.
//
// The first (dst) path is the binary (executable) path
// The second (dstSrc) path is the source script to compile to the bin
func GetTempPaths(script, temp string) (string, string, error) {
	// Get the full path of src
	src, err := filepath.Abs(script)
	if err != nil {
		return "", "", err
	}

	// Get the hashed bin path. Eg: /tmp/goscriptify/ads7s6adada8asdka
	dst := filepath.Join(temp, utils.HashString(src))
	// Set src to dst.go ext
	dstSrc := strings.Join([]string{dst, ".go"}, "")
	return dst, dstSrc, nil
}

// Run the given path as an executable, with the supplied args, and
// forwarding the stdin/out/err.
//
// Return the exit status, and any errors encountered.
func RunExec(p string, args []string,
	stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	if _, err := os.Stat(p); err != nil {
		return 0, err
	}

	cmd := exec.Command(p, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			} else {
				return 0, err
			}
		} else {
			return 0, err
		}
	}

	return 0, nil
}

// Copy, compile, and run the given script with global $args, and
// default options.
func RunScript(p string) {
	opts := ScriptOptions{
		"/tmp/goscriptify",
		os.Stdin, os.Stdout, os.Stderr,
	}
	exit, err := RunScriptWithOpts(p, os.Args[1:], opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %s", err.Error())
		if exit == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exit)
}

// Copy, compile, and run the given script with the given options.
//
// Returns the exit status and any encountered errors
func RunScriptWithOpts(p string, args []string,
	opts ScriptOptions) (int, error) {
	dst, dstSrc, err := GetTempPaths(p, opts.Temp)
	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(opts.Temp, 0777)
	if err != nil {
		return 0, err
	}

	err = utils.CopyFile(dstSrc, p)
	if err != nil {
		return 0, err
	}

	// In the future we will checksum the source(s), but for now we're
	// just letting go handle the repeat build caching (if at all)
	err = Build(dst, dstSrc)
	if err != nil {
		return 0, err
	}

	return RunExec(dst, args, os.Stdin, os.Stdout, os.Stderr)
}
