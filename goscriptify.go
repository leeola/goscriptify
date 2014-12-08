//
// # GoScriptify
//
package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

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
	return "", nil
}

func main() {}

// Execute the given file as go code.
func Run(p string) (int, error) {
	return 0, nil
}
