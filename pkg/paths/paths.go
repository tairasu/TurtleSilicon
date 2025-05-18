package paths

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"turtlesilicon/pkg/utils"
)

const DefaultCrossOverPath = "/Applications/CrossOver.app"

var (
	CrossoverPath           string
	TurtlewowPath           string
	PatchesAppliedTurtleWoW = false
	PatchesAppliedCrossOver = false
)

func SelectCrossOverPath(myWindow fyne.Window, crossoverPathLabel *widget.RichText, updateAllStatuses func()) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			log.Println("CrossOver path selection cancelled.")
			updateAllStatuses()
			return
		}
		selectedPath := uri.Path()
		if filepath.Ext(selectedPath) == ".app" && utils.DirExists(selectedPath) {
			CrossoverPath = selectedPath
			PatchesAppliedCrossOver = false
			log.Println("CrossOver path set to:", CrossoverPath)
		} else {
			dialog.ShowError(fmt.Errorf("invalid selection: '%s'. Please select a valid .app bundle", selectedPath), myWindow)
			log.Println("Invalid CrossOver path selected:", selectedPath)
		}
		updateAllStatuses()
	}, myWindow)
}

func SelectTurtleWoWPath(myWindow fyne.Window, turtlewowPathLabel *widget.RichText, updateAllStatuses func()) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			log.Println("TurtleWoW path selection cancelled.")
			updateAllStatuses()
			return
		}
		selectedPath := uri.Path()
		if utils.DirExists(selectedPath) {
			TurtlewowPath = selectedPath
			PatchesAppliedTurtleWoW = false
			log.Println("TurtleWoW path set to:", TurtlewowPath)
		} else {
			dialog.ShowError(fmt.Errorf("invalid selection: '%s' is not a valid directory", selectedPath), myWindow)
			log.Println("Invalid TurtleWoW path selected:", selectedPath)
		}
		updateAllStatuses()
	}, myWindow)
}

func UpdatePathLabels(crossoverPathLabel, turtlewowPathLabel *widget.RichText) {
	if CrossoverPath == "" {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
	} else {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: CrossoverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
	}
	crossoverPathLabel.Refresh()

	if TurtlewowPath == "" {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
	} else {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: TurtlewowPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
	}
	turtlewowPathLabel.Refresh()
}

func CheckDefaultCrossOverPath() {
	if info, err := os.Stat(DefaultCrossOverPath); err == nil && info.IsDir() {
		CrossoverPath = DefaultCrossOverPath
		log.Println("Pre-set CrossOver to default:", DefaultCrossOverPath)
	}
}
