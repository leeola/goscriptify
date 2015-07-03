package goscriptify

import (
	"fmt"
	"log"
	"os"
)

// RunScript copies, compiles, and runs the given script with global
// $args, and default options - then Exits the process.
//
// IMPORTANT: This exits the process, captures Stdin, and prints to
// Stdout and Stderr as needed.
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

// RunOneScript will run the first given script that is found. Basically
// a shorthand for FindScript and RunScript
//
// IMPORTANT: This exits the process, captures Stdin, and prints to
// Stdout and Stderr as needed.
func RunOneScript(scripts ...string) {
	s, err := FindScript(scripts)
	if err != nil {
		log.Fatal("Fatal:", err.Error())
	}
	RunScript(s)
}
