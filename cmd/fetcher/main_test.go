package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindMatchesOnTvMaze(t *testing.T) {

}

func TestCreateShowsJSON(t *testing.T) {
	// expect shows.json
}

func TestCreateSeasonsJSON(t *testing.T) {
}

func TestCreateEpisodesJSON(t *testing.T) {
	// expect episodes.json

}
func copyR(src, dest string) {
	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		target := strings.Replace(path, src, dest, -1)
		if info.IsDir() {
			os.MkdirAll(target, 0755)
		} else {
			copyFile(path, target)
		}
		return nil
	})
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}
