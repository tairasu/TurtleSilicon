package ui

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"turtlesilicon/pkg/addons"
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/launcher"
	"turtlesilicon/pkg/mods"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/utils"
	"turtlesilicon/pkg/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// createOptionsComponents initializes all option-related UI components
func createOptionsComponents() {
	// Get current version settings for initial values
	currentVer := GetCurrentVersion()
	if currentVer == nil {
		debug.Printf("Warning: No current version available, using default settings")
		currentVer = &version.GameVersion{Settings: version.VersionSettings{}}
	}

	metalHudCheckbox = widget.NewCheck("Enable Metal Hud (show FPS)", func(checked bool) {
		launcher.EnableMetalHud = checked
		// Save to current version settings
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.EnableMetalHud = checked
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Metal HUD enabled: %v", launcher.EnableMetalHud)
	})
	metalHudCheckbox.SetChecked(currentVer.Settings.EnableMetalHud)
	launcher.EnableMetalHud = currentVer.Settings.EnableMetalHud

	showTerminalCheckbox = widget.NewCheck("Show Terminal", func(checked bool) {
		// Save to current version settings
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.ShowTerminalNormally = checked
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Show terminal normally: %v", checked)
	})
	showTerminalCheckbox.SetChecked(currentVer.Settings.ShowTerminalNormally)

	autoDeleteWdbCheckbox = widget.NewCheck("Auto-delete WDB directory on launch", func(checked bool) {
		launcher.AutoDeleteWdb = checked
		// Save to current version settings
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.AutoDeleteWdb = checked
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Auto-delete WDB enabled: %v", launcher.AutoDeleteWdb)
	})
	autoDeleteWdbCheckbox.SetChecked(currentVer.Settings.AutoDeleteWdb)
	launcher.AutoDeleteWdb = currentVer.Settings.AutoDeleteWdb

	// Create recommended settings button with help icon
	applyRecommendedSettingsButton = widget.NewButton("Apply recommended settings", func() {
		err := launcher.ApplyRecommendedSettings()
		if err != nil {
			debug.Printf("Failed to apply recommended settings: %v", err)
			// Show error dialog if we have a window reference
			if currentWindow != nil {
				dialog.ShowError(fmt.Errorf("failed to apply recommended settings: %v", err), currentWindow)
			}
		} else {
			debug.Printf("Successfully applied recommended settings")
			// Show success dialog if we have a window reference
			if currentWindow != nil {
				dialog.ShowInformation("Success", "Recommended graphics settings have been applied", currentWindow)
			}
			// Update button state
			updateRecommendedSettingsButton()
		}
	})
	applyRecommendedSettingsButton.Importance = widget.MediumImportance

	// Create help button for recommended settings
	recommendedSettingsHelpButton = widget.NewButton("?", func() {
		showRecommendedSettingsHelpPopup()
	})
	recommendedSettingsHelpButton.Importance = widget.MediumImportance
	// Initialize button state
	updateRecommendedSettingsButton()

	// Create Wine registry Option-as-Alt buttons and status
	createWineRegistryComponents()

	// Create graphics settings components
	createGraphicsSettingsComponents()

	// Load environment variables from current version settings
	launcher.CustomEnvVars = currentVer.Settings.EnvironmentVariables

	envVarsEntry = widget.NewEntry()
	envVarsEntry.SetPlaceHolder(`Custom environment variables (KEY=VALUE format)`)
	envVarsEntry.SetText(launcher.CustomEnvVars)
	envVarsEntry.OnChanged = func(text string) {
		launcher.CustomEnvVars = text
		// Save to current version settings
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.EnvironmentVariables = text
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Environment variables updated: %v", launcher.CustomEnvVars)
	}
}

// createPatchingButtons creates all patching-related buttons
func createPatchingButtons(myWindow fyne.Window) {
	patchTurtleWoWButton = widget.NewButton("Patch Game", func() {
		PatchCurrentVersion(myWindow)
	})
	unpatchTurtleWoWButton = widget.NewButton("Unpatch Game", func() {
		UnpatchCurrentVersion(myWindow)
	})
	patchCrossOverButton = widget.NewButton("Patch CrossOver", func() {
		// Ensure CrossOver path is synced for CrossOver patching
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			paths.CrossoverPath = currentVer.CrossOverPath
		}
		patching.PatchCrossOver(myWindow, UpdateAllStatuses)
	})
	unpatchCrossOverButton = widget.NewButton("Unpatch CrossOver", func() {
		// Ensure CrossOver path is synced for CrossOver unpatching
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			paths.CrossoverPath = currentVer.CrossOverPath
		}
		patching.UnpatchCrossOver(myWindow, UpdateAllStatuses)
	})
}

// createServiceButtons creates service-related buttons
func createServiceButtons(myWindow fyne.Window) {
	startServiceButton = widget.NewButton("Start Service", func() {
		// Ensure legacy paths are synced before starting service
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			paths.TurtlewowPath = currentVer.GamePath
			paths.CrossoverPath = currentVer.CrossOverPath
		}
		service.StartRosettaX87Service(myWindow, UpdateAllStatuses)
	})
	stopServiceButton = widget.NewButton("Stop Service", func() {
		service.StopRosettaX87Service(myWindow, UpdateAllStatuses)
	})
}

// createLaunchButton creates the version-aware launch button
func createLaunchButton(myWindow fyne.Window) {
	launchButton = widget.NewButton("Launch Game", func() {
		LaunchCurrentVersion(myWindow)
	})
}

// createBottomBar creates the bottom bar with Options, GitHub, and PLAY buttons
func createBottomBar(myWindow fyne.Window) fyne.CanvasObject {
	// Set the current window for popup functionality
	currentWindow = myWindow

	// Options button
	optionsButton := widget.NewButton("Options", func() {
		showOptionsPopup()
	})

	// Troubleshooting button
	troubleshootingButton = widget.NewButton("Troubleshooting", func() {
		showTroubleshootingPopup()
	})

	// GitHub button
	githubButton := widget.NewButton("GitHub", func() {
		githubURL := "https://github.com/tairasu/TurtleSilicon"
		parsedURL, err := url.Parse(githubURL)
		if err != nil {
			debug.Printf("Error parsing GitHub URL: %v", err)
			return
		}
		fyne.CurrentApp().OpenURL(parsedURL)
	})
	playButtonText = widget.NewRichTextFromMarkdown("# PLAY")
	playButtonText.Wrapping = fyne.TextWrapOff

	playButton = widget.NewButton("", func() {
		LaunchCurrentVersion(myWindow)
	})
	playButton.Importance = widget.HighImportance
	playButton.Disable()

	playButtonWithText := container.NewStack(
		playButton,
		container.NewCenter(playButtonText),
	)

	// Addons button
	addonsButton := widget.NewButton("Addons", func() {
		addonManager := addons.NewAddonManager(myWindow)
		addonManager.ShowAddonManager()
	})

	// Initialize leftButtons container
	refreshLeftButtonsContainer(optionsButton, troubleshootingButton, addonsButton, githubButton, myWindow)

	// Create the large play button with fixed size
	buttonWidth := float32(120)
	buttonHeight := float32(80)
	playButtonWithText.Resize(fyne.NewSize(buttonWidth, buttonHeight))

	// Create a container for the play button that ensures it's positioned at bottom-right
	playButtonContainer := container.NewWithoutLayout(playButtonWithText)
	playButtonContainer.Resize(fyne.NewSize(buttonWidth+40, buttonHeight+20)) // Add padding

	playButtonWithText.Move(fyne.NewPos(-50, -32))

	// Use border layout to position elements
	bottomContainer := container.NewBorder(
		nil,                 // top
		nil,                 // bottom
		leftButtons,         // left
		playButtonContainer, // right - our large play button
		nil,                 // center
	)

	return container.NewPadded(bottomContainer)
}

// createWineRegistryComponents creates Wine registry Option-as-Alt buttons and status
func createWineRegistryComponents() {
	// Create status label to show current state
	optionAsAltStatusLabel = widget.NewRichText()

	// Create enable button
	enableOptionAsAltButton = widget.NewButton("Enable", func() {
		enableOptionAsAltButton.Disable()
		disableOptionAsAltButton.Disable()
		remapOperationInProgress = true

		// Show loading state in status label
		fyne.Do(func() {
			optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enabling...")
			startPulsingEffect()
		})

		// Run in goroutine to avoid blocking UI
		go func() {
			defer func() {
				remapOperationInProgress = false
			}()

			if err := utils.SetOptionAsAltEnabled(true); err != nil {
				debug.Printf("Failed to enable Option-as-Alt mapping: %v", err)
				// Update UI on main thread
				fyne.Do(func() {
					stopPulsingEffect()
					optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enable Failed")
				})
				time.Sleep(2 * time.Second) // Show error briefly
			} else {
				debug.Printf("Successfully enabled Option-as-Alt mapping")
				// Update preferences
				prefs, _ := utils.LoadPrefs()
				prefs.RemapOptionAsAlt = true
				utils.SavePrefs(prefs)
			}

			// Update UI on main thread
			fyne.Do(func() {
				stopPulsingEffect()
				updateWineRegistryStatusWithMethod(true) // Use Wine command for accurate check after modifications
			})
		}()
	})

	// Create disable button
	disableOptionAsAltButton = widget.NewButton("Disable", func() {
		enableOptionAsAltButton.Disable()
		disableOptionAsAltButton.Disable()
		remapOperationInProgress = true

		// Show loading state in status label
		fyne.Do(func() {
			optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disabling...")
			startPulsingEffect()
		})

		// Run in goroutine to avoid blocking UI
		go func() {
			defer func() {
				remapOperationInProgress = false
			}()

			if err := utils.SetOptionAsAltEnabled(false); err != nil {
				debug.Printf("Failed to disable Option-as-Alt mapping: %v", err)
				// Update UI on main thread
				fyne.Do(func() {
					stopPulsingEffect()
					optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disable Failed")
				})
				time.Sleep(2 * time.Second) // Show error briefly
			} else {
				debug.Printf("Successfully disabled Option-as-Alt mapping")
				// Update preferences
				prefs, _ := utils.LoadPrefs()
				prefs.RemapOptionAsAlt = false
				utils.SavePrefs(prefs)
			}

			// Update UI on main thread
			fyne.Do(func() {
				stopPulsingEffect()
				updateWineRegistryStatusWithMethod(true) // Use Wine command for accurate check after modifications
			})
		}()
	})

	// Style the buttons similar to other action buttons
	enableOptionAsAltButton.Importance = widget.MediumImportance
	disableOptionAsAltButton.Importance = widget.MediumImportance

	// Initialize status and button states
	updateWineRegistryStatus()
}

// updateWineRegistryStatus updates the Wine registry status label and button states
func updateWineRegistryStatus() {
	updateWineRegistryStatusWithMethod(false)
}

// updateWineRegistryStatusWithMethod updates status with choice of checking method
func updateWineRegistryStatusWithMethod(useWineCommand bool) {
	if useWineCommand {
		// Use Wine command for accurate check after modifications
		currentWineRegistryEnabled = utils.CheckOptionAsAltEnabled()
	} else {
		// Use fast file-based check for regular status updates
		currentWineRegistryEnabled = utils.CheckOptionAsAltEnabledFast()
	}

	// Update UI with simple white text
	if currentWineRegistryEnabled {
		optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enabled")
	} else {
		optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disabled")
	}

	// Update button states based on current status
	if currentWineRegistryEnabled {
		// If enabled, only show disable button as clickable
		enableOptionAsAltButton.Disable()
		disableOptionAsAltButton.Enable()
	} else {
		// If disabled, only show enable button as clickable
		enableOptionAsAltButton.Enable()
		disableOptionAsAltButton.Disable()
	}
}

// startPulsingEffect starts a pulsing animation for the status label during loading
func startPulsingEffect() {
	if pulsingActive {
		return // Already pulsing
	}

	pulsingActive = true
	pulsingTicker = time.NewTicker(500 * time.Millisecond)

	go func() {
		dots := ""

		for pulsingActive {
			<-pulsingTicker.C
			if pulsingActive {
				// Cycle through different dot patterns for visual effect
				switch len(dots) {
				case 0:
					dots = "."
				case 1:
					dots = ".."
				case 2:
					dots = "..."
				default:
					dots = ""
				}

				// Update the label with pulsing dots
				fyne.Do(func() {
					if pulsingActive && optionAsAltStatusLabel != nil {
						// Use the dots directly in the status text
						if strings.Contains(optionAsAltStatusLabel.String(), "Enabling") {
							optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enabling" + dots)
						} else if strings.Contains(optionAsAltStatusLabel.String(), "Disabling") {
							optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disabling" + dots)
						}
					}
				})
			}
		}
	}()
}

// stopPulsingEffect stops the pulsing animation
func stopPulsingEffect() {
	if !pulsingActive {
		return
	}

	pulsingActive = false
	if pulsingTicker != nil {
		pulsingTicker.Stop()
		pulsingTicker = nil
	}
}

// updateRecommendedSettingsButton updates the state of the recommended settings button
func updateRecommendedSettingsButton() {
	if applyRecommendedSettingsButton == nil {
		return
	}

	// Check if all recommended settings are already applied
	if launcher.CheckRecommendedSettings() {
		applyRecommendedSettingsButton.Disable()
		applyRecommendedSettingsButton.SetText("Settings applied")
	} else {
		applyRecommendedSettingsButton.Enable()
		applyRecommendedSettingsButton.SetText("Apply recommended settings")
	}

	// Help button should always be enabled
	if recommendedSettingsHelpButton != nil {
		recommendedSettingsHelpButton.Enable()
	}
}

// showRecommendedSettingsHelpPopup shows a popup explaining the recommended graphics settings
func showRecommendedSettingsHelpPopup() {
	if currentWindow == nil {
		return
	}

	// Create help content
	helpTitle := widget.NewRichTextFromMarkdown("# 📋 Recommended Graphics Settings")

	// Create individual setting labels for better formatting
	settingsTitle := widget.NewLabel("The following settings will be applied to your Config.wtf file:")
	settingsTitle.TextStyle = fyne.TextStyle{Bold: true}

	setting1 := widget.NewLabel("• Terrain Distance (farclip): 177 - Reduces CPU overhead - more fps")
	setting2 := widget.NewLabel("• Vertex Animation Shaders (M2UseShaders): Enabled - Prevents graphic glitches")
	setting3 := widget.NewLabel("• Multisampling (gxMultisample): 2x - Makes portraits load properly")

	settingsContainer := container.NewVBox(
		settingsTitle,
		widget.NewSeparator(),
		setting1,
		setting2,
		setting3,
		widget.NewSeparator(),
	)

	// Create OK button
	okButton := widget.NewButton("OK", func() {
		// This will be set when the popup is created
	})
	okButton.Importance = widget.MediumImportance

	// Create help content container
	helpContentContainer := container.NewVBox(
		container.NewCenter(helpTitle),
		widget.NewSeparator(),
		settingsContainer,
		widget.NewSeparator(),
		container.NewCenter(okButton),
	)

	// Calculate popup size
	windowSize := currentWindow.Content().Size()
	popupWidth := windowSize.Width * 3 / 4
	popupHeight := windowSize.Height * 3 / 4

	// Create the help popup
	helpPopup := widget.NewModalPopUp(container.NewPadded(helpContentContainer), currentWindow.Canvas())
	helpPopup.Resize(fyne.NewSize(popupWidth, popupHeight))

	// Set the OK button action to hide the help popup
	okButton.OnTapped = func() {
		helpPopup.Hide()
	}

	helpPopup.Show()
}

// createGraphicsSettingsComponents creates all graphics settings checkboxes and buttons
func createGraphicsSettingsComponents() {
	// Get current version settings for initial values
	currentVer := GetCurrentVersion()
	if currentVer == nil {
		debug.Printf("Warning: No current version available for graphics settings")
		currentVer = &version.GameVersion{Settings: version.VersionSettings{}}
	}

	// Create Reduce Terrain Distance setting with help button
	reduceTerrainDistanceCheckbox = widget.NewCheck("", func(checked bool) {
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.ReduceTerrainDistance = checked
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Reduce terrain distance: %v", checked)
		updateApplyGraphicsSettingsButton()
	})
	reduceTerrainDistanceCheckbox.SetChecked(currentVer.Settings.ReduceTerrainDistance)

	reduceTerrainDistanceHelpButton = widget.NewButton("?", func() {
		showGraphicsSettingHelpPopup("Reduce Terrain Distance", "Sets the draw distance to the lowest setting. This will drastically increase your FPS", "High Performance Impact")
	})
	reduceTerrainDistanceHelpButton.Importance = widget.MediumImportance

	// Create Set Multisample to 2x setting with help button
	setMultisampleTo2xCheckbox = widget.NewCheck("", func(checked bool) {
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.SetMultisampleTo2x = checked
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Set multisample to 2x: %v", checked)
		updateApplyGraphicsSettingsButton()
	})
	setMultisampleTo2xCheckbox.SetChecked(currentVer.Settings.SetMultisampleTo2x)

	setMultisampleTo2xHelpButton = widget.NewButton("?", func() {
		showGraphicsSettingHelpPopup("Set Multisample to 2x", "Might reduce your FPS slightly on lower end machines, but makes sure the portraits load properly.", "Medium Performance Impact")
	})
	setMultisampleTo2xHelpButton.Importance = widget.MediumImportance

	// Create Set Shadow LOD to 0 setting with help button
	setShadowLOD0Checkbox = widget.NewCheck("", func(checked bool) {
		currentVer := GetCurrentVersion()
		if currentVer != nil {
			currentVer.Settings.SetShadowLOD0 = checked
			// Track if user manually disabled this setting
			if !checked {
				currentVer.Settings.UserDisabledShadowLOD = true
			} else {
				currentVer.Settings.UserDisabledShadowLOD = false
			}
			SaveCurrentVersion(currentVer)
		}
		debug.Printf("Set shadow LOD to 0: %v (user manually changed)", checked)
		updateApplyGraphicsSettingsButton()
	})
	setShadowLOD0Checkbox.SetChecked(currentVer.Settings.SetShadowLOD0)

	setShadowLOD0HelpButton = widget.NewButton("?", func() {
		showGraphicsSettingHelpPopup("Set Shadow LOD to 0", "Turns off all shadows. This will give you ~10% more FPS.", "High Performance Impact")
	})
	setShadowLOD0HelpButton.Importance = widget.MediumImportance

	applyGraphicsSettingsButton = widget.NewButton("Apply Graphics Settings", func() {
		// Use version-specific settings
		currentVer := GetCurrentVersion()
		if currentVer == nil {
			dialog.ShowError(fmt.Errorf("no current version selected"), currentWindow)
			return
		}

		if currentVer.GamePath == "" {
			dialog.ShowError(fmt.Errorf("game path not set for current version"), currentWindow)
			return
		}

		err := patching.ApplyGraphicsSettingsForVersion(
			currentWindow,
			currentVer.GamePath,
			currentVer.Settings.ReduceTerrainDistance,
			currentVer.Settings.SetMultisampleTo2x,
			currentVer.Settings.SetShadowLOD0,
			currentVer.Settings.EnableLibSiliconPatch,
		)
		if err != nil {
			debug.Printf("Failed to apply graphics settings: %v", err)
			if currentWindow != nil {
				dialog.ShowError(fmt.Errorf("failed to apply graphics settings: %v", err), currentWindow)
			}
		} else {
			debug.Printf("Successfully applied graphics settings")
			if currentWindow != nil {
				dialog.ShowInformation("Success", "Graphics settings have been applied", currentWindow)
			}
			// Refresh checkboxes to reflect current state
			refreshGraphicsSettingsCheckboxes()
		}
	})
	applyGraphicsSettingsButton.Importance = widget.MediumImportance

	// Initialize button state
	updateApplyGraphicsSettingsButton()
}

// updateApplyGraphicsSettingsButton updates the state of the apply graphics settings button
func updateApplyGraphicsSettingsButton() {
	if applyGraphicsSettingsButton == nil {
		return
	}

	// Always enable the button since we need to handle both adding and removing settings
	applyGraphicsSettingsButton.Enable()
	applyGraphicsSettingsButton.SetText("Apply Changes")
}

// refreshGraphicsSettingsCheckboxes updates the checkbox states from current version settings
func refreshGraphicsSettingsCheckboxes() {
	currentVer := GetCurrentVersion()
	if currentVer == nil {
		debug.Printf("Warning: No current version available for refreshing graphics checkboxes")
		return
	}

	if reduceTerrainDistanceCheckbox != nil {
		reduceTerrainDistanceCheckbox.SetChecked(currentVer.Settings.ReduceTerrainDistance)
	}
	if setMultisampleTo2xCheckbox != nil {
		setMultisampleTo2xCheckbox.SetChecked(currentVer.Settings.SetMultisampleTo2x)
	}
	if setShadowLOD0Checkbox != nil {
		setShadowLOD0Checkbox.SetChecked(currentVer.Settings.SetShadowLOD0)
	}

	// Update the apply button state
	updateApplyGraphicsSettingsButton()
}

// showGraphicsSettingHelpPopup shows a help popup for a specific graphics setting
func showGraphicsSettingHelpPopup(title, description, impact string) {
	if currentWindow == nil {
		return
	}

	// Create help content
	helpTitle := widget.NewRichTextFromMarkdown("# " + title)

	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	impactLabel := widget.NewLabel(impact)
	impactLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create OK button
	okButton := widget.NewButton("OK", func() {
		// This will be set when the popup is created
	})
	okButton.Importance = widget.MediumImportance

	// Create help content container
	helpContentContainer := container.NewVBox(
		container.NewCenter(helpTitle),
		widget.NewSeparator(),
		descriptionLabel,
		widget.NewSeparator(),
		impactLabel,
		widget.NewSeparator(),
		container.NewCenter(okButton),
	)

	// Calculate popup size
	windowSize := currentWindow.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 2

	// Create the help popup
	helpPopup := widget.NewModalPopUp(container.NewPadded(helpContentContainer), currentWindow.Canvas())
	helpPopup.Resize(fyne.NewSize(popupWidth, popupHeight))

	// Set the OK button action to hide the help popup
	okButton.OnTapped = func() {
		helpPopup.Hide()
	}

	helpPopup.Show()
}

// refreshLeftButtonsContainer creates or updates the leftButtons container based on current version
func refreshLeftButtonsContainer(optionsButton, troubleshootingButton, addonsButton, githubButton *widget.Button, myWindow fyne.Window) {
	vm := GetCurrentVersionManager()
	if vm != nil {
		currentVer, err := vm.GetCurrentVersion()
		if err == nil && currentVer.SupportsDLLLoading {
			modsButton := widget.NewButton("Mods", func() {
				modManager := mods.NewModManager(myWindow, vm)
				modManager.ShowModManager()
			})
			leftButtons = container.NewHBox(
				optionsButton,
				troubleshootingButton,
				addonsButton,
				modsButton,
				githubButton,
			)
		} else {
			leftButtons = container.NewHBox(
				optionsButton,
				troubleshootingButton,
				addonsButton,
				githubButton,
			)
		}
	} else {
		leftButtons = container.NewHBox(
			optionsButton,
			troubleshootingButton,
			addonsButton,
			githubButton,
		)
	}
}

// RefreshLeftButtons updates the left buttons container for version changes
func RefreshLeftButtons() {
	if leftButtons == nil {
		return
	}

	vm := GetCurrentVersionManager()
	if vm != nil {
		currentVer, err := vm.GetCurrentVersion()
		if err == nil && currentVer.SupportsDLLLoading {
			modsButton := widget.NewButton("Mods", func() {
				modManager := mods.NewModManager(currentWindow, vm)
				modManager.ShowModManager()
			})
			leftButtons.Objects = []fyne.CanvasObject{
				leftButtons.Objects[0], // optionsButton
				leftButtons.Objects[1], // troubleshootingButton
				leftButtons.Objects[2], // addonsButton
				modsButton,
				leftButtons.Objects[len(leftButtons.Objects)-1], // githubButton (last)
			}
		} else {
			// Remove mods button if it exists
			leftButtons.Objects = []fyne.CanvasObject{
				leftButtons.Objects[0],                          // optionsButton
				leftButtons.Objects[1],                          // troubleshootingButton
				leftButtons.Objects[2],                          // addonsButton
				leftButtons.Objects[len(leftButtons.Objects)-1], // githubButton (last)
			}
		}
		leftButtons.Refresh()
	}
}
