package utils

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"golang.org/x/sys/windows/registry"
)

//go:embed SII_Decrypt.exe
var decrypt_bytes []byte

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

func EnsureDecrypt() *os.File {
	if IsFile("SII_Decrypt.exe") {
		data, err := os.ReadFile("SII_Decrypt.exe")
		HandleError(err)

		if !bytes.Equal(data, decrypt_bytes) {
			err = os.Remove("SII_Decrypt.exe")
			HandleError(err)

			return installDecrypt()
		}

		f, err := os.OpenFile("SII_Decrypt.exe", os.O_RDWR, 0o755)
		HandleError(err)

		defer f.Close()

		return f
	}

	return installDecrypt()
}

func installDecrypt() *os.File {
	f, err := os.Create("SII_Decrypt.exe")
	HandleError(err)

	defer f.Close()

	_, err = f.Write(decrypt_bytes)
	HandleError(err)

	err = f.Chmod(0o755)
	HandleError(err)

	return f
}
