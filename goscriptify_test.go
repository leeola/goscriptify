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
