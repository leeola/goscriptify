package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path/filepath"
	"testing"
)

func TestBuild(t *testing.T) {
	dst := filepath.Join("_test", "tmp", "bin")
	os.Remove(dst)

	Convey("Should build the source to the target location", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15.go")
		err := Build(dst, src)
		So(err, ShouldBeNil)
		_, err = os.Stat(dst)
		So(err, ShouldBeNil)
	})

	Convey("Should only allow .go filenames for source", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15")
		err := Build(dst, src)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "go")
		So(err.Error(), ShouldContainSubstring, "ext")
	})

	os.Remove(dst)
}

// This is hard to test, due to GetTempPaths using filepath.Abs()
// So for now we're just checking the format.
func TestGetTempPaths(t *testing.T) {
	Convey("Should return the bin destination", t, func() {
		b, _, err := GetTempPaths("foo", "bar")
		So(err, ShouldBeNil)
		So(b, ShouldStartWith, "bar")
	})

	Convey("Should return the source destination", t, func() {
		_, s, err := GetTempPaths("foo", "bar")
		So(err, ShouldBeNil)
		So(s, ShouldStartWith, "bar")
		So(s, ShouldEndWith, ".go")
	})
}

func TestRunExec(t *testing.T) {
	Convey("Should return exit status", t, nil)
	Convey("Should pipe stdout and stderr", t, nil)
}

func TestRunScriptWithOpts(t *testing.T) {
	Convey("Should copy the source to absolute hashed", t, nil)
}
