package updater

import (
	"context"

	"github.com/coreos/go-semver/semver"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/google/go-github/v61/github"
)

func CheckUpdates() (bool, string, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "djohts", "tpc-truckersmp", nil)
	if err != nil {
		return false, "", err
	}

	if len(releases) == 0 {
		return false, "", nil
	}

	latestRelease := releases[0]
	needsUpdate := semver.New(constants.APP_VERSION).LessThan(*semver.New((*latestRelease.TagName)[1:]))

	return needsUpdate, *latestRelease.TagName, nil
}
