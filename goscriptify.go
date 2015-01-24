//
// # GoScriptify
//
package goscriptify

import (
	"bytes"
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

type BuildError struct {
	Exit    int
	Message string
}

func (e *BuildError) Error() string {
	return fmt.Sprintf("Go build error:\n\n%s", e.Message)
}

// Build the given source to the destination
//
// Currently just using the go runtime to build, for simplicity.
func Build(dst string, srcs []string) error {
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

// Go through a slice of ScriptPaths removing all ScriptPath.Generated
// from the file system if their ScriptPath.Clean is true.
func CleanScripts(ps []ScriptPath) error {
	for _, sPath := range ps {
		if sPath.Clean {
			err := os.Remove(sPath.Generated)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Copy a slice of ScriptPaths from their ScriptPath.Original location
// to the ScriptPath.Generated location.
func CopyScripts(ps []ScriptPath) (err error) {
	for _, sPath := range ps {
		// If they're the same, no need to copy.
		if sPath.Original == sPath.Generated {
			continue
		}
		err = utils.CopyFile(sPath.Generated, sPath.Original)
		if err != nil {
			// We should automatically clean scripts up in the future,
			// i'm just not decided on where this should take place - in the
			// api.
			//CleanScripts(ps[0:i])
			break
		}
	}
	return err
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

type ScriptPath struct {
	Original  string
	Generated string
	Clean     bool
}

// When given a script path and a base destination directory,
// return the formatted temporary paths.
//
// The first (dst) path is the binary (executable) path
func GetPaths(sources []string, temp string) (string, []ScriptPath,
	error) {
	if len(sources) == 0 {
		return "", []ScriptPath{}, errors.New("A source file is required")
	}
	paths := make([]ScriptPath, len(sources))

	// To get a unique "id" of this build, we're combining the abs path
	// of this directory, and all source names, and then hashing it.
	dir, err := os.Getwd()
	if err != nil {
		return "", []ScriptPath{}, err
	}
	h := utils.HashString(strings.Join(append(sources, dir), ""))

	// Get the hashed bin path. Eg: /tmp/goscriptify/ads7s6adada8asdka
	binDst := filepath.Join(temp, h)

	// Loop through all of the source files and generate go build friendly
	// path names as needed.
	for i, source := range sources {
		paths[i] = ScriptPath{Original: source}
		// If the source already ends in .go, no need to do anything
		if filepath.Ext(source) == ".go" {
			paths[i].Generated = source
			paths[i].Clean = false
			continue
		}

		// append .go
		paths[i].Generated = fmt.Sprintf("%s.go", source)
		paths[i].Clean = true

		// If the source.go file exists, we can't replace it. So, choose
		// an alternate, long and ugly name.
		if exists, _ := utils.Exists(paths[i].Generated); exists {
			d := filepath.Dir(source)
			f := filepath.Base(source)
			// Note that we're not checking if this exists currently.
			// Living life on the edge of our seat i guess?
			paths[i].Generated = filepath.Join(d, fmt.Sprintf("%s-%s.go", h, f))
		}
	}

	return binDst, paths, nil
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
	exit, err := RunScriptsWithOpts([]string{p}, os.Args[1:], opts)
	if err != nil {
		if builderr, ok := err.(*BuildError); ok {
			fmt.Fprint(os.Stderr, builderr.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Fatal: %s", err.Error())
		}

		if exit == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exit)
}

// Copy, compile, and run the given script with the given options.
//
// Returns the exit status and any encountered errors
func RunScriptsWithOpts(scripts, args []string,
	opts ScriptOptions) (int, error) {
	binDst, scriptPaths, err := GetPaths(scripts, opts.Temp)
	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(opts.Temp, 0777)
	if err != nil {
		return 0, err
	}

	err = CopyScripts(scriptPaths)
	if err != nil {
		return 0, err
	}

	// Make a slice of sources for the build command
	srcs := make([]string, len(scriptPaths))
	for i, s := range scriptPaths {
		srcs[i] = s.Generated
	}

	// In the future we will checksum the source(s), but for now we're
	// just letting go handle the repeat build caching (if at all)
	err = Build(binDst, srcs)
	if err != nil {
		return 0, err
	}

	// Now cleanup any script mess we made.
	err = CleanScripts(scriptPaths)
	if err != nil {
		return 0, err
	}

	return RunExec(binDst, args, os.Stdin, os.Stdout, os.Stderr)
}
