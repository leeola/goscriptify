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
		err := Build(dst, []string{src})
		So(err, ShouldBeNil)
		_, err = os.Stat(dst)
		So(err, ShouldBeNil)
	})

	os.Remove(dst)

	Convey("Should only allow .go filenames for source", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15")
		err := Build(dst, []string{src})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "go")
		So(err.Error(), ShouldContainSubstring, "ext")
	})

	Convey("Should return build error information", t, func() {
		src := filepath.Join("_test", "fixtures", "synerr.go")
		err := Build(dst, []string{src})
		buildErr, ok := err.(*BuildError)
		So(ok, ShouldBeTrue)
		So(buildErr.Exit, ShouldNotEqual, 0)
		So(buildErr.Error(), ShouldContainSubstring, "syntax error")
	})
}

// This is hard to test, due to GetPaths using os.Getwd()
// So for now we're just checking the format.
func TestGetPaths(t *testing.T) {
	Convey("Should return the bin destination", t, func() {
		s, _, err := GetPaths([]string{"foo"}, "bar")
		So(err, ShouldBeNil)
		So(s, ShouldStartWith, "bar")
	})

	Convey("Should return the source destinations", t, func() {
		_, ps, err := GetPaths([]string{"foo"}, "bar")
		So(err, ShouldBeNil)
		So(ps[0].Generated, ShouldEqual, "foo.go")
	})

	Convey("Should only append .go if it's missing", t, func() {
		_, ps, err := GetPaths([]string{"foo.go"}, "bar")
		So(err, ShouldBeNil)
		So(ps[0].Generated, ShouldEqual, "foo.go")
	})

	Convey("Should choose an alternate filename when "+
		"the source.go already exists", t, func() {
		_, ps, err := GetPaths([]string{"_test/fixtures/exit15"}, "bar")
		So(err, ShouldBeNil)
		// The chosen filename will be hashed, so.. just like before
		// we can't test the exact match. We can test the start and end,
		// and ensure that it does *not* equal the exit15.go path.
		So(ps[0].Generated, ShouldStartWith, "_test/fixtures/")
		So(ps[0].Generated, ShouldEndWith, "exit15.go")
		So(ps[0].Generated, ShouldNotEqual, "_test/fixtures/exit15.go")
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

func TestRunScriptsWithOpts(t *testing.T) {
	dst := filepath.Join("_test", "tmp")

	Convey("Should run a .go file", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15.go")
		opts := ScriptOptions{dst, nil, ioutil.Discard, ioutil.Discard}
		exit, err := RunScriptsWithOpts([]string{e}, []string{}, opts)
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)
	})

	Convey("Should run a no-ext go file", t, func() {
		e := filepath.Join("_test", "fixtures", "exit15")
		opts := ScriptOptions{dst, nil, ioutil.Discard, ioutil.Discard}
		exit, err := RunScriptsWithOpts([]string{e}, []string{}, opts)
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
		RunScriptsWithOpts([]string{src}, []string{}, opts)

		// Now check to make sure the nestedDst exists
		_, err := os.Stat(nestedDst)
		So(err, ShouldBeNil)
	})
}

// TODO: Find a way to make FindScript tests pass on OSX.
// NOTE: Due to OSX's case insensitivity, it's hard (maybe possible?)
// to know the *actual* filename of the found file. Tests, then
// have to ignore the string output of this function, as it will
// fail on OSX. I'd love to see a workaround for this issue.
//
// So, a fail on OSX does not currently mean truly failing tests.
// Run on Linux to confirm.
func TestFindScript(t *testing.T) {
	fixDir := filepath.Join("_test", "fixtures")

	Convey("Should find a given script", t, func() {
		s, err := FindScript([]string{
			filepath.Join(fixDir, "foo"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "foo"))
	})

	Convey("Should find the given subdirectory script", t, func() {
		s, err := FindScript([]string{
			filepath.Join(fixDir, "baz", "bat"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz", "bat"))
	})

	Convey("Should find the first given script", t, func() {
		s, err := FindScript([]string{
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bang", "boom", "flash"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "bar"))

		s, err = FindScript([]string{
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bang", "boom", "flash"),
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz", "bat"))

		s, err = FindScript([]string{
			filepath.Join(fixDir, "bang", "boom", "flash"),
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "bang", "boom", "flash"))
	})
}
