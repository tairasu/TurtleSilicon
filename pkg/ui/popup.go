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
	gameOptionsContainer := container.NewVBox(
		optionsTitle,
		widget.NewSeparator(),
		metalHudCheckbox,
		showTerminalCheckbox,
		vanillaTweaksCheckbox,
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
		popup.Hide()
	}

	popup.Show()
}
