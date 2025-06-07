package ui

import (
	"log"
	"net/url"

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
