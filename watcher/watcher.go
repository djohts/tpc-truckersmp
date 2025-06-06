package watcher

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.tcp.direct/kayos/sendkeys"
	"github.com/bradhe/stopwatch"
	"github.com/charmbracelet/log"
	"github.com/djohts/tpc-truckersmp/config"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/decrypt"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/fsnotify/fsnotify"
)

var (
	profileList   []string
	watchPathList []string

	keyboard *sendkeys.KBWrap
)

func Init() {
	decrypt.EnsureDecrypt()

	if config.Get().Auto {
		addCamsWatchers()

		var err error
		keyboard, err = sendkeys.NewKBWrapWithOptions()
		utils.HandleError(err)
	}

	err := getProfileList()
	utils.HandleError(err)
	if len(profileList) == 0 {
		utils.HandleError(errors.New("no local profiles found"))
	}

	addSaveWatchers()

	watch, err := fsnotify.NewWatcher()
	utils.HandleError(err)
	defer watch.Close()

	err = addPathToWatch(watch)
	utils.HandleError(err)

	go watchFiles(watch)
	select {}
}

func addCamsWatchers() {
	games := []string{constants.ETS, constants.ATS}

	for _, game := range games {
		gamePath := filepath.Join(constants.DocumentsPath, game)
		camsPath := filepath.Join(gamePath, "cams.txt")

		if utils.IsDir(gamePath) && utils.IsFile(camsPath) {
			watchPathList = append(watchPathList, camsPath)
		}
	}
}

func addSaveWatchers() {
	for _, profilePath := range profileList {
		savePath := filepath.Join(profilePath, "save")
		if utils.IsDir(savePath) {
			watchPathList = append(watchPathList, savePath)
		}
	}
	profileList = nil
}

func listProfiles(path string, d fs.DirEntry, err error) error {
	if err != nil || d == nil || !d.IsDir() {
		return err
	}
	if filepath.Base(filepath.Dir(path)) == "profiles" {
		profileList = append(profileList, path)
	}
	return nil
}

func getProfileList() error {
	paths := []struct {
		name string
		get  func() (string, error)
	}{
		{"ETS2", getEts2Path},
		{"ATS", getAtsPath},
	}

	for _, p := range paths {
		if path, err := p.get(); err == nil {
			if err := filepath.WalkDir(path, listProfiles); err != nil {
				return err
			}
		}
	}
	return nil
}

func getEts2Path() (string, error) {
	return getGamePath(constants.ETS)
}

func getAtsPath() (string, error) {
	return getGamePath(constants.ATS)
}

func getGamePath(game string) (string, error) {
	path := filepath.Join(constants.DocumentsPath, game)
	if utils.IsDir(path) {
		return path, nil
	}
	return "", fmt.Errorf("%s not found", game)
}

func decryptSii(filePath string) (bool, error) {
	decryptFile := decrypt.EnsureDecrypt()
	execPath, err := filepath.Abs(decryptFile.Name())
	utils.HandleError(err)

	cmd := exec.Command(execPath, filePath)
	log.Debugf("Decrypting %s", utils.FormatPath(filePath, constants.DocumentsPath))
	buf, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
				return false, nil
			}
			return false, errors.New(string(buf))
		}
		return false, err
	}
	return true, nil
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	output := make([]string, 0)

	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	return output, scanner.Err()
}

func writeFile(filePath string, output string) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(output)
	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func addPathToWatch(watch *fsnotify.Watcher) error {
	for _, watchPath := range watchPathList {
		if err := watch.Add(watchPath); err != nil {
			return err
		}
		log.Info("Monitoring " + utils.FormatPath(watchPath, constants.DocumentsPath))
	}
	return nil
}

func watchFiles(watch *fsnotify.Watcher) {
	var isOnCooldown bool
	var mu sync.Mutex

	for {
		select {
		case ev := <-watch.Events:
			{
				if ev.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				base := filepath.Base(ev.Name)

				if base == "quicksave" {
					mu.Lock()
					if isOnCooldown {
						mu.Unlock()
						continue
					}
					isOnCooldown = true
					mu.Unlock()

					go func() {
						time.Sleep(750 * time.Millisecond)
						mu.Lock()
						isOnCooldown = false
						mu.Unlock()
					}()

					time.Sleep(250 * time.Millisecond)

					flushWatch := stopwatch.Start()
					done, err := flushChange(filepath.Join(ev.Name, "game.sii"))
					flushWatch.Stop()
					utils.HandleError(err)

					if done {
						log.Infof("Updated %s (%dms)", utils.FormatPath(filepath.Join(ev.Name, "game.sii"), constants.DocumentsPath), flushWatch.Milliseconds().Nanoseconds())
					}
				} else if base == "cams.txt" {
					err := keyboard.Type(config.Get().Keybinds.Quicksave)
					utils.HandleError(err)
					log.Info("Detected cams.txt update, sending quicksave keybind")
				}
			}

		case err := <-watch.Errors:
			{
				utils.HandleError(err)
			}
		}
	}
}

func flushChange(filePath string) (bool, error) {
	if !utils.IsFile(filePath) {
		return false, nil
	}
	needEdit, err := decryptSii(filePath)
	if err != nil {
		return false, err
	}
	if !needEdit {
		log.Debug("No need to edit " + utils.FormatPath(filePath, constants.DocumentsPath))
		return false, nil
	}

	if !utils.IsFile(filePath) {
		return false, nil
	}
	sii, err := readFile(filePath)
	if err != nil {
		return false, err
	}

	camsPath := filepath.Join(constants.DocumentsPath, constants.ETS, "cams.txt")
	if strings.Contains(filePath, constants.ATS) {
		camsPath = filepath.Join(constants.DocumentsPath, constants.ATS, "cams.txt")
	}
	if !utils.IsFile(camsPath) {
		return false, nil
	}
	cams, err := readFile(camsPath)
	if err != nil {
		return false, err
	}

	if len(cams) > 0 {
		location, rotation := parseCamsCoordinate(cams)
		output := editSii(sii, location, rotation)

		if !utils.IsFile(filePath) {
			return false, nil
		}
		err = writeFile(filePath, output)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func parseCamsCoordinate(cams []string) (string, string) {
	row := strings.ReplaceAll(cams[len(cams)-1], ";", ",")
	parts := strings.Split(row, " , ")
	location := parts[1]
	rotation := strings.Replace(parts[2], ",", ";", 1)
	log.Infof("Target: (%s) (%s)", location, rotation)
	return location, rotation
}

func editSii(siiArray []string, location, rotation string) string {
	cfg := config.Get().Features
	attachTrailer := cfg.AttachTrailer
	refuel := cfg.Refuel
	teleport := cfg.Teleport
	refuelLevel := fmt.Sprintf("%d", cfg.RefuelRelative)

	myTruckNameless := ""
	attachTrailerState := 0

	for i := 0; i < len(siiArray); i++ {
		line := siiArray[i]

		switch {
		case strings.HasPrefix(line, " my_truck: _nameless"):
			myTruckNameless = strings.Split(line, ": ")[1]

		case strings.HasPrefix(line, " assigned_trailer: _nameless") && attachTrailer:
			attachTrailerState = 1

		case strings.HasPrefix(line, " assigned_trailer_connected: false") && attachTrailerState == 1:
			attachTrailerState = 2
			siiArray[i] = " assigned_trailer_connected: true"

		case strings.HasPrefix(line, " nav_node_position:") && attachTrailerState == 2:
			siiArray[i] = " nav_node_position: (0, 0, 0)"

		case strings.HasPrefix(line, " truck_placement:") && teleport:
			siiArray[i] = " truck_placement: (" + location + ") (" + rotation + ")"

		case strings.HasPrefix(line, " trailer_placement:") && (attachTrailerState == 2 || teleport):
			siiArray[i] = " trailer_placement: (0, 0, 0) (" + rotation + ")"

		case strings.HasPrefix(line, " slave_trailer_placements[") && (attachTrailerState == 2 || teleport):
			siiArray[i] = strings.Split(line, ": ")[0] + ": (0, 0, 0) (" + rotation + ")"

		case strings.HasPrefix(line, " trailer_body_wear:"),
			strings.HasPrefix(line, " trailer_body_wear_unfixable:"),
			strings.HasPrefix(line, " chassis_wear:"),
			strings.HasPrefix(line, " chassis_wear_unfixable:"),
			strings.HasPrefix(line, " engine_wear:"),
			strings.HasPrefix(line, " engine_wear_unfixable:"),
			strings.HasPrefix(line, " transmission_wear:"),
			strings.HasPrefix(line, " transmission_wear_unfixable:"),
			strings.HasPrefix(line, " cabin_wear:"),
			strings.HasPrefix(line, " cabin_wear_unfixable:"),
			strings.HasPrefix(line, " wheels_wear:"),
			strings.HasPrefix(line, " wheels_wear_unfixable:"):
			siiArray[i] = strings.Split(line, ":")[0] + ": 0"

		case strings.HasPrefix(line, " wheels_wear["),
			strings.HasPrefix(line, " wheels_wear_unfixable["):
			siiArray[i] = ""

		case strings.HasPrefix(line, " fuel_relative:") && refuel && i >= 7 && strings.Contains(siiArray[i-7], myTruckNameless):
			siiArray[i] = " fuel_relative: " + refuelLevel
		}
	}

	return strings.Join(siiArray, "\n")
}
