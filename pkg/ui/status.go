package ui

import (
	"path/filepath"
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
	updateVersionStatus()
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

// updateVersionStatus updates status for the current version
func updateVersionStatus() {
	currentVer := GetCurrentVersion()
	if currentVer == nil {
		// Fall back to old system
		updateCrossoverStatus()
		updateTurtleWoWStatus()
		return
	}

	// Update CrossOver status for current version
	if currentVer.CrossOverPath == "" {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		// Disable both CrossOver buttons when path not set
		if patchCrossOverButton != nil {
			patchCrossOverButton.Disable()
		}
		if unpatchCrossOverButton != nil {
			unpatchCrossOverButton.Disable()
		}
	} else {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: currentVer.CrossOverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
		wineloader2Path := filepath.Join(currentVer.CrossOverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
		if utils.PathExists(wineloader2Path) {
			crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
			gamePatched, _ := paths.GetVersionPatchingStatus(currentVer.ID)
			paths.SetVersionPatchingStatus(currentVer.ID, gamePatched, true)
			// Update CrossOver button states - patches applied
			if patchCrossOverButton != nil {
				patchCrossOverButton.Disable()
			}
			if unpatchCrossOverButton != nil {
				unpatchCrossOverButton.Enable()
			}
		} else {
			crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			gamePatched, _ := paths.GetVersionPatchingStatus(currentVer.ID)
			paths.SetVersionPatchingStatus(currentVer.ID, gamePatched, false)
			// Update CrossOver button states - patches not applied
			if patchCrossOverButton != nil {
				patchCrossOverButton.Enable()
			}
			if unpatchCrossOverButton != nil {
				unpatchCrossOverButton.Disable()
			}
		}
	}
	crossoverPathLabel.Refresh()
	crossoverStatusLabel.Refresh()

	// Update Game status for current version
	if currentVer.GamePath == "" {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
		// Disable both buttons when path not set
		if patchTurtleWoWButton != nil {
			patchTurtleWoWButton.Disable()
		}
		if unpatchTurtleWoWButton != nil {
			unpatchTurtleWoWButton.Disable()
		}
	} else {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: currentVer.GamePath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}

		// Check if patches are Applied using version-aware checking
		patchesApplied := patching.CheckVersionPatchingStatus(currentVer.GamePath, currentVer.UsesRosettaPatching, currentVer.UsesDivxDecoderPatch)
		if patchesApplied {
			turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
			// Update button states - patches applied
			if patchTurtleWoWButton != nil {
				patchTurtleWoWButton.Disable()
			}
			if unpatchTurtleWoWButton != nil {
				unpatchTurtleWoWButton.Enable()
			}
		} else {
			turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not Applied", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			// Update button states - patches not applied
			if patchTurtleWoWButton != nil {
				patchTurtleWoWButton.Enable()
			}
			if unpatchTurtleWoWButton != nil {
				unpatchTurtleWoWButton.Disable()
			}
		}
	}
	turtlewowPathLabel.Refresh()
	turtlewowStatusLabel.Refresh()
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

		// Use version-aware checking for legacy fallback
		patchesApplied := patching.CheckVersionPatchingStatus(paths.TurtlewowPath, true, false) // Assuming legacy is rosetta patching
		paths.PatchesAppliedTurtleWoW = patchesApplied
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
	launchEnabled := false

	// Use version-aware checking
	currentVer := GetCurrentVersion()
	if currentVer != nil {
		// Check if both game and CrossOver paths are set
		gamePatchesApplied := currentVer.GamePath != "" && patching.CheckVersionPatchingStatus(currentVer.GamePath, currentVer.UsesRosettaPatching, currentVer.UsesDivxDecoderPatch)

		// Check CrossOver status
		crossoverPatchesApplied := false
		if currentVer.CrossOverPath != "" {
			wineloader2Path := filepath.Join(currentVer.CrossOverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
			crossoverPatchesApplied = utils.PathExists(wineloader2Path)
		}

		launchEnabled = gamePatchesApplied && crossoverPatchesApplied
	} else {
		// Fallback to legacy system
		launchEnabled = paths.PatchesAppliedTurtleWoW && paths.PatchesAppliedCrossOver &&
			paths.TurtlewowPath != "" && paths.CrossoverPath != ""
	}

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
			// Check if we can enable the service button - use version-aware checking
			currentVer := GetCurrentVersion()
			canStartService := false

			if currentVer != nil && currentVer.GamePath != "" {
				// Check if patches are applied for current version
				patchesApplied := patching.CheckVersionPatchingStatus(currentVer.GamePath, currentVer.UsesRosettaPatching, currentVer.UsesDivxDecoderPatch)
				canStartService = patchesApplied
			} else if paths.TurtlewowPath != "" {
				// Fallback to legacy system
				canStartService = paths.PatchesAppliedTurtleWoW
			}

			if canStartService {
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
