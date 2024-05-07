package downloader

import (
	"github.com/djohts/tpc-truckersmp/updater"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/google/go-github/v61/github"
)

func DownloadSiiDecrypt() (bool, error) {
	release, err := updater.GetLatestRelease()
	if err != nil {
		return false, err
	}

	asset := utils.FindOne(release.Assets, func(asset **github.ReleaseAsset) bool {
		return *(*asset).Name == "SII_Decrypt.exe"
	})
	if asset == nil {
		return false, nil
	}

	err = updater.DownloadFile(*(*asset).BrowserDownloadURL, "SII_Decrypt.exe")
	if err != nil {
		return false, err
	}

	return true, nil
}
