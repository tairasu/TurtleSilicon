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

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"
)

// showOptionsPopup creates and shows an integrated popup window for options
func showOptionsPopup() {
	if currentWindow == nil {
		return
	}

	// Check graphics settings presence and update preferences before showing UI
	patching.CheckGraphicsSettingsPresence()

	// Load graphics settings from Config.wtf and update preferences
	if err := patching.LoadGraphicsSettingsFromConfig(); err != nil {
		debug.Printf("Warning: failed to load graphics settings from Config.wtf: %v", err)
	}

	// Refresh checkbox states to reflect current settings
	refreshGraphicsSettingsCheckboxes()

	// Create General tab content
	generalTitle := widget.NewLabel("General Settings")
	generalTitle.TextStyle = fyne.TextStyle{Bold: true}

	generalContainer := container.NewVBox(
		generalTitle,
		widget.NewSeparator(),
		metalHudCheckbox,
		showTerminalCheckbox,
		vanillaTweaksCheckbox,
		autoDeleteWdbCheckbox,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, container.NewHBox(enableOptionAsAltButton, disableOptionAsAltButton), optionAsAltStatusLabel),
	)

	// Create Graphics tab content
	graphicsTitle := widget.NewLabel("Graphics Settings")
	graphicsTitle.TextStyle = fyne.TextStyle{Bold: true}

	graphicsDescription := widget.NewLabel("Select graphics settings to apply to Config.wtf:")
	graphicsDescription.TextStyle = fyne.TextStyle{Italic: true}

	// Create bold text labels for each setting
	terrainLabel := widget.NewLabel("Reduce Terrain Distance")
	terrainLabel.TextStyle = fyne.TextStyle{Bold: true}

	multisampleLabel := widget.NewLabel("Set Multisample to 2x")
	multisampleLabel.TextStyle = fyne.TextStyle{Bold: true}

	shadowLabel := widget.NewLabel("Set Shadow LOD to 0")
	shadowLabel.TextStyle = fyne.TextStyle{Bold: true}

	libSiliconPatchLabel := widget.NewLabel("Enable libSiliconPatch")
	libSiliconPatchLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create setting rows with help buttons between checkbox and label
	terrainRow := container.NewHBox(
		reduceTerrainDistanceCheckbox,
		reduceTerrainDistanceHelpButton,
		terrainLabel)
	multisampleRow := container.NewHBox(
		setMultisampleTo2xCheckbox,
		setMultisampleTo2xHelpButton,
		multisampleLabel)
	shadowRow := container.NewHBox(
		setShadowLOD0Checkbox,
		setShadowLOD0HelpButton,
		shadowLabel)
	libSiliconPatchRow := container.NewHBox(
		libSiliconPatchCheckbox,
		libSiliconPatchHelpButton,
		libSiliconPatchLabel)

	graphicsContainer := container.NewVBox(
		graphicsTitle,
		widget.NewSeparator(),
		graphicsDescription,
		widget.NewSeparator(),
		terrainRow,
		multisampleRow,
		shadowRow,
		libSiliconPatchRow,
		widget.NewSeparator(),
		container.NewCenter(applyGraphicsSettingsButton),
	)

	// Create Environment Variables tab content
	envVarsTitle := widget.NewLabel("Environment Variables")
	envVarsTitle.TextStyle = fyne.TextStyle{Bold: true}
	envVarsContainer := container.NewVBox(
		envVarsTitle,
		widget.NewSeparator(),
		envVarsEntry,
	)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("General", container.NewScroll(generalContainer)),
		container.NewTabItem("Graphics", container.NewScroll(graphicsContainer)),
		container.NewTabItem("Environment", container.NewScroll(envVarsContainer)),
	)

	// Set tab location to top
	tabs.SetTabLocation(container.TabLocationTop)

	// Create popup title
	optionsTitle := widget.NewLabel("Options")
	optionsTitle.TextStyle = fyne.TextStyle{Bold: true}

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge
	closeButton.Resize(fyne.NewSize(30, 30))

	// Create top bar with close button on left and title in center
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		container.NewCenter(optionsTitle),
	)

	// Create the popup content with close button
	popupContent := container.NewBorder(
		topBar,                    // top
		nil,                       // bottom
		nil,                       // left
		nil,                       // right
		container.NewPadded(tabs), // center
	)

	// Create a modal popup
	popup := widget.NewModalPopUp(popupContent, currentWindow.Canvas())

	canvasSize := currentWindow.Canvas().Size()
	popup.Resize(canvasSize)

	// Add keyboard shortcut for Escape key
	canvas := currentWindow.Canvas()
	originalOnTypedKey := canvas.OnTypedKey()

	// Set the close button action to hide the popup
	closeAction := func() {
		// Restore original key handler before closing
		canvas.SetOnTypedKey(originalOnTypedKey)
		if remapOperationInProgress {
			// Show warning popup instead of closing
			showRemapWarningPopup()
		} else {
			popup.Hide()
		}
	}

	closeButton.OnTapped = closeAction

	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			closeAction()
			return
		}
		if originalOnTypedKey != nil {
			originalOnTypedKey(key)
		}
	})

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
		deleteWDBDirectoriesInPopup()
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

	// Create popup title
	troubleshootingPopupTitle := widget.NewLabel("Troubleshooting")
	troubleshootingPopupTitle.TextStyle = fyne.TextStyle{Bold: true}

	// Create square close button without text padding
	troubleshootingCloseButton = widget.NewButton("✕", func() {})
	troubleshootingCloseButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	troubleshootingCloseButton.Resize(fyne.NewSize(24, 24))
	troubleshootingCloseButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge
	troubleshootingCloseButton.Resize(fyne.NewSize(30, 30))

	// Create top bar with close button on left and title in center
	topBar := container.NewBorder(
		nil,
		nil,
		troubleshootingCloseButton,
		nil,
		container.NewCenter(troubleshootingPopupTitle),
	)

	popupContent := container.NewBorder(
		topBar,                               // top
		nil,                                  // bottom
		nil,                                  // left
		nil,                                  // right
		container.NewPadded(scrollContainer), // center
	)

	popup := widget.NewModalPopUp(popupContent, currentWindow.Canvas())

	canvasSize := currentWindow.Canvas().Size()
	popup.Resize(canvasSize)

	// Add keyboard shortcut for Escape key
	canvas := currentWindow.Canvas()
	originalOnTypedKey := canvas.OnTypedKey()

	closeAction := func() {
		// Restore original key handler before closing
		canvas.SetOnTypedKey(originalOnTypedKey)
		popup.Hide()
	}

	troubleshootingCloseButton.OnTapped = closeAction

	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			closeAction()
			return
		}
		if originalOnTypedKey != nil {
			originalOnTypedKey(key)
		}
	})

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

// deleteWDBDirectoriesInPopup deletes WDB directories for troubleshooting popup
func deleteWDBDirectoriesInPopup() {
	currentVer := GetCurrentVersion()
	gamePath := ""

	if currentVer != nil {
		gamePath = currentVer.GamePath
	} else {
		// Fall back to legacy path
		gamePath = paths.TurtlewowPath
	}

	if gamePath == "" {
		dialog.ShowInformation("Path Not Set", "No game path is set. Please set your game directory first.", currentWindow)
		return
	}

	// Check for WDB directories
	wdbPath := filepath.Join(gamePath, "WDB")
	cacheWdbPath := filepath.Join(gamePath, "Cache", "WDB")

	wdbExists := utils.DirExists(wdbPath)
	cacheWdbExists := utils.DirExists(cacheWdbPath)

	if !wdbExists && !cacheWdbExists {
		dialog.ShowInformation("WDB Not Found", "No WDB directories found in your game folder.", currentWindow)
		return
	}

	// Build message based on what exists
	message := "Are you sure you want to delete the WDB directories? This will remove all cached data. No important data will be lost.\n\nDirectories to delete:\n"
	if wdbExists {
		message += "- " + wdbPath + "\n"
	}
	if cacheWdbExists {
		message += "- " + cacheWdbPath + "\n"
	}

	dialog.NewConfirm("Delete WDB Directories", message, func(confirm bool) {
		if confirm {
			var errors []string

			// Delete main WDB directory
			if wdbExists {
				if err := os.RemoveAll(wdbPath); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to delete %s: %v", wdbPath, err))
				}
			}

			// Delete Cache/WDB directory
			if cacheWdbExists {
				if err := os.RemoveAll(cacheWdbPath); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to delete %s: %v", cacheWdbPath, err))
				}
			}

			if len(errors) > 0 {
				dialog.ShowError(fmt.Errorf("Some WDB directories could not be deleted:\n%s", strings.Join(errors, "\n")), currentWindow)
			} else {
				dialog.ShowInformation("WDB Deleted", "WDB directories deleted successfully.", currentWindow)
			}
		}
	}, currentWindow).Show()
}
