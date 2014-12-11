package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
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
