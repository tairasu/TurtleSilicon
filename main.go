package main

import (
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/ui"
	"turtlesilicon/pkg/utils"

	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const appVersion = "1.1.1"

func main() {
	myApp := app.NewWithID("com.tairasu.turtlesilicon")
	myWindow := myApp.NewWindow("TurtleSilicon v" + appVersion)
	myWindow.Resize(fyne.NewSize(650, 500))
	myWindow.SetFixedSize(true)

	// Check for updates
	go func() {
		prefs, _ := utils.LoadPrefs()
		latest, notes, update, err := utils.CheckForUpdate(appVersion)
		log.Printf("DEBUG RAW: latest=%q", latest)
		latestVersion := strings.TrimLeft(latest, "v.")
		log.Printf("DEBUG: appVersion=%q, latest=%q, latestVersion=%q, suppressed=%q, update=%v, err=%v\n",
			appVersion, latest, latestVersion, prefs.SuppressedUpdateVersion, update, err)
		// Always skip popup if versions match
		if latestVersion == appVersion {
			return
		}
		if err == nil && update && prefs.SuppressedUpdateVersion != latestVersion {
			checkbox := widget.NewCheck("Do not show this anymore", func(bool) {})
			content := container.NewVBox(
				widget.NewLabel("A new version ("+latestVersion+") is available!"),
				widget.NewLabel("Release notes:\n\n"+notes),
				checkbox,
			)
			dialog.ShowCustomConfirm("Update Available", "OK", "Cancel", content, func(ok bool) {
				if checkbox.Checked {
					prefs.SuppressedUpdateVersion = latestVersion
					utils.SavePrefs(prefs)
				}
			}, myWindow)
		}
	}()

	content := ui.CreateUI(myWindow)
	myWindow.SetContent(content)

	// Set up cleanup when window closes
	myWindow.SetCloseIntercept(func() {
		log.Println("Application closing, cleaning up RosettaX87 service...")
		service.CleanupService()
		myApp.Quit()
	})

	myWindow.ShowAndRun()
}
