//
// # GoScriptify
//
package goscriptify

import (
	"errors"
	"fmt"
	"io"
	"log"
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
func FindScript(ps []string) (string, error) {
	var exists, isDir bool
	for _, p := range ps {
		// Not checking for error here, because we only care about finding
		// a valid, readable, script. If anything stops that (permissions/etc)
		// we don't care - it may as well not exist.
		exists, isDir, _ = utils.Exists(p)
		if exists && !isDir {
			return p, nil
		}
	}
	return "", errors.New(fmt.Sprint("Cannot find", ps))
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
		if exists, _, _ := utils.Exists(paths[i].Generated); exists {
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

// A shorthand for FindScript and RunScript
func RunOneScript(scripts ...string) {
	s, err := FindScript(scripts)
	if err != nil {
		log.Fatal("Fatal:", err.Error())
	}
	RunScript(s)
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
