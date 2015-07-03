package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "STDOUT: Exiting 15")
	fmt.Fprint(os.Stderr, "STDERR: Exiting 15")
	os.Exit(15)
}

// vim: set filetype=go:
