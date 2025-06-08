package ui

import (
	"log"
	"net/url"
	"strings"
	"time"

	"turtlesilicon/pkg/launcher"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createOptionsComponents initializes all option-related UI components
func createOptionsComponents() {
	// Load preferences for initial values
	prefs, _ := utils.LoadPrefs()

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

	// Create Wine registry Option-as-Alt buttons and status
	createWineRegistryComponents()

	// Load environment variables from preferences
	if prefs.EnvironmentVariables != "" {
		launcher.CustomEnvVars = prefs.EnvironmentVariables
	}

	envVarsEntry = widget.NewEntry()
	envVarsEntry.SetPlaceHolder(`Custom environment variables (KEY=VALUE format)`)
	envVarsEntry.SetText(launcher.CustomEnvVars)
	envVarsEntry.OnChanged = func(text string) {
		launcher.CustomEnvVars = text
		// Save to preferences
		prefs, _ := utils.LoadPrefs()
		prefs.EnvironmentVariables = text
		utils.SavePrefs(prefs)
		log.Printf("Environment variables updated: %v", launcher.CustomEnvVars)
	}
}

// createPatchingButtons creates all patching-related buttons
func createPatchingButtons(myWindow fyne.Window) {
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
}

// createServiceButtons creates service-related buttons
func createServiceButtons(myWindow fyne.Window) {
	startServiceButton = widget.NewButton("Start Service", func() {
		service.StartRosettaX87Service(myWindow, UpdateAllStatuses)
	})
	stopServiceButton = widget.NewButton("Stop Service", func() {
		service.StopRosettaX87Service(myWindow, UpdateAllStatuses)
	})
}

// createLaunchButton creates the legacy launch button
func createLaunchButton(myWindow fyne.Window) {
	launchButton = widget.NewButton("Launch Game", func() {
		launcher.LaunchGame(myWindow)
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

	// GitHub button
	githubButton := widget.NewButton("GitHub", func() {
		githubURL := "https://github.com/tairasu/TurtleSilicon"
		parsedURL, err := url.Parse(githubURL)
		if err != nil {
			log.Printf("Error parsing GitHub URL: %v", err)
			return
		}
		fyne.CurrentApp().OpenURL(parsedURL)
	})
	playButtonText = widget.NewRichTextFromMarkdown("# PLAY")
	playButtonText.Wrapping = fyne.TextWrapOff

	playButton = widget.NewButton("", func() {
		launcher.LaunchGame(myWindow)
	})
	playButton.Importance = widget.HighImportance
	playButton.Disable()

	playButtonWithText := container.NewStack(
		playButton,
		container.NewCenter(playButtonText),
	)

	leftButtons := container.NewHBox(
		optionsButton,
		widget.NewSeparator(), // Visual separator
		githubButton,
	)

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

		// Show loading state in status label
		fyne.Do(func() {
			optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enabling...")
			startPulsingEffect()
		})

		// Run in goroutine to avoid blocking UI
		go func() {
			if err := utils.SetOptionAsAltEnabled(true); err != nil {
				log.Printf("Failed to enable Option-as-Alt mapping: %v", err)
				// Update UI on main thread
				fyne.Do(func() {
					stopPulsingEffect()
					optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Enable Failed")
				})
				time.Sleep(2 * time.Second) // Show error briefly
			} else {
				log.Printf("Successfully enabled Option-as-Alt mapping")
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

		// Show loading state in status label
		fyne.Do(func() {
			optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disabling...")
			startPulsingEffect()
		})

		// Run in goroutine to avoid blocking UI
		go func() {
			if err := utils.SetOptionAsAltEnabled(false); err != nil {
				log.Printf("Failed to disable Option-as-Alt mapping: %v", err)
				// Update UI on main thread
				fyne.Do(func() {
					stopPulsingEffect()
					optionAsAltStatusLabel.ParseMarkdown("**Remap Option key as Alt key:** Disable Failed")
				})
				time.Sleep(2 * time.Second) // Show error briefly
			} else {
				log.Printf("Successfully disabled Option-as-Alt mapping")
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
			select {
			case <-pulsingTicker.C:
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
