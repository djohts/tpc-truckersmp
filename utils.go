package main

import "os"

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