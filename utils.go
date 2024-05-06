package main

import (
	"fmt"
	"os"
	"strings"
)

func isFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}
	return true
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if s.IsDir() {
		return true
	}
	return false
}

func formatPath(path string, parentPath string) string {
	return fmt.Sprintf("...%s", strings.Replace(path, parentPath, "", 1))
}
