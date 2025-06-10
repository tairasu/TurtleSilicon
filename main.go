package main

import (
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/ui"
	"turtlesilicon/pkg/utils"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const appVersion = "1.2.2"

func main() {
	TSApp := app.NewWithID("com.tairasu.turtlesilicon")
	TSWindow := TSApp.NewWindow("TurtleSilicon v" + appVersion)
	TSWindow.Resize(fyne.NewSize(650, 500))
	TSWindow.SetFixedSize(true)

	// Check for updates
	go func() {
		prefs, _ := utils.LoadPrefs()
		updateInfo, updateAvailable, err := utils.CheckForUpdateWithAssets(appVersion)
		if err != nil {
			debug.Printf("Failed to check for updates: %v", err)
			return
		}

		if !updateAvailable {
			debug.Printf("No updates available")
			return
		}

		latestVersion := strings.TrimPrefix(updateInfo.TagName, "v")
		debug.Printf("Update available: current=%s, latest=%s", appVersion, latestVersion)

		// Skip if user has suppressed this version
		if prefs.SuppressedUpdateVersion == latestVersion {
			debug.Printf("Update suppressed by user: %s", latestVersion)
			return
		}

		// Show enhanced update dialog
		ui.ShowUpdateDialog(updateInfo, appVersion, TSWindow)
	}()

	content := ui.CreateUI(TSWindow)
	TSWindow.SetContent(content)

	// Set up cleanup when window closes
	TSWindow.SetCloseIntercept(func() {
		debug.Println("Application closing, cleaning up RosettaX87 service...")
		service.CleanupService()
		TSApp.Quit()
	})

	TSWindow.ShowAndRun()
}
