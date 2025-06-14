package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// showOptionsPopup creates and shows an integrated popup window for options
func showOptionsPopup() {
	if currentWindow == nil {
		return
	}

	// Create options content with better organization and smaller titles
	optionsTitle := widget.NewLabel("Options")
	optionsTitle.TextStyle = fyne.TextStyle{Bold: true}
	// Create label for recommended settings
	recommendedSettingsLabel := widget.NewLabel("Graphics settings:")

	gameOptionsContainer := container.NewVBox(
		optionsTitle,
		widget.NewSeparator(),
		metalHudCheckbox,
		showTerminalCheckbox,
		vanillaTweaksCheckbox,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, recommendedSettingsLabel, container.NewHBox(applyRecommendedSettingsButton, recommendedSettingsHelpButton), nil),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, container.NewHBox(enableOptionAsAltButton, disableOptionAsAltButton), optionAsAltStatusLabel),
	)

	envVarsTitle := widget.NewLabel("Environment Variables")
	envVarsTitle.TextStyle = fyne.TextStyle{Bold: true}
	envVarsContainer := container.NewVBox(
		envVarsTitle,
		widget.NewSeparator(),
		envVarsEntry,
	)

	// Create a scrollable container for all options
	optionsContent := container.NewVBox(
		gameOptionsContainer,
		envVarsContainer,
	)

	scrollContainer := container.NewScroll(optionsContent)

	// Create close button
	closeButton := widget.NewButton("Close", func() {
		// This will be set when the popup is created
	})

	// Create the popup content with close button
	popupContent := container.NewBorder(
		nil,                                  // top
		container.NewCenter(closeButton),     // bottom
		nil,                                  // left
		nil,                                  // right
		container.NewPadded(scrollContainer), // center
	)

	// Get the window size and calculate 2/3 size
	windowSize := currentWindow.Content().Size()
	popupWidth := windowSize.Width * 5 / 6
	popupHeight := windowSize.Height * 5 / 6

	// Create a modal popup
	popup := widget.NewModalPopUp(popupContent, currentWindow.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	// Set the close button action to hide the popup
	closeButton.OnTapped = func() {
		if remapOperationInProgress {
			// Show warning popup instead of closing
			showRemapWarningPopup()
		} else {
			popup.Hide()
		}
	}

	popup.Show()
}

// showRemapWarningPopup shows a warning popup when user tries to close options during remap operation
func showRemapWarningPopup() {
	if currentWindow == nil {
		return
	}

	// Create warning content
	warningTitle := widget.NewRichTextFromMarkdown("# ⚠️ Please Wait")
	warningMessage := widget.NewRichTextFromMarkdown("**Remap operation is in progress.**\n\nThe wine registry is being modified. This will take a moment.\n\nPlease wait for the operation to complete before closing the options.")

	// Create OK button
	okButton := widget.NewButton("OK", func() {
		// This will be set when the popup is created
	})
	okButton.Importance = widget.HighImportance

	// Create warning content container
	warningContent := container.NewVBox(
		container.NewCenter(warningTitle),
		widget.NewSeparator(),
		warningMessage,
		widget.NewSeparator(),
		container.NewCenter(okButton),
	)

	// Calculate smaller popup size
	windowSize := currentWindow.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 2

	// Create the warning popup
	warningPopup := widget.NewModalPopUp(container.NewPadded(warningContent), currentWindow.Canvas())
	warningPopup.Resize(fyne.NewSize(popupWidth, popupHeight))

	// Set the OK button action to hide the warning popup
	okButton.OnTapped = func() {
		warningPopup.Hide()
	}

	warningPopup.Show()
}
