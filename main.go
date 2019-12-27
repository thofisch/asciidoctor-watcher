package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var watcher *fsnotify.Watcher
var watches = make(map[string]bool)
var sourcePath = ""
var targetPath = ""

var knownExtensions = []string{".asciidoc", ".adoc", ".asc"}
var defaultIndexFile = "index"

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: watcher <watch_path> <output_path>")
		os.Exit(1)
	}

	sourcePath = ensureDirectory(os.Args[1])
	targetPath = ensureDirectory(os.Args[2])
	defaultIndexFile = findDefaultIndexFile(sourcePath)

	Dir(sourcePath, targetPath)

	fmt.Printf("Watch %q => %q ...\n", sourcePath, targetPath)

	watcher, _ = fsnotify.NewWatcher()
	defer func() {
		err := watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(sourcePath, watchDir); err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go watch()

	<-done
}

func findDefaultIndexFile(sourcePath string) string {
	tested := make([]string, len(knownExtensions))

	for i, ext := range knownExtensions {
		indexFile := "index" + ext
		tested[i] = indexFile
		_, err := os.Stat(filepath.Join(sourcePath, indexFile))
		if err == nil {
			return indexFile
		}
	}

	log.Fatalf("Unable to find a default index file (%s)", strings.Join(tested, ", "))
	return ""
}

func watch() {
	for {
		select {
		case event := <-watcher.Events:
			name := event.Name
			op := event.Op

			fmt.Printf("EVENT! %q %q\n", name, op)

			if !(name == "" || name == "." || name == "..") {
				handleFsEvent(name, op)
			}

		case err := <-watcher.Errors:
			fmt.Println("ERROR", err)
		}
	}
}

func handleFsEvent(path string, op fsnotify.Op) {
	info := getFsInfo(path)
	newPath := filepath.Join(targetPath, info.relativePath)
	regularFile := !isAsciiDoctorFile(info)

	if op&fsnotify.Create == fsnotify.Create {
		if info.isDir {
			watchPath(path)
			mkDir(newPath)
			return
		} else {
			if regularFile {
				copyFile(info.absolutePath, newPath)
			}

			rebuild()
		}
	}

	if op&fsnotify.Remove == fsnotify.Remove {
		if info.isDir {
			unwatchPath(path)
			removeAll(newPath)
			return
		} else {
			if regularFile {
				rm(newPath)
			}
			rebuild()
		}
	}

	if op&fsnotify.Write == fsnotify.Write {
		if !info.isDir {
			if regularFile {
				copyFile(info.absolutePath, newPath)
			}
			rebuild()
		}
	}

	if op&fsnotify.Rename == fsnotify.Rename {
		fmt.Printf("R: %q\n", path)
	}

	if op&fsnotify.Chmod == fsnotify.Chmod {
		if !info.isDir {
			if regularFile {
				copyFile(info.absolutePath, newPath)
			}
			rebuild()
		}
	}
}

func watchDir(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() {
		return watchPath(path)
	}
	return nil
}

func watchPath(path string) error {
	fmt.Printf("W: %q\n", path)

	watches[path] = true

	return watcher.Add(path)
}

func unwatchPath(path string) error {
	fmt.Printf("U: %q\n", path)

	delete(watches, path)

	return watcher.Remove(path)
}

func isAsciiDoctorFile(info fsInfo) bool {
	return contains(knownExtensions, info.extension)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func rebuild() {
	fmt.Printf("asciidoctor index.adoc -D %s\n", targetPath)

	cmd := exec.Command("asciidoctor", defaultIndexFile, "-D", targetPath)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(stdout))
}
