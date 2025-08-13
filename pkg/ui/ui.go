package ui

import (
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/epochsilicon"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func CreateUI(myWindow fyne.Window) fyne.CanvasObject {
	// Initialize UI component variables
	crossoverPathLabel = widget.NewRichText()
	turtlewowPathLabel = widget.NewRichText()
	turtlewowStatusLabel = widget.NewRichText()
	crossoverStatusLabel = widget.NewRichText()
	serviceStatusLabel = widget.NewRichText()

	// Initialize version system
	if err := InitializeVersionSystem(); err != nil {
		debug.Printf("Error initializing version system: %v", err)
		// Fall back to old system if version system fails
		prefs, _ := utils.LoadPrefs()
		if prefs.TurtleWoWPath != "" {
			paths.TurtlewowPath = prefs.TurtleWoWPath
		}
		if prefs.CrossOverPath != "" {
			paths.CrossoverPath = prefs.CrossOverPath
		}
	}

	// Check for first-time users after UI is ready
	defer func() {
		CheckForFirstTimeUser(myWindow)
	}()

	// Create all UI components
	createOptionsComponents()
	createPatchingButtons(myWindow)
	createServiceButtons(myWindow)
	createLaunchButton(myWindow)

	// Check default CrossOver path
	paths.CheckDefaultCrossOverPath()

	// Check graphics settings presence and set default state
	patching.CheckGraphicsSettingsPresence()

	// Load graphics settings from Config.wtf and update UI
	if err := patching.LoadGraphicsSettingsFromConfig(); err != nil {
		// Log error but continue - this is not critical for app startup
		debug.Printf("Warning: failed to load graphics settings from Config.wtf: %v", err)
	} else {
		// Refresh checkbox states to reflect loaded settings
		refreshGraphicsSettingsCheckboxes()
	}

	// Create header, main content and bottom bar
	headerContent := createHeaderContainer()
	mainContent := createMainContent(myWindow)
	bottomBar := createBottomBar(myWindow)

	// Setup version dropdown after UI components are created
	SetupVersionDropdown(myWindow)

	// Refresh UI to display current version settings and paths
	RefreshUIForCurrentVersion()

	// Initial UI state update
	UpdateAllStatuses()

	// For EpochSilicon, automatically check for updates on app launch if already patched
	go func() {
		if currentVersion := GetCurrentVersion(); currentVersion != nil && currentVersion.ID == "epochsilicon" && currentVersion.GamePath != "" {
			// Check if we're already patched by looking for required files
			if missingFiles, err := epochsilicon.CheckEpochSiliconFiles(currentVersion.GamePath); err == nil {
				if len(missingFiles) == 0 {
					// All files exist, so we're patched - check for updates
					debug.Printf("EpochSilicon detected on startup, checking for updates...")
					epochsilicon.CheckForUpdatesWithProgress(myWindow, currentVersion.GamePath, func(updatesAvailable []epochsilicon.RequiredFile, err error) {
						if err != nil {
							debug.Printf("Failed to check for updates on startup: %v", err)
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
							debug.Printf("EpochSilicon is up to date on startup")
						}
					})
				} else {
					// Missing files detected - show download dialog
					debug.Printf("EpochSilicon missing files detected on startup: %d files", len(missingFiles))
					epochsilicon.ShowMissingFilesDialog(myWindow, missingFiles, func() {
						epochsilicon.DownloadMissingFiles(myWindow, currentVersion.GamePath, missingFiles, func(success bool) {
							if success {
								dialog.ShowInformation("Download Complete", "All Project Epoch files have been downloaded successfully!", myWindow)
								// Refresh the UI to reflect any changes
								UpdateAllStatuses()
							}
						})
					})
				}
			}
		}
	}()

	// Create layout with header at top, main content moved up to avoid bottom bar, and bottom bar
	// Use VBox to position main content higher up instead of centering it
	mainContentContainer := container.NewVBox(
		mainContent,
	)

	// Add horizontal padding to the main content
	paddedMainContent := container.NewPadded(mainContentContainer)

	layout := container.NewBorder(
		headerContent,     // top
		bottomBar,         // bottom
		nil,               // left
		nil,               // right
		paddedMainContent, // main content with horizontal padding
	)

	return layout
}
