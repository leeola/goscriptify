package goscriptify

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildDir(t *testing.T) {
	fixDir := filepath.Join("_test", "fixtures")
	dst := filepath.Join("_test", "tmp", "bin")
	os.Remove(dst)

	Convey("Should build the source to the target relative location", t, func() {
		err := BuildDir(dst, filepath.Join(fixDir, "exit15_dir"))
		So(err, ShouldBeNil)
		_, err = os.Stat(dst)
		So(err, ShouldBeNil)
	})

	Convey("Should build the source to the target absolute location", t, func() {
		absDst, err := filepath.Abs(dst)
		So(err, ShouldBeNil)

		err = BuildDir(absDst, filepath.Join(fixDir, "exit15_dir"))
		So(err, ShouldBeNil)
		_, err = os.Stat(dst)
		So(err, ShouldBeNil)
	})

	os.Remove(dst)

	Convey("Should return build error information", t, func() {
		err := BuildDir(dst, filepath.Join(fixDir, "synerr_dir"))
		buildErr, ok := err.(*BuildError)
		So(ok, ShouldBeTrue)
		So(buildErr.Exit, ShouldNotEqual, 0)
		So(buildErr.Error(), ShouldContainSubstring, "syntax error")
	})
}

func TestBuildFiles(t *testing.T) {
	dst := filepath.Join("_test", "tmp", "bin")
	os.Remove(dst)

	Convey("Should build the source to the target location", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15.go")
		err := BuildFiles(dst, []string{src})
		So(err, ShouldBeNil)
		_, err = os.Stat(dst)
		So(err, ShouldBeNil)
	})

	os.Remove(dst)

	Convey("Should only allow .go filenames for source", t, func() {
		src := filepath.Join("_test", "fixtures", "exit15")
		err := BuildFiles(dst, []string{src})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "go")
		So(err.Error(), ShouldContainSubstring, "ext")
	})

	Convey("Should return build error information", t, func() {
		src := filepath.Join("_test", "fixtures", "synerr.go")
		err := BuildFiles(dst, []string{src})
		buildErr, ok := err.(*BuildError)
		So(ok, ShouldBeTrue)
		So(buildErr.Exit, ShouldNotEqual, 0)
		So(buildErr.Error(), ShouldContainSubstring, "syntax error")
	})
}
