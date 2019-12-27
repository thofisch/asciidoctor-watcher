package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ensureDirectory(path string) string {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("%q is not a valid directory!\n", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	return absPath
}

type fsInfo struct {
	absolutePath string
	relativePath string
	extension    string
	isDir        bool
}

func getFsInfo(path string) fsInfo {
	rel, _ := filepath.Rel(sourcePath, path)
	isDir := false

	stat, err := os.Stat(path)
	if err == nil {
		isDir = stat.IsDir()
	} else {
		_, ok := watches[path]
		isDir = ok
	}

	return fsInfo{
		absolutePath: path,
		relativePath: rel,
		extension:    strings.ToLower(filepath.Ext(path)),
		isDir:        isDir,
	}
}

func copyFile(srcPath string, dstPath string) (err error) {
	fmt.Printf("cp %q %q\n", srcPath, dstPath)

	sfi, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	dfi, err := os.Stat(dstPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return nil
		}
	}

	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		cErr := out.Close()
		if err == nil {
			err = cErr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return nil
	}

	err = out.Sync()

	return err
}

func removeAll(path string) error {
	fmt.Printf("rm -rf %q\n", path)

	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, d := range dir {
		os.RemoveAll(filepath.Join(path, d.Name()))
	}

	os.Remove(path)

	return nil
}

func mkDir(path string) {
	fmt.Printf("mkdir %q\n", path)
	os.Mkdir(path, os.ModePerm)
}

func rm(path string) {
	fmt.Printf("rm %q\n", path)
	os.Remove(path)
}

func File(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	fmt.Printf("cp %q %q\n", src, dst)

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func Dir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	fmt.Printf("mkdir %q\n", dst)
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = Dir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {

			if contains(knownExtensions, filepath.Ext(fd.Name())) {
				continue
			}

			if err = File(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
