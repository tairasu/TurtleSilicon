package main

import (
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/ui"
	"turtlesilicon/pkg/utils"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const appVersion = "1.2.1"

func main() {
	TSApp := app.NewWithID("com.tairasu.turtlesilicon")
	TSWindow := TSApp.NewWindow("TurtleSilicon v" + appVersion)
	TSWindow.Resize(fyne.NewSize(650, 500))
	TSWindow.SetFixedSize(true)

	// Check for updates
	go func() {
		prefs, _ := utils.LoadPrefs()
		latest, notes, update, err := utils.CheckForUpdate(appVersion)
		debug.Printf("DEBUG RAW: latest=%q", latest)
		latestVersion := strings.TrimLeft(latest, "v.")
		debug.Printf("DEBUG: appVersion=%q, latest=%q, latestVersion=%q, suppressed=%q, update=%v, err=%v\n",
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
			}, TSWindow)
		}
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
