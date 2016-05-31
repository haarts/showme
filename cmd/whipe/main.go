package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var fake = flag.Bool("fake", false, "only print what it is about to do")

func rmSymlink(path string, info os.FileInfo, err error) error {
	if info.Mode()&os.ModeSymlink != 0 {
		dereferenced, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if *fake {
			fmt.Printf("removing '%s'\n", dereferenced)
		} else {
			err = os.Remove(dereferenced)
			return err
		}

	}
	return nil
}

func main() {
	flag.Parse()
	root := flag.Args()[0]

	filepath.Walk(root, rmSymlink)
	if *fake {
		fmt.Printf("removing '%s'\n", root)
	} else {
		os.RemoveAll(root)

	}
}
