package util

import (
	"os"
	"path/filepath"
)

func FindProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			panic("Failed to find project root")
		}
		dir = parentDir
	}
}
