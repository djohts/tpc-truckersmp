package updater

import (
	"context"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/coreos/go-semver/semver"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/google/go-github/v61/github"
	"github.com/minio/selfupdate"
)

func CheckUpdates() (bool, string, error) {
	latestRelease, err := getLatestRelease()
	if err != nil {
		return false, "", err
	}

	needsUpdate := semver.New(constants.APP_VERSION).LessThan(*semver.New((*latestRelease.TagName)[1:]))

	return needsUpdate, *latestRelease.TagName, nil
}

func UpdateSelf() (bool, error) {
	latestRelease, err := getLatestRelease()
	if err != nil {
		return false, err
	}

	asset := utils.FindOne(latestRelease.Assets, func(asset **github.ReleaseAsset) bool {
		return *(*asset).Name == "tpc.exe"
	})
	if asset == nil {
		return false, nil
	}

	log.Info("Downloading latest version...")
	res, err := http.Get(*(*asset).BrowserDownloadURL)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	log.Info("Applying update...")
	err = selfupdate.Apply(res.Body, selfupdate.Options{})
	if err != nil {
		return false, err
	}

	return true, nil
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
