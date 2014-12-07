//
// # GoScriptify
//
package main

import ()

func Build(dst, src string) error {
	return nil
}

func FindScript(s string) (string, error) {
	return "", nil
}

// Find files from one of the acceptable names and return
// the given name.
//
// NOTE: Due to OSX's case insensitivity, it's hard (maybe possible?)
// to know the *actual* filename of the found file. Tests, then
// have to ignore the string output of this function, as it will
// fail on OSX. I'd love to see a workaround for this issue.
func FindScripts(ss string) {
}

func main() {}

// Execute the given file as go code.
func Run(p string) (int, error) {
	return 0, nil
}
