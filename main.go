package main

import (
	"errors"

	"github.com/charmbracelet/log"
	"github.com/djohts/tpc-truckersmp/config"
	"github.com/djohts/tpc-truckersmp/constants"
	"github.com/djohts/tpc-truckersmp/utils"
	"github.com/djohts/tpc-truckersmp/watcher"
)

func main() {
	if !utils.IsFile("SII_Decrypt.exe") {
		utils.HandleError(errors.New("SII_Decrypt.exe does not exist"))
	}

	log.Info("================= TPC For TruckersMP =================")
	log.Info("Usage: 0. Type g_debug_camera 1 in console (only once)")
	log.Info("       1. Alt+F12 to save coordinate of freecam")
	log.Info("       2. Make a quicksave & reload 1-2 seconds later")
	log.Info("======================================================")

	err := config.Init()
	utils.HandleError(err)

	DocumentsPath := &constants.DocumentsPath
	documents, err := utils.GetDocumentsPath()
	utils.HandleError(err)
	*DocumentsPath = documents

	if config.Get().Debug {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	}

	if config.Get().Auto && config.Get().Keybinds.Quicksave == "" {
		utils.HandleError(errors.New("set keybinds.quicksave in config.yaml to use auto mode"))
	}

	watcher.Init()
}
