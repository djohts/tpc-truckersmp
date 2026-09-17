package updater

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/coreos/go-semver/semver"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/google/go-github/v90/github"
)

var p *tea.Program

// ReleaseChangelog holds the tag and cleaned body of a GitHub release.
type ReleaseChangelog struct {
	Tag  string
	Body string
}

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

// GetChangelogsBetween returns changelogs for all releases newer than the
// currently installed version, ordered oldest-first so callers can display
// them in chronological order.
func GetChangelogsBetween() ([]ReleaseChangelog, error) {
	if constants.APP_VERSION == "dev" {
		return nil, nil
	}

	client, err := github.NewClient()
	if err != nil {
		return nil, err
	}
	releases, _, err := client.Repositories.ListReleases(context.Background(), "djohts", "tpc-truckersmp", nil)
	if err != nil {
		return nil, err
	}

	currentVersion := semver.New(constants.APP_VERSION)
	var changelogs []ReleaseChangelog
	for _, release := range releases {
		if release.TagName == nil || len(*release.TagName) < 2 {
			continue
		}
		tagVersion, err := semver.NewVersion((*release.TagName)[1:])
		if err != nil {
			continue
		}
		if currentVersion.LessThan(*tagVersion) {
			body := stripChecksumsSection(release.GetBody())
			changelogs = append(changelogs, ReleaseChangelog{Tag: *release.TagName, Body: body})
		}
	}

	// GitHub returns releases newest-first; reverse to oldest-first.
	for i, j := 0, len(changelogs)-1; i < j; i, j = i+1, j-1 {
		changelogs[i], changelogs[j] = changelogs[j], changelogs[i]
	}

	return changelogs, nil
}

// stripChecksumsSection removes any markdown heading whose text contains
// "sha", "checksum", or "hash" (case-insensitive), along with all lines
// that follow it until the next heading of the same or higher level.
func stripChecksumsSection(body string) string {
	lines := strings.Split(body, "\n")
	var result []string
	inChecksumsSection := false
	checksumLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			headerText := strings.ToLower(strings.TrimSpace(trimmed[level:]))
			if strings.Contains(headerText, "sha") || strings.Contains(headerText, "checksum") || strings.Contains(headerText, "hash") {
				inChecksumsSection = true
				checksumLevel = level
				continue
			}
			// A heading at the same or higher level ends the checksums section.
			if inChecksumsSection && level <= checksumLevel {
				inChecksumsSection = false
			}
		}

		if !inChecksumsSection {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
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
		progress: progress.New(progress.WithDefaultBlend()),
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
	client, err := github.NewClient()
	if err != nil {
		return nil, err
	}
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
