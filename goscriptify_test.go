package goscriptify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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

	os.Remove(dst)

	Convey("Should only allow .go filenames for source", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15")
		err := Build(dst, src)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "go")
		So(err.Error(), ShouldContainSubstring, "ext")
	})

	Convey("Should return build error information", t, func() {
		src := filepath.Join("_test", "fixtures", "synerr.go")
		err := Build(dst, src)
		buildErr, ok := err.(*BuildError)
		So(ok, ShouldBeTrue)
		So(buildErr.Exit, ShouldNotEqual, 0)
		So(buildErr.Error(), ShouldContainSubstring, "syntax error")
	})
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
	Convey("Should return exit status", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15.bash")
		exit, err := RunExec(e, []string{},
			nil, ioutil.Discard, ioutil.Discard)
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)
	})

	Convey("Should pass args to the bin", t, func() {
		e := filepath.Join("_test", "fixtures", "exitarg.bash")
		exit, err := RunExec(e, []string{"25"},
			nil, ioutil.Discard, ioutil.Discard)
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 25)
	})

	Convey("Should pipe stdout and stderr", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15.bash")
		var stdo bytes.Buffer
		var stde bytes.Buffer
		RunExec(e, []string{}, nil, &stdo, &stde)
		So(stdo.String(), ShouldEqual, "STDOUT: Exiting 15")
		So(stde.String(), ShouldEqual, "STDERR: Exiting 15")
	})

	Convey("Should pipe stdin", t, func() {
		e := filepath.Join("_test", "fixtures", "echoinput.bash")
		var i bytes.Buffer
		var o bytes.Buffer
		fmt.Fprint(&i, "Writing to STDIN")
		RunExec(e, []string{}, &i, &o, ioutil.Discard)
		So(o.String(), ShouldEqual, "Echoing Writing to STDIN")
	})
}

func TestRunScriptWithOpts(t *testing.T) {
	dst := filepath.Join("_test", "tmp")

	Convey("Should run a .go file", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15.go")
		opts := ScriptOptions{dst, nil, ioutil.Discard, ioutil.Discard}
		exit, err := RunScriptWithOpts(e, []string{}, opts)
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)
	})

	Convey("Should run a no-ext go file", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15")
		opts := ScriptOptions{dst, nil, ioutil.Discard, ioutil.Discard}
		exit, err := RunScriptWithOpts(e, []string{}, opts)
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)
	})

	Convey("Should create all dest dirs if they don't exist", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15")
		nestedDstRoot := filepath.Join(dst, "nested")
		nestedDst := filepath.Join(nestedDstRoot, "dirs")

		// Just to be safe, remove the dir ahead of time
		os.RemoveAll(nestedDstRoot)

		opts := ScriptOptions{nestedDst, nil, ioutil.Discard, ioutil.Discard}
		RunScriptWithOpts(src, []string{}, opts)

		// Now check to make sure the nestedDst exists
		_, err := os.Stat(nestedDst)
		So(err, ShouldBeNil)
	})
}
