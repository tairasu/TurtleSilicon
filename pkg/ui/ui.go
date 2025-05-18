package ui

import (
	"log"
	"os" // Added import for os.ReadFile
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"turtlesilicon/pkg/launcher" // Corrected import path
	"turtlesilicon/pkg/patching" // Corrected import path
	"turtlesilicon/pkg/paths"    // Corrected import path
	"turtlesilicon/pkg/utils"    // Corrected import path
)

var (
	crossoverPathLabel   *widget.RichText
	turtlewowPathLabel   *widget.RichText
	turtlewowStatusLabel *widget.RichText
	crossoverStatusLabel *widget.RichText
	launchButton         *widget.Button
	patchTurtleWoWButton *widget.Button
	patchCrossOverButton *widget.Button
	metalHudCheckbox     *widget.Check
)

func UpdateAllStatuses() {
	// Update Crossover Path and Status
	if paths.CrossoverPath == "" {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		paths.PatchesAppliedCrossOver = false // Reset if path is cleared
	} else {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: paths.CrossoverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		wineloader2Path := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
		if utils.PathExists(wineloader2Path) {
			paths.PatchesAppliedCrossOver = true
		} else {
			// paths.PatchesAppliedCrossOver = false // Only set to false if not already true from a patch action this session
		}
	}
	crossoverPathLabel.Refresh()

	if paths.PatchesAppliedCrossOver {
		crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		if patchCrossOverButton != nil {
			patchCrossOverButton.Disable()
		}
	} else {
		crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		if patchCrossOverButton != nil {
			if paths.CrossoverPath != "" {
				patchCrossOverButton.Enable()
			} else {
				patchCrossOverButton.Disable()
			}
		}
	}
	crossoverStatusLabel.Refresh()

	// Update TurtleWoW Path and Status
	if paths.TurtlewowPath == "" {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		paths.PatchesAppliedTurtleWoW = false // Reset if path is cleared
	} else {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: paths.TurtlewowPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		winerosettaDllPath := filepath.Join(paths.TurtlewowPath, "winerosetta.dll")
		d3d9DllPath := filepath.Join(paths.TurtlewowPath, "d3d9.dll")
		rosettaX87DirPath := filepath.Join(paths.TurtlewowPath, "rosettax87")
		dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")
		rosettaX87ExePath := filepath.Join(rosettaX87DirPath, "rosettax87")
		libRuntimeRosettaX87Path := filepath.Join(rosettaX87DirPath, "libRuntimeRosettax87")

		dllsFileValid := false
		if utils.PathExists(dllsTextFile) {
			if fileContent, err := os.ReadFile(dllsTextFile); err == nil {
				if strings.Contains(string(fileContent), "winerosetta.dll") {
					dllsFileValid = true
				}
			}
		}

		if utils.PathExists(winerosettaDllPath) && utils.PathExists(d3d9DllPath) && utils.DirExists(rosettaX87DirPath) &&
			utils.PathExists(rosettaX87ExePath) && utils.PathExists(libRuntimeRosettaX87Path) && dllsFileValid {
			paths.PatchesAppliedTurtleWoW = true
		} else {
			// paths.PatchesAppliedTurtleWoW = false
		}
	}
	turtlewowPathLabel.Refresh()

	if paths.PatchesAppliedTurtleWoW {
		turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		if patchTurtleWoWButton != nil {
			patchTurtleWoWButton.Disable()
		}
	} else {
		turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		if patchTurtleWoWButton != nil {
			if paths.TurtlewowPath != "" {
				patchTurtleWoWButton.Enable()
			} else {
				patchTurtleWoWButton.Disable()
			}
		}
	}
	turtlewowStatusLabel.Refresh()

	// Update Launch Button State
	if launchButton != nil {
		if paths.PatchesAppliedTurtleWoW && paths.PatchesAppliedCrossOver && paths.TurtlewowPath != "" && paths.CrossoverPath != "" {
			launchButton.Enable()
		} else {
			launchButton.Disable()
		}
	}
}

func CreateUI(myWindow fyne.Window) fyne.CanvasObject {
	crossoverPathLabel = widget.NewRichText()
	turtlewowPathLabel = widget.NewRichText()
	turtlewowStatusLabel = widget.NewRichText()
	crossoverStatusLabel = widget.NewRichText()

	metalHudCheckbox = widget.NewCheck("Enable Metal Hud (show FPS)", func(checked bool) {
		launcher.EnableMetalHud = checked
		log.Printf("Metal HUD enabled: %v", launcher.EnableMetalHud)
	})
	metalHudCheckbox.SetChecked(launcher.EnableMetalHud)

	patchTurtleWoWButton = widget.NewButton("Patch TurtleWoW", func() {
		patching.PatchTurtleWoW(myWindow, UpdateAllStatuses)
	})
	patchCrossOverButton = widget.NewButton("Patch CrossOver", func() {
		patching.PatchCrossOver(myWindow, UpdateAllStatuses)
	})
	launchButton = widget.NewButton("Launch Game", func() {
		launcher.LaunchGame(myWindow)
	})

	paths.CheckDefaultCrossOverPath()

	pathSelectionForm := widget.NewForm(
		widget.NewFormItem("CrossOver Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			paths.SelectCrossOverPath(myWindow, crossoverPathLabel, UpdateAllStatuses)
		}), crossoverPathLabel)),
		widget.NewFormItem("TurtleWoW Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			paths.SelectTurtleWoWPath(myWindow, turtlewowPathLabel, UpdateAllStatuses)
		}), turtlewowPathLabel)),
	)

	patchOperationsLayout := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			widget.NewLabel("TurtleWoW Patch:"), turtlewowStatusLabel, patchTurtleWoWButton,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("CrossOver Patch:"), crossoverStatusLabel, patchCrossOverButton,
		),
		widget.NewSeparator(),
	)

	UpdateAllStatuses() // Initial UI state update

	return container.NewVBox(
		pathSelectionForm,
		patchOperationsLayout,
		metalHudCheckbox,
		container.NewPadded(launchButton),
	)
}
