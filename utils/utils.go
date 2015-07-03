package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

// Copy the source path file to the destination
func CopyFile(dstName, srcName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// Hash a string
func HashString(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Exists returns whether the path exists or not, and if it is a directory
// or not.
func Exists(p string) (exists bool, isDir bool, err error) {
	fi, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, fi.IsDir(), nil
}
