package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"turtlesilicon/pkg/ui" // Updated import path
)

const appVersion = "1.0.4" 

func main() {
	myApp := app.NewWithID("com.tairasu.turtlesilicon")
	myWindow := myApp.NewWindow("TurtleSilicon v" + appVersion) // Updated title
	myWindow.Resize(fyne.NewSize(650, 450))
	myWindow.SetFixedSize(true)

	content := ui.CreateUI(myWindow) // Use the CreateUI function from the ui package
	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}
