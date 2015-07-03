package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCopyFile(t *testing.T) {
	dst := filepath.Join("..", "_test", "tmp", "copyfile")
	os.Remove(dst)

	Convey("Should copy the file to the destination", t, func() {
		src := filepath.Join("..", "_test", "fixtures", "foo")
		err := CopyFile(dst, src)
		So(err, ShouldBeNil)
		bs, _ := ioutil.ReadFile(dst)
		So(string(bs), ShouldEqual, "foo\n")
	})

	Convey("Should replace the destination", t, func() {
		src := filepath.Join("..", "_test", "fixtures", "bar")
		err := CopyFile(dst, src)
		So(err, ShouldBeNil)
		bs, _ := ioutil.ReadFile(dst)
		So(string(bs), ShouldEqual, "bar\n")
	})
}

func TestHashString(t *testing.T) {
	Convey("Should return the md5 of the input", t, func() {
		s := HashString("foo")
		So(s, ShouldEqual, "acbd18db4cc2f85cedef654fccc4a4d8")
	})
}

func TestExists(t *testing.T) {
	fixDir := filepath.Join("..", "_test", "fixtures")

	Convey("Should return if the path exists or not", t, func() {
		exists, _, _ := Exists(filepath.Join(fixDir, "foo"))
		So(exists, ShouldBeTrue)

		exists, _, _ = Exists(filepath.Join(fixDir, "baz"))
		So(exists, ShouldBeTrue)

		exists, _, _ = Exists(filepath.Join(fixDir, "baz", "bat"))
		So(exists, ShouldBeTrue)

		exists, _, _ = Exists(filepath.Join(fixDir, "idontexist"))
		So(exists, ShouldBeFalse)
	})

	Convey("Should not return an error if the path doesn't exist", t, func() {
		_, _, err := Exists(filepath.Join(fixDir, "idontexist"))
		So(err, ShouldBeNil)
	})

	Convey("Should return if the path is a dir or not", t, func() {
		_, isDir, _ := Exists(filepath.Join(fixDir, "baz"))
		So(isDir, ShouldBeTrue)

		_, isDir, _ = Exists(filepath.Join(fixDir, "foo"))
		So(isDir, ShouldBeFalse)

		_, isDir, _ = Exists(filepath.Join(fixDir, "baz", "bat"))
		So(isDir, ShouldBeFalse)

		_, isDir, _ = Exists(filepath.Join(fixDir, "idontexist"))
		So(isDir, ShouldBeFalse)
	})
}
