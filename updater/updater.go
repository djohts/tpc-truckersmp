package updater

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/coreos/go-semver/semver"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/google/go-github/v76/github"
)

var p *tea.Program

func CheckUpdates() (bool, string, error) {
	if constants.APP_VERSION == "dev" {
		return false, "", nil
	}

	release, err := getLatestRelease()
	if err != nil {
		return false, "", err
	}

	needsUpdate := semver.New(constants.APP_VERSION).LessThan(*semver.New((*release.TagName)[1:]))

	return needsUpdate, *release.TagName, nil
}

func UpdateSelf() (bool, error) {
	release, err := getLatestRelease()
	if err != nil {
		return false, err
	}

	asset := utils.FindOne(release.Assets, func(asset **github.ReleaseAsset) bool {
		return *(*asset).Name == "tpc.exe"
	})
	if asset == nil {
		return false, errors.New("executable file not found in release")
	}

	log.Info("Downloading latest version...")
	filename := "tpc-" + *release.TagName + ".exe"
	err = DownloadFile(*(*asset).BrowserDownloadURL, filename)
	if err != nil {
		return false, errors.New("failed to download latest version: " + err.Error())
	}

	log.Info("Verifying checksum...")
	checksumAsset := utils.FindOne(release.Assets, func(asset **github.ReleaseAsset) bool {
		return *(*asset).Name == "checksums.txt"
	})
	if checksumAsset == nil {
		return false, errors.New("checksum file not found in release")
	}

	checksumBytes := make([]byte, 64)
	resp, err := http.Get(*(*checksumAsset).BrowserDownloadURL)
	if err != nil {
		return false, errors.New("failed to download checksum file: " + err.Error())
	}
	defer resp.Body.Close()
	_, err = resp.Body.Read(checksumBytes)
	if err != nil {
		return false, errors.New("failed to read checksum data: " + err.Error())
	}

	file, err := os.Open(filename)
	if err != nil {
		return false, errors.New("failed to open downloaded file: " + err.Error())
	}
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return false, errors.New("failed to read file for checksum: " + err.Error())
	}
	file.Close()
	if fmt.Sprintf("%x", h.Sum(nil)) != string(checksumBytes) {
		return false, fmt.Errorf("checksum mismatch. expected: %s, got: %x", checksumBytes, h.Sum(nil))
	}

	log.Info("Applying update...")
	if err := ApplyUpdate(file); err != nil {
		return false, errors.New("failed to apply update: " + err.Error())
	}

	return true, nil
}

func ApplyUpdate(file *os.File) error {
	executablePath, err := os.Executable()
	if err != nil {
		return errors.New("failed to get executable path: " + err.Error())
	}

	if err := os.Rename(executablePath, "tpc-"+constants.APP_VERSION+"-old.exe"); err != nil {
		return errors.New("failed to rename old executable: " + err.Error())
	}
	if err := os.Rename(file.Name(), executablePath); err != nil {
		return errors.New("failed to rename new executable: " + err.Error())
	}

	return nil
}

func DownloadFile(url, filename string) error {
	res, err := http.Get(url)
	if err != nil {
		return errors.New("failed to download file: " + err.Error())
	}
	defer res.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return errors.New("failed to create file: " + err.Error())
	}
	defer file.Close()

	pw := &progressWriter{
		total:  int(res.ContentLength),
		file:   file,
		reader: res.Body,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	m := model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}
	// Start Bubble Tea
	p = tea.NewProgram(m)

	// Start the download
	go pw.Start()

	if _, err := p.Run(); err != nil {
		return errors.New("failed to run program: " + err.Error())
	}

	return nil
}

func getLatestRelease() (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "djohts", "tpc-truckersmp", nil)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, nil
	}

	return releases[0], nil
}

type progressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
}

func (pw *progressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		p.Send(progressErrMsg{err})
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}
	return len(p), nil
}
