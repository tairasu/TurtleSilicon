package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"howett.net/plist"

	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"
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
		autoDeleteWdbCheckbox,
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
	popupHeight := windowSize.Height * 9 / 10

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

// showTroubleshootingPopup creates and shows a popup window for troubleshooting actions
func showTroubleshootingPopup() {
	if currentWindow == nil {
		return
	}

	// --- CrossOver Version Check ---
	crossoverVersion := getCrossoverVersion(paths.CrossoverPath)
	var crossoverStatusShort *widget.Label
	var crossoverStatusDetail *widget.Label
	if crossoverVersion == "" {
		crossoverStatusShort = widget.NewLabel("Not found")
		crossoverStatusDetail = widget.NewLabel("")
	} else if isCrossoverVersionRecommended(crossoverVersion) {
		crossoverStatusShort = widget.NewLabelWithStyle("✔ "+crossoverVersion, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
		crossoverStatusDetail = widget.NewLabelWithStyle("✔ Recommended version of CrossOver installed", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	} else {
		crossoverStatusShort = widget.NewLabelWithStyle("⚠️ "+crossoverVersion, fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})
		crossoverStatusDetail = widget.NewLabelWithStyle("⚠️ Please update to CrossOver 25.0.1 or later!", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	}
	crossoverStatusDetail.Wrapping = fyne.TextWrapWord

	// --- Delete WDB Directory ---
	wdbDeleteButton = widget.NewButton("Delete", func() {
		wdbPath := filepath.Join(paths.TurtlewowPath, "WDB")
		if !utils.DirExists(wdbPath) {
			dialog.ShowInformation("WDB Not Found", "No WDB directory found in your TurtleWoW folder.", currentWindow)
			return
		}
		dialog.NewConfirm("Delete WDB Directory", "Are you sure you want to delete the WDB directory? This will remove all cached data. No important data will be lost.", func(confirm bool) {
			if confirm {
				err := os.RemoveAll(wdbPath)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Failed to delete WDB: %v", err), currentWindow)
				} else {
					dialog.ShowInformation("WDB Deleted", "WDB directory deleted successfully.", currentWindow)
				}
			}
		}, currentWindow).Show()
	})

	// --- Delete Wine Prefixes ---
	wineDeleteButton = widget.NewButton("Delete", func() {
		homeDir, _ := os.UserHomeDir()
		userWine := filepath.Join(homeDir, ".wine")
		turtleWine := filepath.Join(paths.TurtlewowPath, ".wine")
		msg := "Are you sure you want to delete the following Wine prefixes?\n\n- " + userWine + "\n- " + turtleWine + "\n\nThis cannot be undone."
		dialog.NewConfirm("Delete Wine Prefixes", msg, func(confirm bool) {
			if confirm {
				err1 := os.RemoveAll(userWine)
				err2 := os.RemoveAll(turtleWine)
				if err1 != nil && !os.IsNotExist(err1) {
					dialog.ShowError(fmt.Errorf("Failed to delete ~/.wine: %v", err1), currentWindow)
					return
				}
				if err2 != nil && !os.IsNotExist(err2) {
					dialog.ShowError(fmt.Errorf("Failed to delete TurtleWoW/.wine: %v", err2), currentWindow)
					return
				}
				dialog.ShowInformation("Wine Prefixes Deleted", "Wine prefixes deleted successfully.", currentWindow)
			}
		}, currentWindow).Show()
	})

	troubleshootingTitle := widget.NewLabel("Troubleshooting")
	troubleshootingTitle.TextStyle = fyne.TextStyle{Bold: true}

	rowCrossover := container.NewBorder(nil, nil, widget.NewLabel("CrossOver version:"), crossoverStatusShort, nil)
	rowWDB := container.NewBorder(nil, nil, widget.NewLabel("Delete WDB directory (cache):"), wdbDeleteButton, nil)
	rowWine := container.NewBorder(nil, nil, widget.NewLabel("Delete Wine prefixes (~/.wine & TurtleWoW/.wine):"), wineDeleteButton, nil)
	appMgmtNote := widget.NewLabel("Please ensure TurtleSilicon is enabled in System Settings > Privacy & Security > App Management.")
	appMgmtNote.Wrapping = fyne.TextWrapWord
	appMgmtNote.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		troubleshootingTitle,
		widget.NewSeparator(),
		rowCrossover,
		crossoverStatusDetail,
		rowWDB,
		rowWine,
		appMgmtNote,
	)

	scrollContainer := container.NewScroll(content)

	troubleshootingCloseButton = widget.NewButton("Close", func() {})

	popupContent := container.NewBorder(
		nil, // top
		container.NewCenter(troubleshootingCloseButton), // bottom
		nil,                                  // left
		nil,                                  // right
		container.NewPadded(scrollContainer), // center
	)

	windowSize := currentWindow.Content().Size()
	popupWidth := windowSize.Width * 5 / 6
	popupHeight := windowSize.Height * 5 / 6

	popup := widget.NewModalPopUp(popupContent, currentWindow.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	troubleshootingCloseButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}

// getCrossoverVersion reads the Info.plist and returns the version string, or "" if not found
func getCrossoverVersion(appPath string) string {
	if appPath == "" {
		return ""
	}
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	f, err := os.Open(plistPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	var data struct {
		Version string `plist:"CFBundleShortVersionString"`
	}
	decoder := plist.NewDecoder(f)
	if err := decoder.Decode(&data); err != nil {
		return ""
	}
	return data.Version
}

// isCrossoverVersionRecommended returns true if version >= 25.0.1
func isCrossoverVersionRecommended(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return false
	}
	major := parts[0]
	minor := parts[1]
	patch := parts[2]
	if major > "25" {
		return true
	}
	if major == "25" && minor >= "0" && patch >= "1" {
		return true
	}
	return false
}
