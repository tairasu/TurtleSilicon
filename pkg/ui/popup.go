package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	"turtlesilicon/pkg/version"
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

	graphicsContainer := container.NewVBox(
		graphicsTitle,
		widget.NewSeparator(),
		graphicsDescription,
		widget.NewSeparator(),
		terrainRow,
		multisampleRow,
		shadowRow,
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

// showDebugLogPopup creates and shows a popup window with debug log content
func showDebugLogPopup() {
	if currentWindow == nil {
		return
	}

	// Create debug info structure
	debugInfo := &debug.DebugInfo{
		CrossoverPath:            paths.CrossoverPath,
		TurtlewowPath:            paths.TurtlewowPath,
		PatchesAppliedTurtleWoW:  paths.PatchesAppliedTurtleWoW,
		PatchesAppliedCrossOver:  paths.PatchesAppliedCrossOver,
		RosettaX87ServiceRunning: paths.RosettaX87ServiceRunning,
		ServiceStarting:          paths.ServiceStarting,
	}

	// Get current version info
	var gameVersionInfo *debug.GameVersionInfo = nil
	if vm, err := version.LoadVersionManager(); err == nil {
		if currentVer, err := vm.GetCurrentVersion(); err == nil {
			gameVersionInfo = &debug.GameVersionInfo{
				ID:                   currentVer.ID,
				DisplayName:          currentVer.DisplayName,
				WoWVersion:           currentVer.WoWVersion,
				GamePath:             currentVer.GamePath,
				ExecutableName:       currentVer.ExecutableName,
				SupportsDLLLoading:   currentVer.SupportsDLLLoading,
				UsesRosettaPatching:  currentVer.UsesRosettaPatching,
				UsesDivxDecoderPatch: currentVer.UsesDivxDecoderPatch,
				Settings: debug.GameVersionSettings{
					RemapOptionAsAlt:      currentVer.Settings.RemapOptionAsAlt,
					AutoDeleteWdb:         currentVer.Settings.AutoDeleteWdb,
					EnableMetalHud:        currentVer.Settings.EnableMetalHud,
					SaveSudoPassword:      currentVer.Settings.SaveSudoPassword,
					ShowTerminalNormally:  currentVer.Settings.ShowTerminalNormally,
					EnvironmentVariables:  currentVer.Settings.EnvironmentVariables,
					ReduceTerrainDistance: currentVer.Settings.ReduceTerrainDistance,
					SetMultisampleTo2x:    currentVer.Settings.SetMultisampleTo2x,
					SetShadowLOD0:         currentVer.Settings.SetShadowLOD0,
					EnableLibSiliconPatch: currentVer.Settings.EnableLibSiliconPatch,
				},
			}
		}
	}

	// Generate debug log content
	debugContent := debug.GenerateDebugLog(debugInfo, gameVersionInfo)

	// Create text entry with debug log content
	debugTextEntry := widget.NewMultiLineEntry()
	debugTextEntry.SetText(debugContent)
	debugTextEntry.Wrapping = fyne.TextWrapOff

	// Create scrollable container for the text
	scrollContainer := container.NewScroll(debugTextEntry)

	// Create copy button
	copyButton := widget.NewButton("Copy to Clipboard", func() {
		currentWindow.Clipboard().SetContent(debugContent)
		dialog.ShowInformation("Copied", "Debug log copied to clipboard!", currentWindow)
	})
	copyButton.Importance = widget.HighImportance

	// Create close button
	closeButton := widget.NewButton("Close", func() {})

	// Create buttons container
	buttonsContainer := container.NewHBox(
		copyButton,
		widget.NewSeparator(),
		closeButton,
	)

	// Create title
	titleLabel := widget.NewLabel("Debug Log")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	instructionLabel := widget.NewLabel("Copy this debug information and send it to support:")
	instructionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Create main content
	content := container.NewBorder(
		container.NewVBox(titleLabel, instructionLabel, widget.NewSeparator()), // top
		buttonsContainer, // bottom
		nil,              // left
		nil,              // right
		scrollContainer,  // center
	)

	// Create popup
	popup := widget.NewModalPopUp(container.NewPadded(content), currentWindow.Canvas())

	// Set close button action
	closeButton.OnTapped = func() {
		popup.Hide()
	}

	// Set popup size (80% of window size)
	canvasSize := currentWindow.Canvas().Size()
	popupWidth := canvasSize.Width * 0.8
	popupHeight := canvasSize.Height * 0.8
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	popup.Show()
}

// showTroubleshootingPopup creates and shows a popup window for troubleshooting actions
func showTroubleshootingPopup() {
	if currentWindow == nil {
		return
	}

	// Get current version for version-specific troubleshooting options
	currentVer := GetCurrentVersion()
	if currentVer == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), currentWindow)
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
		crossoverStatusDetail = widget.NewLabelWithStyle("⚠️ Please update to a newer version of CrossOver!", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
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

		// Use current version's game path instead of hardcoded TurtleWoW path
		gamePath := ""
		if currentVer != nil && currentVer.GamePath != "" {
			gamePath = currentVer.GamePath
		} else {
			// Fall back to legacy path if no current version
			gamePath = paths.TurtlewowPath
		}

		gameWine := filepath.Join(gamePath, ".wine")
		msg := "Are you sure you want to delete the following Wine prefixes?\n\n- " + userWine + "\n- " + gameWine + "\n\nThis cannot be undone."
		dialog.NewConfirm("Delete Wine Prefixes", msg, func(confirm bool) {
			if confirm {
				err1 := os.RemoveAll(userWine)
				err2 := os.RemoveAll(gameWine)
				if err1 != nil && !os.IsNotExist(err1) {
					dialog.ShowError(fmt.Errorf("Failed to delete ~/.wine: %v", err1), currentWindow)
					return
				}
				if err2 != nil && !os.IsNotExist(err2) {
					dialog.ShowError(fmt.Errorf("Failed to delete game/.wine: %v", err2), currentWindow)
					return
				}
				dialog.ShowInformation("Wine Prefixes Deleted", "Wine prefixes deleted successfully.", currentWindow)
			}
		}, currentWindow).Show()
	})

	// --- Generate Debug Log ---
	debugLogButton := widget.NewButton("Show Debug Log", func() {
		showDebugLogPopup()
	})

	troubleshootingTitle := widget.NewLabel("Troubleshooting")
	troubleshootingTitle.TextStyle = fyne.TextStyle{Bold: true}

	// Base troubleshooting content
	content := container.NewVBox(
		troubleshootingTitle,
		widget.NewSeparator(),
	)

	// Version-specific content
	if currentVer.ID == "turtlesilicon" {
		// TurtleWoW-specific section
		turtleTitle := widget.NewLabel("TurtleWoW Specific")
		turtleTitle.TextStyle = fyne.TextStyle{Bold: true}

		downloadButton := widget.NewButton("Download Fresh Installation", func() {
			downloadURL := "http://eudl.turtle-wow.org/twmoa_1172.zip"
			// Open URL in browser
			cmd := exec.Command("open", downloadURL)
			if err := cmd.Start(); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to open download URL: %v", err), currentWindow)
			}
		})
		downloadButton.Importance = widget.HighImportance

		rowDownload := container.NewBorder(nil, nil, widget.NewLabel("Download fresh TurtleWoW client:"), downloadButton, nil)

		content.Add(turtleTitle)
		content.Add(widget.NewSeparator())
		content.Add(rowDownload)
		content.Add(widget.NewSeparator())
	} else if currentVer.ID == "epochsilicon" {
		// EpochSilicon-specific section
		epochTitle := widget.NewLabel("Project Epoch Specific")
		epochTitle.TextStyle = fyne.TextStyle{Bold: true}

		epochNotice := widget.NewLabel("⚠️ Important: This version only works if the Project Epoch client is updated using the official Ascension launcher. Unfortunately, Ascension does not share any other way to update Project Epoch files.")
		epochNotice.Wrapping = fyne.TextWrapWord
		epochNotice.TextStyle = fyne.TextStyle{Italic: true}

		content.Add(epochTitle)
		content.Add(widget.NewSeparator())
		content.Add(epochNotice)
		content.Add(widget.NewSeparator())
	}

	rowCrossover := container.NewBorder(nil, nil, widget.NewLabel("CrossOver version:"), crossoverStatusShort, nil)
	rowWDB := container.NewBorder(nil, nil, widget.NewLabel("Delete WDB directory (cache):"), wdbDeleteButton, nil)

	// Create version-aware wine prefix label
	wineLabel := "Delete Wine prefixes (~/.wine & Game/.wine):"
	if currentVer != nil && currentVer.GamePath != "" {
		gameName := filepath.Base(currentVer.GamePath)
		wineLabel = fmt.Sprintf("Delete Wine prefixes (~/.wine & %s/.wine):", gameName)
	}
	rowWine := container.NewBorder(nil, nil, widget.NewLabel(wineLabel), wineDeleteButton, nil)

	rowDebugLog := container.NewBorder(nil, nil, widget.NewLabel("Show debug log for support:"), debugLogButton, nil)
	appMgmtNote := widget.NewLabel("Please ensure TurtleSilicon is enabled in System Settings > Privacy & Security > App Management.")
	appMgmtNote.Wrapping = fyne.TextWrapWord
	appMgmtNote.TextStyle = fyne.TextStyle{Italic: true}

	content.Add(rowCrossover)
	content.Add(crossoverStatusDetail)
	content.Add(rowWDB)
	content.Add(rowWine)
	content.Add(rowDebugLog)
	content.Add(widget.NewSeparator())
	content.Add(appMgmtNote)

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
	if len(parts) < 2 {
		return false
	}

	// Convert version parts to integers for proper comparison
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	// Handle patch version (default to 0 if not present)
	patch := 0
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return false
		}
	}

	// Check if version >= 25.0.1
	if major > 25 {
		return true
	}
	if major == 25 {
		if minor > 0 {
			return true
		}
		if minor == 0 && patch >= 1 {
			return true
		}
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

// showNewUserTurtleWoWPopup shows a popup for new TurtleWoW users to download the client
func showNewUserTurtleWoWPopup() {
	if currentWindow == nil {
		return
	}

	// Create popup content
	titleText := widget.NewRichTextFromMarkdown("# Welcome to TurtleSilicon!")
	titleText.Wrapping = fyne.TextWrapOff

	messageText := widget.NewLabel("It looks like you're new to TurtleWoW! To get started, you'll need to download the TurtleWoW client.")
	messageText.Wrapping = fyne.TextWrapWord

	downloadInfoText := widget.NewLabel("Click the button below to download the official TurtleWoW client (about 9GB):")
	downloadInfoText.Wrapping = fyne.TextWrapWord
	downloadInfoText.TextStyle = fyne.TextStyle{Bold: true}

	// Download button
	downloadButton := widget.NewButton("Download TurtleWoW Client", func() {
		downloadURL := "http://eudl.turtle-wow.org/twmoa_1172.zip"
		cmd := exec.Command("open", downloadURL)
		if err := cmd.Start(); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to open download URL: %v", err), currentWindow)
		} else {
			// Show success message
			dialog.ShowInformation("Download Started", "The TurtleWoW download has been started in your browser.", currentWindow)
		}
	})
	downloadButton.Importance = widget.HighImportance

	// Instructions
	instructionsText := widget.NewLabel("After downloading:\n\n" +
		"1. Extract the ZIP file to a location on your Mac (e.g., Desktop or Applications)\n" +
		"2. Return to TurtleSilicon and click \"Set/Change\" next to Game Path\n" +
		"3. Select the extracted TurtleWoW folder\n" +
		"4. Follow the patching steps to optimize for Apple Silicon\n\n" +
		"You can dismiss this popup if you already have TurtleWoW installed.")
	instructionsText.Wrapping = fyne.TextWrapWord
	instructionsText.TextStyle = fyne.TextStyle{Italic: true}

	// Buttons
	dismissButton := widget.NewButton("Dismiss", func() {
		// This will be set when the popup is created
	})
	dismissButton.Importance = widget.LowImportance

	buttonsContainer := container.NewHBox(
		dismissButton,
		downloadButton,
	)

	// Create content container
	content := container.NewVBox(
		container.NewCenter(titleText),
		widget.NewSeparator(),
		messageText,
		widget.NewSeparator(),
		downloadInfoText,
		widget.NewSeparator(),
		instructionsText,
		widget.NewSeparator(),
		container.NewCenter(buttonsContainer),
	)

	// Calculate popup size
	windowSize := currentWindow.Canvas().Size()
	popupWidth := windowSize.Width * 3 / 4
	popupHeight := windowSize.Height * 3 / 4

	// Create popup
	popup := widget.NewModalPopUp(container.NewPadded(content), currentWindow.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	// Set button actions
	dismissButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}

// CheckAndShowNewUserPopup checks if the new user popup should be shown and shows it
func CheckAndShowNewUserPopup() {
	// Only show for TurtleWoW when no game path is set
	currentVer := GetCurrentVersion()
	if currentVer != nil && currentVer.ID == "turtlesilicon" && currentVer.GamePath == "" {
		debug.Printf("Showing new user popup for TurtleWoW")
		showNewUserTurtleWoWPopup()
	}
}
