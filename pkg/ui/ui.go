package ui

import (
	"log"
	"net/url"
	"os" // Added import for os.ReadFile
	"path/filepath"
	"strings"
	"time"

	"turtlesilicon/pkg/launcher" // Corrected import path
	"turtlesilicon/pkg/patching" // Corrected import path
	"turtlesilicon/pkg/paths"    // Corrected import path
	"turtlesilicon/pkg/service"  // Added service import
	"turtlesilicon/pkg/utils"    // Corrected import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	crossoverPathLabel     *widget.RichText
	turtlewowPathLabel     *widget.RichText
	turtlewowStatusLabel   *widget.RichText
	crossoverStatusLabel   *widget.RichText
	serviceStatusLabel     *widget.RichText
	launchButton           *widget.Button
	patchTurtleWoWButton   *widget.Button
	patchCrossOverButton   *widget.Button
	unpatchTurtleWoWButton *widget.Button
	unpatchCrossOverButton *widget.Button
	startServiceButton     *widget.Button
	stopServiceButton      *widget.Button
	metalHudCheckbox       *widget.Check
	showTerminalCheckbox   *widget.Check
	vanillaTweaksCheckbox  *widget.Check
	envVarsEntry           *widget.Entry
	pulsingActive          = false
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
		// Now requires service to be running as well
		if paths.PatchesAppliedTurtleWoW && paths.PatchesAppliedCrossOver &&
			paths.TurtlewowPath != "" && paths.CrossoverPath != "" && service.IsServiceRunning() {
			launchButton.Enable()
		} else {
			launchButton.Disable()
		}
	}

	// Update Service Status
	if paths.ServiceStarting {
		// Show pulsing "Starting..." when service is starting
		if serviceStatusLabel != nil {
			if !pulsingActive {
				pulsingActive = true
				go startPulsingAnimation()
			}
		}
		if startServiceButton != nil {
			startServiceButton.Disable()
		}
		if stopServiceButton != nil {
			stopServiceButton.Disable()
		}
	} else if service.IsServiceRunning() {
		pulsingActive = false
		paths.RosettaX87ServiceRunning = true
		if serviceStatusLabel != nil {
			serviceStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Running", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
			serviceStatusLabel.Refresh()
		}
		if startServiceButton != nil {
			startServiceButton.Disable()
		}
		if stopServiceButton != nil {
			stopServiceButton.Enable()
		}
	} else {
		pulsingActive = false
		paths.RosettaX87ServiceRunning = false
		if serviceStatusLabel != nil {
			serviceStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Stopped", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			serviceStatusLabel.Refresh()
		}
		if startServiceButton != nil {
			if paths.TurtlewowPath != "" && paths.PatchesAppliedTurtleWoW {
				startServiceButton.Enable()
			} else {
				startServiceButton.Disable()
			}
		}
		if stopServiceButton != nil {
			stopServiceButton.Disable()
		}
	}
}

// startPulsingAnimation creates a pulsing effect for the "Starting..." text
func startPulsingAnimation() {
	dots := 0
	for pulsingActive && paths.ServiceStarting {
		var text string
		switch dots % 4 {
		case 0:
			text = "Starting"
		case 1:
			text = "Starting."
		case 2:
			text = "Starting.."
		case 3:
			text = "Starting..."
		}

		if serviceStatusLabel != nil {
			fyne.DoAndWait(func() {
				serviceStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: text, Style: widget.RichTextStyle{ColorName: theme.ColorNamePrimary}}}
				serviceStatusLabel.Refresh()
			})
		}

		time.Sleep(500 * time.Millisecond)
		dots++
	}
}

func CreateUI(myWindow fyne.Window) fyne.CanvasObject {
	// Load saved paths from prefs
	prefs, _ := utils.LoadPrefs()
	if prefs.TurtleWoWPath != "" {
		paths.TurtlewowPath = prefs.TurtleWoWPath
	}
	if prefs.CrossOverPath != "" {
		paths.CrossoverPath = prefs.CrossOverPath
	}

	crossoverPathLabel = widget.NewRichText()
	turtlewowPathLabel = widget.NewRichText()
	turtlewowStatusLabel = widget.NewRichText()
	crossoverStatusLabel = widget.NewRichText()
	serviceStatusLabel = widget.NewRichText()

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

	showTerminalCheckbox = widget.NewCheck("Show Terminal", func(checked bool) {
		// Save to preferences
		prefs, _ := utils.LoadPrefs()
		prefs.ShowTerminalNormally = checked
		utils.SavePrefs(prefs)
		log.Printf("Show terminal normally: %v", checked)
	})
	showTerminalCheckbox.SetChecked(prefs.ShowTerminalNormally)

	vanillaTweaksCheckbox = widget.NewCheck("Enable vanilla-tweaks", func(checked bool) {
		launcher.EnableVanillaTweaks = checked
		// Save to preferences
		prefs, _ := utils.LoadPrefs()
		prefs.EnableVanillaTweaks = checked
		utils.SavePrefs(prefs)
		log.Printf("Vanilla-tweaks enabled: %v", launcher.EnableVanillaTweaks)
	})
	vanillaTweaksCheckbox.SetChecked(prefs.EnableVanillaTweaks)
	launcher.EnableVanillaTweaks = prefs.EnableVanillaTweaks

	// Load environment variables from preferences
	if prefs.EnvironmentVariables != "" {
		launcher.CustomEnvVars = prefs.EnvironmentVariables
	}

	envVarsEntry = widget.NewEntry()
	envVarsEntry.SetPlaceHolder(`Custom environment variables`)
	envVarsEntry.SetText(launcher.CustomEnvVars)
	envVarsEntry.OnChanged = func(text string) {
		launcher.CustomEnvVars = text
		// Save to preferences
		prefs, _ := utils.LoadPrefs()
		prefs.EnvironmentVariables = text
		utils.SavePrefs(prefs)
		log.Printf("Environment variables updated: %v", launcher.CustomEnvVars)
	}

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
	startServiceButton = widget.NewButton("Start Service", func() {
		service.StartRosettaX87Service(myWindow, UpdateAllStatuses)
	})
	stopServiceButton = widget.NewButton("Stop Service", func() {
		service.StopRosettaX87Service(myWindow, UpdateAllStatuses)
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
		container.NewGridWithColumns(4,
			widget.NewLabel("RosettaX87 Service:"), serviceStatusLabel, startServiceButton, stopServiceButton,
		),
		widget.NewSeparator(),
	)

	UpdateAllStatuses() // Initial UI state update

	// Set up periodic status updates to keep service status in sync
	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second) // Check every 5 seconds
	// 		fyne.DoAndWait(func() {
	// 			UpdateAllStatuses()
	// 		})
	// 	}
	// }()

	// Create GitHub link
	githubURL := "https://github.com/tairasu/TurtleSilicon"
	parsedURL, err := url.Parse(githubURL)
	if err != nil {
		log.Printf("Error parsing GitHub URL: %v", err)
	}
	githubLink := widget.NewHyperlink("GitHub Repository", parsedURL)
	githubContainer := container.NewCenter(githubLink)

	return container.NewPadded(
		container.NewVBox(
			logoContainer,
			pathSelectionForm,
			patchOperationsLayout,
			container.NewGridWithColumns(3,
				metalHudCheckbox,
				showTerminalCheckbox,
				vanillaTweaksCheckbox,
			),
			widget.NewSeparator(),
			widget.NewLabel("Environment Variables:"),
			envVarsEntry,
			container.NewPadded(launchButton),
			widget.NewSeparator(),
			githubContainer,
		),
	)
}
