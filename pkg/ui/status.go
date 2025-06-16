package ui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	pulsingActive = false
)

// UpdateAllStatuses updates all UI components based on current application state
func UpdateAllStatuses() {
	updateCrossoverStatus()
	updateTurtleWoWStatus()
	updatePlayButtonState()
	updateServiceStatus()

	// Update Wine registry status if components are initialized
	if optionAsAltStatusLabel != nil {
		updateWineRegistryStatus()
	}

	// Update recommended settings button if component is initialized
	if applyRecommendedSettingsButton != nil {
		updateRecommendedSettingsButton()
	}
}

// updateCrossoverStatus updates CrossOver path and patch status
func updateCrossoverStatus() {
	if paths.CrossoverPath == "" {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		paths.PatchesAppliedCrossOver = false // Reset if path is cleared
	} else {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: paths.CrossoverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		wineloader2Path := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
		if utils.PathExists(wineloader2Path) {
			paths.PatchesAppliedCrossOver = true
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
}

// updateTurtleWoWStatus updates TurtleWoW path and patch status
func updateTurtleWoWStatus() {
	if paths.TurtlewowPath == "" {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		paths.PatchesAppliedTurtleWoW = false // Reset if path is cleared
	} else {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: paths.TurtlewowPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}

		// Check if all required files exist
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

		// Check if patched files have the correct size (matches bundled versions)
		winerosettaDllCorrectSize := utils.CompareFileWithBundledResource(winerosettaDllPath, "winerosetta/winerosetta.dll")
		d3d9DllCorrectSize := utils.CompareFileWithBundledResource(d3d9DllPath, "winerosetta/d3d9.dll")
		libSiliconPatchCorrectSize := utils.CompareFileWithBundledResource(libSiliconPatchDllPath, "winerosetta/libSiliconPatch.dll")
		rosettaX87CorrectSize := utils.CompareFileWithBundledResource(rosettaX87ExePath, "rosettax87/rosettax87")
		libRuntimeRosettaX87CorrectSize := utils.CompareFileWithBundledResource(libRuntimeRosettaX87Path, "rosettax87/libRuntimeRosettax87")

		// Check if shadowLOD setting is applied
		shadowLODApplied := patching.CheckShadowLODSetting()

		if utils.PathExists(winerosettaDllPath) && utils.PathExists(d3d9DllPath) && utils.PathExists(libSiliconPatchDllPath) &&
			utils.DirExists(rosettaX87DirPath) && utils.PathExists(rosettaX87ExePath) &&
			utils.PathExists(libRuntimeRosettaX87Path) && dllsFileValid &&
			winerosettaDllCorrectSize && d3d9DllCorrectSize && libSiliconPatchCorrectSize &&
			rosettaX87CorrectSize && libRuntimeRosettaX87CorrectSize && shadowLODApplied {
			paths.PatchesAppliedTurtleWoW = true
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
}

// updatePlayButtonState enables/disables play and launch buttons based on current state
func updatePlayButtonState() {
	launchEnabled := paths.PatchesAppliedTurtleWoW && paths.PatchesAppliedCrossOver &&
		paths.TurtlewowPath != "" && paths.CrossoverPath != "" && service.IsServiceRunning()

	if launchButton != nil {
		if launchEnabled {
			launchButton.Enable()
		} else {
			launchButton.Disable()
		}
	}

	if playButton != nil && playButtonText != nil {
		if launchEnabled {
			playButton.Enable()
			// Update text to show enabled state with white color
			playButtonText.Segments = []widget.RichTextSegment{
				&widget.TextSegment{
					Text: "PLAY",
					Style: widget.RichTextStyle{
						SizeName:  theme.SizeNameHeadingText,
						ColorName: theme.ColorNameForegroundOnPrimary,
					},
				},
			}
		} else {
			playButton.Disable()
			// Update text to show disabled state with dimmed color and different text
			playButtonText.Segments = []widget.RichTextSegment{
				&widget.TextSegment{
					Text: "PLAY",
					Style: widget.RichTextStyle{
						SizeName:  theme.SizeNameHeadingText,
						ColorName: theme.ColorNameDisabled,
					},
				},
			}
		}
		playButtonText.Refresh()
	}
}

// updateServiceStatus updates RosettaX87 service status and related buttons
func updateServiceStatus() {
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
