package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"golang.org/x/sys/windows/registry"
)

func FormatPath(path string, parentPath string) string {
	return fmt.Sprintf("...%s", strings.Replace(path, parentPath, "", 1))
}

func IsFile(path string) bool {
	s, err := os.Stat(path)
	return err == nil && !s.IsDir()
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	return err == nil && s.IsDir()
}

func GetDocumentsPath() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Personal")
	if err != nil {
		return "", err
	}

	path = strings.ReplaceAll(path, "%USERPROFILE%", os.Getenv("USERPROFILE"))
	path = strings.TrimSpace(path)

	return path, nil
}

func HandleError(err error) {
	if err != nil {
		log.Helper()
		log.Error("Fatal error, stop working!")
		log.Error(err)

		fmt.Printf("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		os.Exit(1)
	}
}

func FindOne[T any](slice []T, filter func(*T) bool) (element *T) {
	for i := range slice {
		if filter(&slice[i]) {
			return &slice[i]
		}
	}

	return nil
}

func Filter[T any](slice []T, filter func(*T) bool) []*T {
	var ret []*T = make([]*T, 0)

	for i := range slice {
		if filter(&slice[i]) {
			ret = append(ret, &slice[i])
		}
	}

	return ret
}
