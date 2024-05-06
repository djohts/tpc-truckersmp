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
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if s.IsDir() {
		return true
	}
	return false
}

func GetDocumentsPath() (string, error) {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\User Shell Folders", registry.ALL_ACCESS)
	if err != nil {
		return "", err
	}

	defer key.Close()
	path, _, _ := key.GetStringValue("Personal")
	path = strings.TrimSpace(strings.Replace(path, "%USERPROFILE%", os.Getenv("USERPROFILE"), -1))

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
