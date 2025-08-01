package ui

import (
	"fmt"
	"os"
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/epochsilicon"
	"turtlesilicon/pkg/launcher"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/service"
	"turtlesilicon/pkg/utils"
	"turtlesilicon/pkg/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	currentVersionManager *version.VersionManager
	currentVersion        *version.GameVersion
)

// InitializeVersionSystem initializes the version management system
func InitializeVersionSystem() error {
	vm, err := version.LoadVersionManager()
	if err != nil {
		return fmt.Errorf("failed to load version manager: %v", err)
	}

	currentVersionManager = vm

	// Get current version
	currentVer, err := vm.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %v", err)
	}
	currentVersion = currentVer

	// Migrate old preferences if needed
	if err := migrateOldPreferences(); err != nil {
		debug.Printf("Warning: failed to migrate old preferences: %v", err)
	}

	// Check and set default CrossOver path for all versions
	checkDefaultCrossOverPathForAllVersions()

	// Sync legacy paths for backward compatibility
	syncLegacyPaths()

	return nil
}

// migrateOldPreferences migrates old UserPrefs to the new version system
func migrateOldPreferences() error {
	oldPrefs, err := utils.LoadPrefs()
	if err != nil {
		return err
	}

	// If no old paths are set, no migration needed
	if oldPrefs.TurtleWoWPath == "" && oldPrefs.CrossOverPath == "" {
		return nil
	}

	// Migrate TurtleSilicon version settings
	turtleSiliconVersion, err := currentVersionManager.GetVersion("turtlesilicon")
	if err != nil {
		return err
	}

	// Update paths
	if oldPrefs.TurtleWoWPath != "" {
		turtleSiliconVersion.GamePath = oldPrefs.TurtleWoWPath
	}
	if oldPrefs.CrossOverPath != "" {
		turtleSiliconVersion.CrossOverPath = oldPrefs.CrossOverPath
	}

	// Migrate settings
	turtleSiliconVersion.Settings.EnableVanillaTweaks = oldPrefs.EnableVanillaTweaks
	turtleSiliconVersion.Settings.RemapOptionAsAlt = oldPrefs.RemapOptionAsAlt
	turtleSiliconVersion.Settings.AutoDeleteWdb = oldPrefs.AutoDeleteWdb
	turtleSiliconVersion.Settings.EnableMetalHud = oldPrefs.EnableMetalHud
	turtleSiliconVersion.Settings.SaveSudoPassword = oldPrefs.SaveSudoPassword
	turtleSiliconVersion.Settings.ShowTerminalNormally = oldPrefs.ShowTerminalNormally
	turtleSiliconVersion.Settings.EnvironmentVariables = oldPrefs.EnvironmentVariables
	turtleSiliconVersion.Settings.ReduceTerrainDistance = oldPrefs.ReduceTerrainDistance
	turtleSiliconVersion.Settings.SetMultisampleTo2x = oldPrefs.SetMultisampleTo2x
	turtleSiliconVersion.Settings.SetShadowLOD0 = oldPrefs.SetShadowLOD0
	turtleSiliconVersion.Settings.EnableLibSiliconPatch = oldPrefs.EnableLibSiliconPatch
	turtleSiliconVersion.Settings.UserDisabledShadowLOD = oldPrefs.UserDisabledShadowLOD
	turtleSiliconVersion.Settings.UserDisabledLibSiliconPatch = oldPrefs.UserDisabledLibSiliconPatch

	// Save the updated version
	if err := currentVersionManager.UpdateVersion(turtleSiliconVersion); err != nil {
		return err
	}

	// Clear the old paths from prefs.json to prevent future overrides
	oldPrefs.TurtleWoWPath = ""
	oldPrefs.CrossOverPath = ""
	if err := utils.SavePrefs(oldPrefs); err != nil {
		debug.Printf("Warning: failed to clear old paths from prefs.json: %v", err)
	} else {
		debug.Printf("Cleared old paths from prefs.json after migration")
	}

	debug.Printf("Successfully migrated old preferences to TurtleSilicon version")
	return nil
}

// checkDefaultCrossOverPathForAllVersions checks and sets default CrossOver path for all versions
func checkDefaultCrossOverPathForAllVersions() {
	defaultCrossOverPath := "/Applications/CrossOver.app"

	// Check if default path exists
	if info, err := os.Stat(defaultCrossOverPath); err != nil || !info.IsDir() {
		debug.Printf("Default CrossOver path not found: %s", defaultCrossOverPath)
		return
	}

	debug.Printf("Default CrossOver path found: %s", defaultCrossOverPath)

	// Set default path for all versions that don't have a CrossOver path set
	for _, versionID := range currentVersionManager.GetVersionList() {
		version, err := currentVersionManager.GetVersion(versionID)
		if err != nil {
			continue
		}

		if version.CrossOverPath == "" {
			version.CrossOverPath = defaultCrossOverPath
			currentVersionManager.UpdateVersion(version)
			debug.Printf("Set default CrossOver path for version %s: %s", versionID, defaultCrossOverPath)
		}
	}
}

// SetupVersionDropdown configures the version dropdown in the UI
func SetupVersionDropdown(myWindow fyne.Window) {
	if VersionDropdown == nil {
		debug.Printf("VersionDropdown is nil, cannot setup")
		return
	}

	// Get all versions for the dropdown in the specified order
	versionOrder := []string{"turtlesilicon", "epochsilicon", "vanillasilicon", "burningsilicon", "wrathsilicon"}
	versions := []string{}
	for _, versionID := range versionOrder {
		if ver, err := currentVersionManager.GetVersion(versionID); err == nil {
			versions = append(versions, ver.DisplayName)
		}
	}

	VersionDropdown.Options = versions

	// Set current selection
	if currentVersion != nil {
		VersionDropdown.SetSelected(currentVersion.DisplayName)
	}

	// Set callback for version changes
	VersionDropdown.OnChanged = func(selected string) {
		onVersionChanged(selected, myWindow)
	}

	// Setup the new title button approach
	setupVersionTitleButton(myWindow)
}

// setupVersionTitleButton configures the custom version title button
func setupVersionTitleButton(myWindow fyne.Window) {
	if VersionTitleButton == nil || VersionTitleText == nil {
		debug.Printf("Version title components are nil, cannot setup")
		return
	}

	// Update the title text to show current version
	updateVersionTitleText()

	// Set callback for the title button to show version selection popup
	VersionTitleButton.OnTapped = func() {
		showVersionSelectionPopup(myWindow)
	}
}

// updateVersionTitleText updates the large title text with the current version name
func updateVersionTitleText() {
	if VersionTitleText == nil || currentVersion == nil {
		return
	}

	// Use markdown for large, prominent text
	titleMarkdown := fmt.Sprintf("# %s", currentVersion.DisplayName)
	VersionTitleText.ParseMarkdown(titleMarkdown)
}

// showVersionSelectionPopup shows a popup for version selection
func showVersionSelectionPopup(myWindow fyne.Window) {
	if currentVersionManager == nil {
		return
	}

	// Get all versions in the specified order
	versionOrder := []string{"turtlesilicon", "epochsilicon", "vanillasilicon", "burningsilicon", "wrathsilicon"}

	// Create popup content container first so we can reference it
	popupContent := container.NewVBox()
	popup := widget.NewModalPopUp(container.NewPadded(popupContent), myWindow.Canvas())

	// Create header with title and close button
	popupTitle := widget.NewRichTextFromMarkdown("## Select Version")
	closeButton := widget.NewButton("✕", func() {
		popup.Hide()
	})
	closeButton.Importance = widget.LowImportance

	headerContainer := container.NewBorder(
		nil, nil, closeButton, nil,
		container.NewCenter(popupTitle),
	)

	// Create version buttons with consistent width using grid layout
	var versionButtons []fyne.CanvasObject
	for _, versionID := range versionOrder {
		if ver, err := currentVersionManager.GetVersion(versionID); err == nil {
			versionName := ver.DisplayName
			versionButton := widget.NewButton(versionName, func(selectedName string) func() {
				return func() {
					popup.Hide()
					onVersionChanged(selectedName, myWindow)
				}
			}(versionName))

			// Highlight current version
			if currentVersion != nil && ver.ID == currentVersion.ID {
				versionButton.Importance = widget.HighImportance
			} else {
				versionButton.Importance = widget.MediumImportance
			}

			versionButtons = append(versionButtons, versionButton)
		}
	}

	// Build the popup content
	popupContent.Add(headerContainer)
	popupContent.Add(widget.NewSeparator())

	// Create a grid container for consistent button widths
	buttonsGrid := container.NewGridWithColumns(1)
	for _, button := range versionButtons {
		buttonsGrid.Add(button)
	}

	// Add the grid with some padding
	popupContent.Add(container.NewPadded(buttonsGrid))

	// Size and show popup - smaller and more compact
	popup.Resize(fyne.NewSize(350, 300))
	popup.Show()
}

// onVersionChanged handles version selection changes
func onVersionChanged(selectedDisplayName string, myWindow fyne.Window) {
	debug.Printf("Version changed to: %s", selectedDisplayName)

	// Find the version ID by display name
	var selectedVersionID string
	for _, versionID := range currentVersionManager.GetVersionList() {
		if ver, err := currentVersionManager.GetVersion(versionID); err == nil {
			if ver.DisplayName == selectedDisplayName {
				selectedVersionID = versionID
				break
			}
		}
	}

	if selectedVersionID == "" {
		debug.Printf("Error: Could not find version ID for display name: %s", selectedDisplayName)
		return
	}

	// Stop any running rosettax87 service before switching versions
	// Each version should have its own independent service
	service.StopRosettaX87ServiceSilent()

	// Switch to the new version
	if err := currentVersionManager.SetCurrentVersion(selectedVersionID); err != nil {
		debug.Printf("Error switching to version %s: %v", selectedVersionID, err)
		dialog.ShowError(fmt.Errorf("failed to switch to version %s: %v", selectedDisplayName, err), myWindow)
		return
	}

	// Update current version reference
	var err error
	currentVersion, err = currentVersionManager.GetCurrentVersion()
	if err != nil {
		debug.Printf("Error getting current version after switch: %v", err)
		return
	}

	// Sync legacy paths for backward compatibility with service and other components
	syncLegacyPaths()

	// Update all UI elements for the new version
	RefreshUIForCurrentVersion()
	updateUIForCurrentVersion()
	UpdateAllStatuses()

	// Update the title text to show the new version
	updateVersionTitleText()

	// Update the logo to match the new version
	updateLogoForVersion(selectedVersionID)

	// For EpochSilicon, automatically check for updates if already patched
	if selectedVersionID == "epochsilicon" && currentVersion != nil && currentVersion.GamePath != "" {
		// Check if we're already patched by looking for required files
		missingFiles, err := epochsilicon.CheckEpochSiliconFiles(currentVersion.GamePath)
		if err == nil && len(missingFiles) == 0 {
			// All files exist, so we're patched - check for updates
			debug.Printf("EpochSilicon is already patched, checking for updates...")
			epochsilicon.CheckForUpdatesWithProgress(myWindow, currentVersion.GamePath, func(updatesAvailable []epochsilicon.RequiredFile, err error) {
				if err != nil {
					debug.Printf("Failed to check for updates: %v", err)
				} else if len(updatesAvailable) > 0 {
					epochsilicon.ShowUpdatePromptDialog(myWindow, updatesAvailable, func() {
						epochsilicon.DownloadMissingFiles(myWindow, currentVersion.GamePath, updatesAvailable, func(success bool) {
							if success {
								dialog.ShowInformation("Update Complete", "All Project Epoch files have been updated successfully!", myWindow)
								// Refresh the UI to reflect any changes
								UpdateAllStatuses()
							}
						})
					})
				} else {
					debug.Printf("EpochSilicon is up to date")
				}
			})
		}
	}

	debug.Printf("Successfully switched to version: %s", selectedDisplayName)
}

// syncLegacyPaths updates legacy path variables for backward compatibility
func syncLegacyPaths() {
	if currentVersion == nil {
		return
	}

	// Update legacy paths for backward compatibility with service and other components
	paths.TurtlewowPath = currentVersion.GamePath
	paths.CrossoverPath = currentVersion.CrossOverPath

	debug.Printf("Synced legacy paths: TurtleWoW=%s, CrossOver=%s", paths.TurtlewowPath, paths.CrossoverPath)
}

// updateUIForCurrentVersion updates all UI elements based on the current version
func updateUIForCurrentVersion() {
	if currentVersion == nil {
		return
	}

	// Update path labels
	updateVersionPathLabels()

	// Update checkboxes and settings
	updateVersionSettings()

	// Update button states based on version capabilities
	updateVersionCapabilities()
}

// updateVersionPathLabels updates the path labels for the current version
func updateVersionPathLabels() {
	if currentVersion == nil || crossoverPathLabel == nil || turtlewowPathLabel == nil {
		return
	}

	// Update CrossOver path label
	if currentVersion.CrossOverPath == "" {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
	} else {
		crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: currentVersion.CrossOverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
	}
	crossoverPathLabel.Refresh()

	// Update Game path label (rename from TurtleWoW for other versions)
	if currentVersion.GamePath == "" {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
	} else {
		turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: currentVersion.GamePath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
	}
	turtlewowPathLabel.Refresh()
}

// updateVersionSettings updates all checkboxes and settings for the current version
func updateVersionSettings() {
	if currentVersion == nil {
		return
	}

	settings := currentVersion.Settings

	// Update checkboxes if they exist
	if metalHudCheckbox != nil {
		metalHudCheckbox.SetChecked(settings.EnableMetalHud)
	}
	if showTerminalCheckbox != nil {
		showTerminalCheckbox.SetChecked(settings.ShowTerminalNormally)
	}
	if vanillaTweaksCheckbox != nil {
		vanillaTweaksCheckbox.SetChecked(settings.EnableVanillaTweaks)
		if !currentVersion.SupportsVanillaTweaks {
			vanillaTweaksCheckbox.Disable()
		} else {
			vanillaTweaksCheckbox.Enable()
		}
		vanillaTweaksCheckbox.Refresh()
	}
	if autoDeleteWdbCheckbox != nil {
		autoDeleteWdbCheckbox.SetChecked(settings.AutoDeleteWdb)
	}

	// Update graphics settings checkboxes
	if reduceTerrainDistanceCheckbox != nil {
		reduceTerrainDistanceCheckbox.SetChecked(settings.ReduceTerrainDistance)
	}
	if setMultisampleTo2xCheckbox != nil {
		setMultisampleTo2xCheckbox.SetChecked(settings.SetMultisampleTo2x)
	}
	if setShadowLOD0Checkbox != nil {
		setShadowLOD0Checkbox.SetChecked(settings.SetShadowLOD0)
	}
	if libSiliconPatchCheckbox != nil {
		// libSiliconPatch is only available for TurtleSilicon
		if currentVersion.ID == "turtlesilicon" {
			libSiliconPatchCheckbox.SetChecked(settings.EnableLibSiliconPatch)
			libSiliconPatchCheckbox.Enable()
		} else {
			libSiliconPatchCheckbox.SetChecked(false)
			libSiliconPatchCheckbox.Disable()
		}
	}

	// Update environment variables entry
	if envVarsEntry != nil {
		envVarsEntry.SetText(settings.EnvironmentVariables)
	}
}

// updateVersionCapabilities updates UI elements based on version capabilities
func updateVersionCapabilities() {
	if currentVersion == nil {
		return
	}

	// Update vanilla tweaks checkbox availability
	if vanillaTweaksCheckbox != nil {
		if currentVersion.SupportsVanillaTweaks {
			vanillaTweaksCheckbox.Enable()
		} else {
			vanillaTweaksCheckbox.Disable()
			vanillaTweaksCheckbox.SetChecked(false)
		}
	}

	// Update libSiliconPatch checkbox availability (TurtleSilicon only)
	if libSiliconPatchCheckbox != nil {
		if currentVersion.ID == "turtlesilicon" {
			libSiliconPatchCheckbox.Enable()
		} else {
			libSiliconPatchCheckbox.Disable()
			libSiliconPatchCheckbox.SetChecked(false)
		}
	}
}

// GetCurrentVersionManager returns the current version manager
func GetCurrentVersionManager() *version.VersionManager {
	return currentVersionManager
}

// GetCurrentVersion returns the current version
func GetCurrentVersion() *version.GameVersion {
	return currentVersion
}

// SaveCurrentVersion saves the current version's settings
func SaveCurrentVersion(ver *version.GameVersion) error {
	if currentVersionManager == nil {
		return fmt.Errorf("version manager not initialized")
	}

	// Update the version in the manager
	if err := currentVersionManager.UpdateVersion(ver); err != nil {
		return fmt.Errorf("failed to update version: %v", err)
	}

	// Update the current version reference
	currentVersion = ver

	debug.Printf("Saved settings for version: %s", ver.ID)
	return nil
}

// RefreshUIForCurrentVersion updates all UI components to reflect the current version's settings
func RefreshUIForCurrentVersion() {
	if currentVersion == nil {
		return
	}

	debug.Printf("Refreshing UI for version: %s", currentVersion.DisplayName)

	// Update the logo to match the current version
	updateLogoForVersion(currentVersion.ID)

	// Update launcher variables to match current version settings
	launcher.EnableMetalHud = currentVersion.Settings.EnableMetalHud
	launcher.EnableVanillaTweaks = currentVersion.Settings.EnableVanillaTweaks
	launcher.AutoDeleteWdb = currentVersion.Settings.AutoDeleteWdb
	launcher.CustomEnvVars = currentVersion.Settings.EnvironmentVariables

	// Update UI checkboxes to reflect current version settings
	if metalHudCheckbox != nil {
		metalHudCheckbox.SetChecked(currentVersion.Settings.EnableMetalHud)
	}
	if vanillaTweaksCheckbox != nil {
		vanillaTweaksCheckbox.SetChecked(currentVersion.Settings.EnableVanillaTweaks)
	}
	if autoDeleteWdbCheckbox != nil {
		autoDeleteWdbCheckbox.SetChecked(currentVersion.Settings.AutoDeleteWdb)
	}
	if showTerminalCheckbox != nil {
		showTerminalCheckbox.SetChecked(currentVersion.Settings.ShowTerminalNormally)
	}
	if envVarsEntry != nil {
		envVarsEntry.SetText(currentVersion.Settings.EnvironmentVariables)
	}

	// Update graphics settings checkboxes
	if reduceTerrainDistanceCheckbox != nil {
		reduceTerrainDistanceCheckbox.SetChecked(currentVersion.Settings.ReduceTerrainDistance)
	}
	if setMultisampleTo2xCheckbox != nil {
		setMultisampleTo2xCheckbox.SetChecked(currentVersion.Settings.SetMultisampleTo2x)
	}
	if setShadowLOD0Checkbox != nil {
		setShadowLOD0Checkbox.SetChecked(currentVersion.Settings.SetShadowLOD0)
	}
	if libSiliconPatchCheckbox != nil {
		libSiliconPatchCheckbox.SetChecked(currentVersion.Settings.EnableLibSiliconPatch)
	}
}

// Version-aware path selection
func SelectCurrentVersionGamePath(myWindow fyne.Window) {
	if currentVersion == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), myWindow)
		return
	}

	if currentVersionManager == nil {
		dialog.ShowError(fmt.Errorf("version manager not initialized"), myWindow)
		return
	}

	// Create larger folder dialog - 5/6 of window size
	windowSize := myWindow.Canvas().Size()
	dialogWidth := windowSize.Width * 5 / 6
	dialogHeight := windowSize.Height * 5 / 6

	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			debug.Printf("Game path selection cancelled for version %s", currentVersion.ID)
			return
		}
		selectedPath := uri.Path()

		currentVersion.GamePath = selectedPath
		if err := currentVersionManager.UpdateVersion(currentVersion); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save game path: %v", err), myWindow)
			return
		}

		// For TurtleSilicon, also update prefs.json for backward compatibility
		if currentVersion.ID == "turtlesilicon" {
			if prefs, err := utils.LoadPrefs(); err == nil {
				prefs.TurtleWoWPath = selectedPath
				if err := utils.SavePrefs(prefs); err != nil {
					debug.Printf("Warning: failed to update TurtleWoWPath in prefs.json: %v", err)
				} else {
					debug.Printf("Updated TurtleWoWPath in prefs.json for backward compatibility")
				}
			}
		}

		// Reset patching status for this version
		paths.SetVersionPatchingStatus(currentVersion.ID, false, false)

		// Sync legacy paths
		syncLegacyPaths()

		debug.Printf("Game path set for version %s: %s", currentVersion.ID, selectedPath)
		updateVersionPathLabels()
		UpdateAllStatuses()

		// For EpochSilicon, check required files and offer to download missing ones
		if currentVersion.ID == "epochsilicon" {
			checkEpochSiliconFiles(myWindow, selectedPath)
		}
	}, myWindow)

	folderDialog.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	folderDialog.Show()
}

// Version-aware CrossOver path selection
func SelectCurrentVersionCrossOverPath(myWindow fyne.Window) {
	if currentVersion == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), myWindow)
		return
	}

	// Create larger folder dialog - 5/6 of window size
	windowSize := myWindow.Canvas().Size()
	dialogWidth := windowSize.Width * 5 / 6
	dialogHeight := windowSize.Height * 5 / 6

	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			debug.Printf("CrossOver path selection cancelled for version %s", currentVersion.ID)
			return
		}
		selectedPath := uri.Path()

		currentVersion.CrossOverPath = selectedPath
		if err := currentVersionManager.UpdateVersion(currentVersion); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save CrossOver path: %v", err), myWindow)
			return
		}

		// For TurtleSilicon, also update prefs.json for backward compatibility
		if currentVersion.ID == "turtlesilicon" {
			if prefs, err := utils.LoadPrefs(); err == nil {
				prefs.CrossOverPath = selectedPath
				if err := utils.SavePrefs(prefs); err != nil {
					debug.Printf("Warning: failed to update CrossOverPath in prefs.json: %v", err)
				} else {
					debug.Printf("Updated CrossOverPath in prefs.json for backward compatibility")
				}
			}
		}

		// Reset patching status for this version
		paths.SetVersionPatchingStatus(currentVersion.ID, false, false)

		// Sync legacy paths
		syncLegacyPaths()

		debug.Printf("CrossOver path set for version %s: %s", currentVersion.ID, selectedPath)
		updateVersionPathLabels()
		UpdateAllStatuses()
	}, myWindow)

	folderDialog.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	folderDialog.Show()
}

// Version-aware patching
func PatchCurrentVersion(myWindow fyne.Window) {
	if currentVersion == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), myWindow)
		return
	}

	// For EpochSilicon, check for updates and missing files before patching
	if currentVersion.ID == "epochsilicon" {
		// Check for updates with progress indicator
		epochsilicon.CheckForUpdatesWithProgress(myWindow, currentVersion.GamePath, func(updatesAvailable []epochsilicon.RequiredFile, err error) {
			if err != nil {
				debug.Printf("Failed to check for updates: %v", err)
				// Fall back to missing files check
				missingFiles, err := epochsilicon.CheckEpochSiliconFiles(currentVersion.GamePath)
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				if len(missingFiles) > 0 {
					epochsilicon.ShowMissingFilesDialog(myWindow, missingFiles, func() {
						epochsilicon.DownloadMissingFiles(myWindow, currentVersion.GamePath, missingFiles, func(success bool) {
							if success {
								// After successful download, proceed with patching
								proceedWithPatching(myWindow)
							}
						})
					})
				} else {
					// No missing files, proceed with patching
					proceedWithPatching(myWindow)
				}
			} else if len(updatesAvailable) > 0 {
				epochsilicon.ShowUpdatePromptDialog(myWindow, updatesAvailable, func() {
					epochsilicon.DownloadMissingFiles(myWindow, currentVersion.GamePath, updatesAvailable, func(success bool) {
						if success {
							// After successful update, proceed with patching
							proceedWithPatching(myWindow)
						}
					})
				})
			} else {
				// No updates needed, proceed with patching
				proceedWithPatching(myWindow)
			}
		})
		return
	}

	// Proceed with normal patching
	proceedWithPatching(myWindow)
}

// proceedWithPatching performs the actual patching operation
func proceedWithPatching(myWindow fyne.Window) {
	patching.PatchVersionGame(myWindow, UpdateAllStatuses, currentVersion.GamePath, currentVersion.UsesRosettaPatching, currentVersion.UsesDivxDecoderPatch)

	// Update patching status
	gamePatched := patching.CheckVersionPatchingStatus(currentVersion.GamePath, currentVersion.UsesRosettaPatching, currentVersion.UsesDivxDecoderPatch)
	crossoverPatched, _ := paths.GetVersionPatchingStatus(currentVersion.ID)
	paths.SetVersionPatchingStatus(currentVersion.ID, gamePatched, crossoverPatched)
}

// Version-aware unpatching
func UnpatchCurrentVersion(myWindow fyne.Window) {
	if currentVersion == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), myWindow)
		return
	}

	patching.UnpatchVersionGame(myWindow, UpdateAllStatuses, currentVersion.GamePath, currentVersion.UsesRosettaPatching, currentVersion.UsesDivxDecoderPatch)

	// Update patching status
	gamePatched := patching.CheckVersionPatchingStatus(currentVersion.GamePath, currentVersion.UsesRosettaPatching, currentVersion.UsesDivxDecoderPatch)
	crossoverPatched, _ := paths.GetVersionPatchingStatus(currentVersion.ID)
	paths.SetVersionPatchingStatus(currentVersion.ID, gamePatched, crossoverPatched)
}

// Version-aware launching
func LaunchCurrentVersion(myWindow fyne.Window) {
	if currentVersion == nil {
		dialog.ShowError(fmt.Errorf("no current version selected"), myWindow)
		return
	}

	launcher.LaunchVersionGame(
		myWindow,
		currentVersion.ID,
		currentVersion.GamePath,
		currentVersion.CrossOverPath,
		currentVersion.ExecutableName,
		currentVersion.Settings.EnableMetalHud,
		currentVersion.Settings.EnvironmentVariables,
		currentVersion.Settings.AutoDeleteWdb,
	)
}

// checkEpochSiliconFiles checks for required EpochSilicon files and offers to download missing ones
func checkEpochSiliconFiles(myWindow fyne.Window, gamePath string) {
	// Always update realmlist.wtf for EpochSilicon when setting path
	go func() {
		if err := epochsilicon.UpdateRealmlistForEpochSilicon(gamePath); err != nil {
			debug.Printf("Warning: Failed to update realmlist.wtf: %v", err)
		}
	}()

	// Check for updates with progress indicator
	epochsilicon.CheckForUpdatesWithProgress(myWindow, gamePath, func(updatesAvailable []epochsilicon.RequiredFile, err error) {
		if err != nil {
			debug.Printf("Failed to check for updates: %v", err)
			// Fall back to missing files check
			missingFiles, err := epochsilicon.CheckEpochSiliconFiles(gamePath)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			if len(missingFiles) > 0 {
				epochsilicon.ShowMissingFilesDialog(myWindow, missingFiles, func() {
					epochsilicon.DownloadMissingFiles(myWindow, gamePath, missingFiles, func(success bool) {
						if success {
							dialog.ShowInformation("Download Complete", "All Project Epoch files have been downloaded successfully!", myWindow)
							// Refresh the UI to reflect any changes
							UpdateAllStatuses()
						}
					})
				})
			}
		} else if len(updatesAvailable) > 0 {
			epochsilicon.ShowUpdatePromptDialog(myWindow, updatesAvailable, func() {
				epochsilicon.DownloadMissingFiles(myWindow, gamePath, updatesAvailable, func(success bool) {
					if success {
						dialog.ShowInformation("Update Complete", "All Project Epoch files have been updated successfully!", myWindow)
						// Refresh the UI to reflect any changes
						UpdateAllStatuses()
					}
				})
			})
		}
	})
}
