package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/djohts/tpc-truckersmp/config"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/updater"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/djohts/tpc-truckersmp/watcher"
)

func main() {
	log.SetTimeFormat("15:04:05")
	fmt.Println("tpc-truckersmp", "version", constants.APP_VERSION, "by djohts")

	log.Info("================= TPC For TruckersMP =================")
	log.Info("GitHub: https://github.com/djohts/tpc-truckersmp      ")
	log.Info("Usage: 0. Type g_debug_camera 1 in console (only once)")
	log.Info("       1. Alt + F12 to save coordinate of freecam     ")
	log.Info("       2. Make a quicksave & reload it ~1 second later")
	log.Info("            Note! Making a quicksave is not required  ")
	log.Info("                    when using auto mode              ")
	log.Info("======================================================")

	log.Info("Checking for updates...")
	needsUpdate, latest, err := updater.CheckUpdates()
	if err != nil {
		log.Error("Failed to check for updates", "error", err)
	} else if needsUpdate {
		log.Info("New version available", "current", constants.APP_VERSION, "latest", latest[1:])

		var update bool
		form := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("Update?").Description("Do you want to update to the latest version?").Value(&update)))
		err := form.Run()
		utils.HandleError(err)

		if update {
			log.Info("Updating to the latest version...")
			updated, err := updater.UpdateSelf()
			if !updated {
				log.Error("Failed to update. Please download latest release from GitHub manually and DM me on Discord if you wish this to be fixed.", "error", err)
			} else {
				log.Info("Updated successfully.")
				fmt.Printf("Press Enter to exit...")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
				os.Exit(0)
			}
		}
	} else {
		log.Info("You are using the latest version")
	}

	err = config.Init()
	utils.HandleError(err)

	DocumentsPath := &constants.DocumentsPath
	documents, err := utils.GetDocumentsPath()
	utils.HandleError(err)
	*DocumentsPath = documents

	if config.Get().Debug {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug mode enabled")
		log.SetTimeFormat("15:04:05.000")
	}

	if config.Get().Auto && config.Get().Keybinds.Quicksave == "" {
		utils.HandleError(errors.New("set keybinds.quicksave in config.yaml to use auto mode"))
	}

	watcher.Init()
}
