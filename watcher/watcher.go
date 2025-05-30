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
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/fsnotify/fsnotify"
)

var (
	profileList   []string
	watchPathList []string

	keyboard *sendkeys.KBWrap
)

func Init() {
	utils.EnsureDecrypt()

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
	if ets2Path := filepath.Join(constants.DocumentsPath, constants.ETS); utils.IsDir(ets2Path) &&
		utils.IsFile(filepath.Join(ets2Path, `cams.txt`)) {
		watchPathList = append(watchPathList, filepath.Join(ets2Path, `cams.txt`))
	}

	if atsPath := filepath.Join(constants.DocumentsPath, constants.ATS); utils.IsDir(atsPath) &&
		utils.IsFile(filepath.Join(atsPath, `cams.txt`)) {
		watchPathList = append(watchPathList, filepath.Join(atsPath, `cams.txt`))
	}
}

func addSaveWatchers() {
	for _, profilePath := range profileList {
		if utils.IsDir(filepath.Join(profilePath, `save`)) {
			watchPathList = append(watchPathList, filepath.Join(profilePath, `save`))
		}
	}

	profileList = profileList[0:0]
}

func listProfiles(path string, f fs.DirEntry, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() && filepath.Base(filepath.Dir(path)) == `profiles` {
		profileList = append(profileList, path)
	}

	return nil
}

func getProfileList() error {
	ets2Path, err := getEts2Path()
	if err == nil {
		err := filepath.WalkDir(ets2Path, listProfiles)
		if err != nil {
			return err
		}
	}

	atsPath, err := getAtsPath()
	if err == nil {
		err := filepath.WalkDir(atsPath, listProfiles)
		if err != nil {
			return err
		}
	}

	return nil
}

func getEts2Path() (string, error) {
	ets2Path := filepath.Join(constants.DocumentsPath, constants.ETS)
	if utils.IsDir(ets2Path) {
		return ets2Path, nil
	}

	return "", errors.New("ETS2 not found")
}

func getAtsPath() (string, error) {
	atsPath := filepath.Join(constants.DocumentsPath, constants.ATS)
	if utils.IsDir(atsPath) {
		return atsPath, nil
	}

	return "", errors.New("ATS not found")
}

func decryptSii(filePath string) (bool, error) {
	decryptFile := utils.EnsureDecrypt()
	execPath, err := filepath.Abs(decryptFile.Name())
	utils.HandleError(err)

	cmd := exec.Command(execPath, filePath)
	log.Debug("Decrypting " + utils.FormatPath(filePath, constants.DocumentsPath))
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
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	output := make([]string, 0)
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	return output, nil
}

func writeFile(filePath string, outPut string) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(outPut)
	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func addPathToWatch(watch *fsnotify.Watcher) error {
	for _, watchPath := range watchPathList {
		err := watch.Add(watchPath)
		if err != nil {
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

				if base == `quicksave` {
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
					done, err := flushChange(filepath.Join(ev.Name, `game.sii`))
					flushWatch.Stop()
					utils.HandleError(err)

					if done {
						log.Info("Updated " + utils.FormatPath(filepath.Join(ev.Name, `game.sii`), constants.DocumentsPath) + " (" + fmt.Sprint(flushWatch.Milliseconds().Nanoseconds()) + "ms)")
					}
				} else if base == `cams.txt` {
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

	camsPath := filepath.Join(constants.DocumentsPath, constants.ETS, `cams.txt`)
	if strings.Contains(filePath, constants.ATS) {
		camsPath = filepath.Join(constants.DocumentsPath, constants.ATS, `cams.txt`)
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
	camCoordinate := strings.ReplaceAll(cams[len(cams)-1], `;`, `,`)
	location := strings.Split(camCoordinate, ` , `)[1]
	rotation := strings.Split(camCoordinate, ` , `)[2]
	rotation = strings.Replace(rotation, `,`, `;`, 1)
	log.Info("Target: " + `(` + location + `) (` + rotation + `)`)

	return location, rotation
}

func editSii(siiArray []string, location string, rotation string) string {
	attachTrailer := config.Get().Features.AttachTrailer
	refuel := config.Get().Features.Refuel
	teleport := config.Get().Features.Teleport
	myTruckNameless := ""
	attachTrailerState := 0

	for i := range siiArray {
		if strings.HasPrefix(siiArray[i], " my_truck: _nameless") {
			myTruckNameless = strings.Split(siiArray[i], `: `)[1]
		} else if strings.HasPrefix(siiArray[i], " assigned_trailer: _nameless") && attachTrailer {
			attachTrailerState = 1
		} else if strings.HasPrefix(siiArray[i], " assigned_trailer_connected: false") && attachTrailerState == 1 {
			attachTrailerState = 2
			siiArray[i] = " assigned_trailer_connected: true"
		} else if strings.HasPrefix(siiArray[i], " nav_node_position:") && attachTrailerState == 2 {
			siiArray[i] = " nav_node_position: (0, 0, 0)"
		} else if strings.HasPrefix(siiArray[i], " truck_placement:") && teleport {
			siiArray[i] = " truck_placement: " + `(` + location + `) (` + rotation + `)`
		} else if strings.HasPrefix(siiArray[i], " trailer_placement:") && (attachTrailerState == 2 || teleport) {
			siiArray[i] = ` trailer_placement: (0, 0, 0) (` + rotation + `)`
		} else if strings.HasPrefix(siiArray[i], " slave_trailer_placements[") && (attachTrailerState == 2 || teleport) {
			siiArray[i] = strings.Split(siiArray[i], `: `)[0] + `: (0, 0, 0) (` + rotation + `)`
		} else if strings.HasPrefix(siiArray[i], " trailer_body_wear:") {
			siiArray[i] = " trailer_body_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " trailer_body_wear_unfixable:") {
			siiArray[i] = " trailer_body_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " chassis_wear:") {
			siiArray[i] = " chassis_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " chassis_wear_unfixable:") {
			siiArray[i] = " chassis_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " engine_wear:") {
			siiArray[i] = " engine_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " engine_wear_unfixable:") {
			siiArray[i] = " engine_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " transmission_wear:") {
			siiArray[i] = " transmission_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " transmission_wear_unfixable:") {
			siiArray[i] = " transmission_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " cabin_wear:") {
			siiArray[i] = " cabin_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " cabin_wear_unfixable:") {
			siiArray[i] = " cabin_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " wheels_wear:") {
			siiArray[i] = " wheels_wear: 0"
		} else if strings.HasPrefix(siiArray[i], " wheels_wear_unfixable:") {
			siiArray[i] = " wheels_wear_unfixable: 0"
		} else if strings.HasPrefix(siiArray[i], " wheels_wear[") {
			siiArray[i] = ""
		} else if strings.HasPrefix(siiArray[i], " wheels_wear_unfixable[") {
			siiArray[i] = ""
		} else if strings.HasPrefix(siiArray[i], " fuel_relative:") && strings.Contains(siiArray[i-7], myTruckNameless) && refuel {
			siiArray[i] = " fuel_relative: " + fmt.Sprintf("%d", config.Get().Features.RefuelRelative)
		}
	}

	return strings.Join(siiArray, "\n")
}
