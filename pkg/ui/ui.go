package ui

import (
	"log"
	"net/url"
	"os" // Added import for os.ReadFile
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"turtlesilicon/pkg/launcher" // Corrected import path
	"turtlesilicon/pkg/patching" // Corrected import path
	"turtlesilicon/pkg/paths"    // Corrected import path
	"turtlesilicon/pkg/utils"    // Corrected import path
)

var (
	crossoverPathLabel      *widget.RichText
	turtlewowPathLabel      *widget.RichText
	turtlewowStatusLabel    *widget.RichText
	crossoverStatusLabel    *widget.RichText
	launchButton            *widget.Button
	patchTurtleWoWButton    *widget.Button
	patchCrossOverButton    *widget.Button
	unpatchTurtleWoWButton  *widget.Button
	unpatchCrossOverButton  *widget.Button
	metalHudCheckbox        *widget.Check
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
		if unpatchCrossOverButton != nil {
			unpatchCrossOverButton.Enable()
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
		if unpatchCrossOverButton != nil {
			unpatchCrossOverButton.Disable()
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
		libSiliconPatchDllPath := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
		rosettaX87DirPath := filepath.Join(paths.TurtlewowPath, "rosettax87")
		dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")
		rosettaX87ExePath := filepath.Join(rosettaX87DirPath, "rosettax87")
		libRuntimeRosettaX87Path := filepath.Join(rosettaX87DirPath, "libRuntimeRosettax87")

		dllsFileValid := false
		if utils.PathExists(dllsTextFile) {
			if fileContent, err := os.ReadFile(dllsTextFile); err == nil {
				contentStr := string(fileContent)
				if strings.Contains(contentStr, "winerosetta.dll") && strings.Contains(contentStr, "libSiliconPatch.dll") {
					dllsFileValid = true
				}
			}
		}

		if utils.PathExists(winerosettaDllPath) && utils.PathExists(d3d9DllPath) && utils.PathExists(libSiliconPatchDllPath) && 
			utils.DirExists(rosettaX87DirPath) && utils.PathExists(rosettaX87ExePath) && 
			utils.PathExists(libRuntimeRosettaX87Path) && dllsFileValid {
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
		if unpatchTurtleWoWButton != nil {
			unpatchTurtleWoWButton.Enable()
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
		if unpatchTurtleWoWButton != nil {
			unpatchTurtleWoWButton.Disable()
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

	// Load the application logo
	logoResource, err := fyne.LoadResourceFromPath("Icon.png")
	if err != nil {
		log.Printf("Warning: could not load logo: %v", err)
	}
	
	// Create the logo image with a fixed size
	var logoImage *canvas.Image
	if logoResource != nil {
		logoImage = canvas.NewImageFromResource(logoResource)
		logoImage.FillMode = canvas.ImageFillContain
		logoImage.SetMinSize(fyne.NewSize(100, 100))
	}
	
	// Create a container to center the logo
	var logoContainer fyne.CanvasObject
	if logoImage != nil {
		logoContainer = container.NewCenter(logoImage)
	} else {
		// If logo couldn't be loaded, add an empty space for consistent layout
		logoContainer = container.NewCenter(widget.NewLabel(""))
	}

	metalHudCheckbox = widget.NewCheck("Enable Metal Hud (show FPS)", func(checked bool) {
		launcher.EnableMetalHud = checked
		log.Printf("Metal HUD enabled: %v", launcher.EnableMetalHud)
	})
	metalHudCheckbox.SetChecked(launcher.EnableMetalHud)

	patchTurtleWoWButton = widget.NewButton("Patch TurtleWoW", func() {
		patching.PatchTurtleWoW(myWindow, UpdateAllStatuses)
	})
	unpatchTurtleWoWButton = widget.NewButton("Unpatch TurtleWoW", func() {
		patching.UnpatchTurtleWoW(myWindow, UpdateAllStatuses)
	})
	patchCrossOverButton = widget.NewButton("Patch CrossOver", func() {
		patching.PatchCrossOver(myWindow, UpdateAllStatuses)
	})
	unpatchCrossOverButton = widget.NewButton("Unpatch CrossOver", func() {
		patching.UnpatchCrossOver(myWindow, UpdateAllStatuses)
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
		container.NewGridWithColumns(4,
			widget.NewLabel("TurtleWoW Patch:"), turtlewowStatusLabel, patchTurtleWoWButton, unpatchTurtleWoWButton,
		),
		container.NewGridWithColumns(4,
			widget.NewLabel("CrossOver Patch:"), crossoverStatusLabel, patchCrossOverButton, unpatchCrossOverButton,
		),
		widget.NewSeparator(),
	)

	UpdateAllStatuses() // Initial UI state update

	// Create GitHub link
	githubURL := "https://github.com/tairasu/TurtleSilicon"
	parsedURL, err := url.Parse(githubURL)
	if err != nil {
		log.Printf("Error parsing GitHub URL: %v", err)
	}
	githubLink := widget.NewHyperlink("GitHub Repository", parsedURL)
	githubContainer := container.NewCenter(githubLink)

	return container.NewVBox(
		logoContainer,
		pathSelectionForm,
		patchOperationsLayout,
		metalHudCheckbox,
		container.NewPadded(launchButton),
		widget.NewSeparator(),
		githubContainer,
	)
}
