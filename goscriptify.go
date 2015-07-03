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

func NewScriptPath(h, p string) ScriptPath {
	sp := ScriptPath{Original: p}
	// If the source already ends in .go, no need to do anything
	if filepath.Ext(p) == ".go" {
		sp.Generated = p
		// !IMPORTANT! Don't delete pre-existing go files.
		// Lets not be jerks please?
		sp.Clean = false
		return sp
	}

	// append .go
	sp.Generated = fmt.Sprintf("%s.go", p)
	sp.Clean = true

	// If the source.go file exists, we can't replace it. So, choose
	// an alternate, long and ugly name.
	if exists, _, _ := utils.Exists(sp.Generated); exists {
		d := filepath.Dir(p)
		f := filepath.Base(p)
		// Note that we're not checking if this exists currently.
		// Living life on the edge of our seat i guess?
		sp.Generated = filepath.Join(d, fmt.Sprintf("%s-%s.go", h, f))
	}

	return sp
}

func NewScriptPaths(hash string, paths []string) []ScriptPath {
	sps := make([]ScriptPath, len(paths))
	for i, p := range paths {
		sps[i] = NewScriptPath(hash, p)
	}
	return sps
}

// TODO: Make Clean readonly via hiding and a read method.
type ScriptPath struct {
	// The original, unmodified script path
	Original string

	// The generated script path, which may contain a hash string to make
	// the file unique (to not replace the Original)
	Generated string

	// Whether or not to remove the file at the end.
	Clean bool
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

// FindScript will find a single file from a list of multiple files,
// and returning the first found filename. This is to support multiple
// name types, or cases.
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

// FindScriptOrDir will find a single file or dir from a list of
// paths.
//
// If useDir is true, the found script's directory will be
// returned. If the script is not in a subdirectory, useDir does
// nothing. This makes FindScriptOrDir functionally the same as
// FindScript, with the bonus of accepting an explicit directory
// as a Script.
//
// Examples:
//
//		// This will return "someDir/Builder" if it exists
// 		FindScriptOrDir([]string{"someDir/Builder", "Builder"}, false)
// 		// This will return "someDir" if "someDir/Builder" exists
// 		FindScriptOrDir([]string{"someDir/Builder", "Builder"}, true)
// 		// Both will return "Builder" if "someDir/Builder" does not
// 		// exist
//
func FindScriptOrDir(ps []string, useDir bool) (path string, isDir bool,
	err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	return findScriptOrDir(cwd, ps, useDir)
}

// findScriptOrDir is the implementation of FindScriptOrDir with a
// testable CWD. The CWD is needed because it treats a path of
// `dir/script.go` differently than `script.go` - and as a result we
// can't pass in the full test path for every script, during test usage.
func findScriptOrDir(cwd string, ps []string, useDir bool) (path string,
	isDir bool, err error) {

	var exists bool
	for _, p := range ps {
		// Not checking for error here, because we only care about finding
		// a valid, readable, script. If anything stops that (permissions/etc)
		// we don't care - it may as well not exist.
		exists, isDir, _ = utils.Exists(filepath.Join(cwd, p))
		if exists {
			// If path is a dir, we can't use its dir. Return it directly.
			if isDir || !useDir {
				return p, isDir, nil
			}

			dir := filepath.Dir(p)

			// If path is in the root (given) dir, return the script
			if dir == "." {
				return p, isDir, nil
			}

			// We know that it's in a subdir (because the dir isn't .) and
			// useDir == true, so make sure that the file extension is .go.
			// If it's not, `go build` will be unable to find the given script
			// because it will be looking for .go files in the directory.
			if filepath.Ext(p) != ".go" {
				return "", false, errors.New("FindScriptOrDir: Subdirectory " +
					"scripts must use the .go extension")
			}

			return dir, true, nil
		}
	}
	return "", false, errors.New(fmt.Sprint("Cannot find", ps))
}

// When given a script path and a base destination directory,
// return the formatted temporary paths.
//
// The first (dst) path is the binary (executable) path
//func GetBinDest(sources []string, temp string) (string, []ScriptPath,
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

// GetBinDest generates a md5 of the source paths, and returns that
// and the md5 it generated.
func GetBinDest(sources []string, temp string) (binDst, hash string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	return getBinDest(cwd, sources, temp)
}

// getBinDest is the cwd testable implementation behind GetBinDest
func getBinDest(cwd string, sources []string, temp string) (binDst, hash string,
	err error) {

	if len(sources) == 0 {
		return "", "", errors.New("GetBinDest: A source file is required")
	}

	// To get a unique "id" of this build, we're combining the abs path
	// of the cwd and all source names, and then hashing it.
	h := utils.HashString(strings.Join(append(sources, cwd), ""))

	// Make the hashed bin path. Eg: /tmp/goscriptify/ads7s6adada8asdka
	binDst = filepath.Join(temp, h)

	return binDst, h, nil
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
	err = BuildFiles(binDst, srcs)
	if err != nil {
		// TODO: Add a defer on cleanup here
		return 0, err
	}

	// Now cleanup any script mess we made.
	err = CleanScripts(scriptPaths)
	if err != nil {
		return 0, err
	}

	// TODO: use opts here, and test for it.
	return RunExec(binDst, args, os.Stdin, os.Stdout, os.Stderr)
}

func RunScriptDirWithOpts(dir string, args []string, opts ScriptOptions) (int, error) {
	binDst, _, err := GetBinDest([]string{dir}, opts.Temp)
	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(opts.Temp, 0777)
	if err != nil {
		return 0, err
	}

	// In the future we will checksum the source(s), but for now we're
	// just letting go handle the repeat build caching (if at all)
	err = BuildDir(binDst, dir)
	if err != nil {
		return 0, err
	}

	return RunExec(binDst, args, opts.Stdin, opts.Stdout, opts.Stderr)
}
