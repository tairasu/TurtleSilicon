package ui

import (
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/launcher"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	// Check for new TurtleWoW users after UI is ready
	defer func() {
		CheckForFirstTimeUser(myWindow)
		// Show new user popup for TurtleWoW if needed
		CheckAndShowNewUserPopup()
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

	// Set up callback for status updates from other packages
	SetStatusUpdateCallback(UpdateAllStatuses)

	// Set up launcher callback for triggering UI updates when patch status changes
	launcher.SetUIUpdateCallback(UpdateAllStatuses)

	// Initial UI state update
	UpdateAllStatuses()

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
