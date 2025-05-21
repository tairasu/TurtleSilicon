package main

import (
	"turtlesilicon/pkg/ui"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const appVersion = "1.0.6"

func main() {
	myApp := app.NewWithID("com.tairasu.turtlesilicon")
	myWindow := myApp.NewWindow("TurtleSilicon v" + appVersion)
	myWindow.Resize(fyne.NewSize(650, 450))
	myWindow.SetFixedSize(true)

	// Check for updates
	go func() {
		prefs, _ := utils.LoadPrefs()
		latest, notes, update, err := utils.CheckForUpdate(appVersion)
		if err == nil && update && latest != appVersion && prefs.SuppressedUpdateVersion != latest {
			checkbox := widget.NewCheck("Do not show this anymore", func(bool) {})
			content := container.NewVBox(
				widget.NewLabel("A new version ("+latest+") is available!"),
				widget.NewLabel("Release notes:\n\n"+notes),
				checkbox,
			)
			dialog.ShowCustomConfirm("Update Available", "OK", "Cancel", content, func(ok bool) {
				if checkbox.Checked {
					prefs.SuppressedUpdateVersion = latest
					utils.SavePrefs(prefs)
				}
			}, myWindow)
		}
	}()

	content := ui.CreateUI(myWindow)
	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}
