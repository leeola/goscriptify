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

func TestGetBinDest(t *testing.T) {
	Convey("Should return the bin destination", t, func() {
		// String to be hashed ends up as foobar
		bin, _, err := getBinDest("bar", []string{"foo"}, "baz")
		So(err, ShouldBeNil)
		So(bin, ShouldEqual, "baz/3858f62230ac3c915f300c664312c63f")
	})

	Convey("Should return the hash", t, func() {
		// String to be hashed ends up as foobar
		_, h, err := getBinDest("bar", []string{"foo"}, "baz")
		So(err, ShouldBeNil)
		So(h, ShouldEqual, "3858f62230ac3c915f300c664312c63f")
	})
}

func TestNewScriptPath(t *testing.T) {
	Convey("Should choose an alternate filename when "+
		"the source.go already exists", t, func() {
		sp := NewScriptPath(
			"acbd18db4cc2f85cedef654fccc4a4d8",
			"_test/fixtures/exit0",
		)
		So(sp.Original, ShouldEqual, "_test/fixtures/exit0")
		So(sp.Generated, ShouldEqual,
			"_test/fixtures/acbd18db4cc2f85cedef654fccc4a4d8-exit0.go")
		So(sp.Clean, ShouldBeTrue)
	})

	Convey("Should only append .go if it's missing", t, func() {
		sp := NewScriptPath(
			"acbd18db4cc2f85cedef654fccc4a4d8",
			"_test/fixtures/exit0.go",
		)
		So(sp.Original, ShouldEqual, "_test/fixtures/exit0.go")
		So(sp.Generated, ShouldEqual, "_test/fixtures/exit0.go")
		So(sp.Clean, ShouldBeFalse)
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

func TestRunScriptDirWithOpts(t *testing.T) {
	fixDir := filepath.Join("_test", "fixtures")
	tmpDir := filepath.Join("_test", "tmp")

	Convey("Should run a go dir", t, func() {
		p := filepath.Join(fixDir, "exit15_dir")
		exit, err := RunScriptDirWithOpts(p, []string{}, ScriptOptions{
			Temp:  tmpDir,
			Stdin: nil, Stdout: ioutil.Discard, Stderr: ioutil.Discard,
		})
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)
	})

	Convey("Should create all temp dirs if they don't exist", t, func() {
		src := filepath.Join(fixDir, "exit15_dir")
		tmpRoot := filepath.Join(tmpDir, "nested")
		tmp := filepath.Join(tmpRoot, "dirs")

		// Just to be safe, remove the dir ahead of time
		os.RemoveAll(tmpRoot)

		exit, err := RunScriptDirWithOpts(src, []string{}, ScriptOptions{
			Temp:  tmp,
			Stdin: nil, Stdout: ioutil.Discard, Stderr: ioutil.Discard,
		})
		So(err, ShouldBeNil)
		So(exit, ShouldEqual, 15)

		// Now check to make sure the tmp exists
		_, err = os.Stat(tmp)
		So(err, ShouldBeNil)
	})

	Convey("Should output stdout and stderr", t, func() {
		p := filepath.Join(fixDir, "exit15_dir")
		var stdout, stderr bytes.Buffer

		_, err := RunScriptDirWithOpts(p, []string{}, ScriptOptions{
			Temp:  tmpDir,
			Stdin: nil, Stdout: &stdout, Stderr: &stderr,
		})
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(&stdout)
		So(string(b), ShouldEqual, "STDOUT: Exiting 15")
		b, _ = ioutil.ReadAll(&stderr)
		So(string(b), ShouldEqual, "STDERR: Exiting 15")
	})
}

func TestRunScriptsWithOpts(t *testing.T) {
	fixDir := filepath.Join("_test", "fixtures")
	tmpDir := filepath.Join("_test", "tmp")
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

	Convey("Should output stdout and stderr", t, func() {
		src := filepath.Join(fixDir, "exit15.go")
		var stdout, stderr bytes.Buffer

		_, err := RunScriptsWithOpts([]string{src}, []string{}, ScriptOptions{
			Temp:  tmpDir,
			Stdin: nil, Stdout: &stdout, Stderr: &stderr,
		})
		So(err, ShouldBeNil)

		b, _ := ioutil.ReadAll(&stdout)
		So(string(b), ShouldEqual, "STDOUT: Exiting 15")
		b, _ = ioutil.ReadAll(&stderr)
		So(string(b), ShouldEqual, "STDERR: Exiting 15")
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

	Convey("Should not return a dir", t, func() {
		s, err := FindScript([]string{
			filepath.Join(fixDir, "baz"),
			filepath.Join(fixDir, "baz", "bat"),
		})
		So(err, ShouldBeNil)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz", "bat"))
	})
}

// TODO: Find a way to make FindScriptOrDir tests pass on OSX.
// NOTE: Due to OSX's case insensitivity, it's hard (maybe possible?)
// to know the *actual* filename of the found file. Tests, then
// have to ignore the string output of this function, as it will
// fail on OSX. I'd love to see a workaround for this issue.
//
// So, a fail on OSX does not currently mean truly failing tests.
// Run on Linux to confirm.
func TestFindScriptOrDir(t *testing.T) {
	fixDir := filepath.Join("_test", "fixtures")

	Convey("Should find a given script", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "foo"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, filepath.Join(fixDir, "foo"))
	})

	Convey("Should find the given subdirectory script", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "baz", "bat"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz", "bat"))
	})

	Convey("Should find the first given script", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bang", "boom", "flash"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, filepath.Join(fixDir, "bar"))

		s, isDir, err = FindScriptOrDir([]string{
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bang", "boom", "flash"),
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz", "bat"))

		s, isDir, err = FindScriptOrDir([]string{
			filepath.Join(fixDir, "bang", "boom", "flash"),
			filepath.Join(fixDir, "baz", "bat"),
			filepath.Join(fixDir, "bar"),
			filepath.Join(fixDir, "foo"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, filepath.Join(fixDir, "bang", "boom", "flash"))
	})

	Convey("Should find the first given script dir", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "idontexist"),
			filepath.Join(fixDir, "baz"),
			filepath.Join(fixDir, "bang", "boom"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeTrue)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz"))

		s, isDir, err = FindScriptOrDir([]string{
			filepath.Join(fixDir, "idontexist"),
			filepath.Join(fixDir, "bang", "boom"),
			filepath.Join(fixDir, "baz"),
		}, false)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeTrue)
		So(s, ShouldEqual, filepath.Join(fixDir, "bang", "boom"))
	})

	Convey("Should use the directory of a script, if useDir", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "idontexist"),
			filepath.Join(fixDir, "baz", "bat.go"),
		}, true)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeTrue)
		So(s, ShouldEqual, filepath.Join(fixDir, "baz"))

		s, isDir, err = FindScriptOrDir([]string{
			filepath.Join(fixDir, "idontexist"),
			filepath.Join(fixDir, "bang", "boom", "flash.go"),
		}, true)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeTrue)
		So(s, ShouldEqual, filepath.Join(fixDir, "bang", "boom"))
	})

	Convey("Should not use the directory of a directory", t, func() {
		s, isDir, err := FindScriptOrDir([]string{
			filepath.Join(fixDir, "idontexist"),
			filepath.Join(fixDir, "bang", "boom"),
		}, true)
		So(err, ShouldBeNil)
		So(isDir, ShouldBeTrue)
		So(s, ShouldEqual, filepath.Join(fixDir, "bang", "boom"))
	})

	Convey("Should not allow non-.go extensions in subdirs if useDir", t,
		func() {
			_, _, err := FindScriptOrDir([]string{
				filepath.Join(fixDir, "idontexist"),
				filepath.Join(fixDir, "bang", "boom", "flash"),
			}, true)
			So(err, ShouldNotBeNil)
		})
}

// TODO: Find a way to make findScriptOrDir tests pass on OSX.
// NOTE: Due to OSX's case insensitivity, it's hard (maybe possible?)
// to know the *actual* filename of the found file. Tests, then
// have to ignore the string output of this function, as it will
// fail on OSX. I'd love to see a workaround for this issue.
//
// So, a fail on OSX does not currently mean truly failing tests.
// Run on Linux to confirm.
func TestFindScriptOrDirlower(t *testing.T) {
	cwd, _ := os.Getwd()
	fixDir := filepath.Join("_test", "fixtures")
	// Fake the cwd from the root of goscriptify, to
	// the fixDir
	fixDirCwd := filepath.Join(cwd, fixDir)

	Convey("Should find from the given cwd", t, func() {
		s, isDir, err := findScriptOrDir(
			filepath.Join(fixDirCwd, "bang", "boom"),
			[]string{"idontexist", "flash.go"},
			false,
		)

		So(err, ShouldBeNil)
		So(isDir, ShouldBeFalse)
		So(s, ShouldEqual, "flash.go")
	})

	Convey("Should not allow non-.go extensions in subdirs if useDir", t,
		func() {
			_, _, err := findScriptOrDir(fixDirCwd,
				[]string{"idontexist", "exit0"}, true)
			So(err, ShouldBeNil)
		})
}
