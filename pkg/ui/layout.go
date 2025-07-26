package ui

import (
	"turtlesilicon/pkg/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createHeaderContainer() fyne.CanvasObject {
	versionTitleText := widget.NewRichTextFromMarkdown("# TurtleSilicon")
	versionTitleText.Wrapping = fyne.TextWrapOff

	versionTitleButton := widget.NewButton("", func() {
		debug.Printf("Version title button clicked")
	})

	versionTitleButton.Importance = widget.MediumImportance

	versionTitleContainer := container.NewStack(
		versionTitleButton,
		container.NewCenter(versionTitleText),
	)

	versionTitleContainer.Resize(fyne.NewSize(600, 80))

	versionDropdown := widget.NewSelect([]string{"TurtleSilicon"}, func(selected string) {
		debug.Printf("Version selected: %s", selected)
	})
	versionDropdown.SetSelected("TurtleSilicon")
	versionDropdown.Hide()

	titleDropdownContainer := container.NewCenter(versionTitleContainer)

	VersionDropdown = versionDropdown
	VersionTitleButton = versionTitleButton
	VersionTitleText = versionTitleText

	// Add click hint text
	clickHintText := widget.NewLabel("Click to change version")
	clickHintText.Alignment = fyne.TextAlignCenter
	clickHintText.TextStyle = fyne.TextStyle{Italic: true}

	subtitleText := widget.NewLabel("A Vanilla WoW launcher for Apple Silicon Macs")
	subtitleText.Alignment = fyne.TextAlignCenter

	headerContainer := container.NewVBox(
		titleDropdownContainer,
		clickHintText,
		subtitleText,
	)

	return headerContainer
}

// getVersionIconPath returns the appropriate icon path for the given version
func getVersionIconPath(versionID string) string {
	switch versionID {
	case "epochsilicon":
		return "img/project-epoch.png"
	case "turtlesilicon":
		return "Icon.png" // Default TurtleSilicon icon
	default:
		return "Icon.png" // Fallback to default icon for other versions
	}
}

// createLogoContainer creates and returns the application logo container
func createLogoContainer() fyne.CanvasObject {
	// Get the current version to determine which icon to use
	currentVer := GetCurrentVersion()
	var iconPath string
	if currentVer != nil {
		iconPath = getVersionIconPath(currentVer.ID)
	} else {
		iconPath = "Icon.png" // Default fallback
	}

	// Load the version-specific logo
	logoResource, err := fyne.LoadResourceFromPath(iconPath)
	if err != nil {
		debug.Printf("Warning: could not load logo from %s: %v", iconPath, err)
		// Try fallback to default icon
		logoResource, err = fyne.LoadResourceFromPath("Icon.png")
		if err != nil {
			debug.Printf("Warning: could not load fallback logo: %v", err)
		}
	}

	// Create the logo image with a smaller fixed size since we have a header now
	if logoResource != nil {
		logoImage = canvas.NewImageFromResource(logoResource)
		logoImage.FillMode = canvas.ImageFillContain
		logoImage.SetMinSize(fyne.NewSize(80, 80))
	}

	// Create a container to center the logo
	if logoImage != nil {
		logoContainer = container.NewCenter(logoImage)
	} else {
		// If logo couldn't be loaded, add an empty space for consistent layout
		logoContainer = container.NewCenter(widget.NewLabel(""))
	}

	return logoContainer
}

// updateLogoForVersion updates the logo to match the current version
func updateLogoForVersion(versionID string) {
	if logoImage == nil {
		debug.Printf("Warning: logoImage is nil, cannot update")
		return
	}

	iconPath := getVersionIconPath(versionID)
	logoResource, err := fyne.LoadResourceFromPath(iconPath)
	if err != nil {
		debug.Printf("Warning: could not load logo from %s: %v", iconPath, err)
		// Try fallback to default icon
		logoResource, err = fyne.LoadResourceFromPath("Icon.png")
		if err != nil {
			debug.Printf("Warning: could not load fallback logo: %v", err)
			return
		}
	}

	// Update the logo image resource
	logoImage.Resource = logoResource
	logoImage.Refresh()
	debug.Printf("Updated logo to: %s", iconPath)
}

// createPathSelectionForm creates the form for selecting CrossOver and game paths
func createPathSelectionForm(myWindow fyne.Window) *widget.Form {
	pathSelectionForm := widget.NewForm(
		widget.NewFormItem("CrossOver Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			SelectCurrentVersionCrossOverPath(myWindow)
		}), crossoverPathLabel)),
		widget.NewFormItem("Game Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			SelectCurrentVersionGamePath(myWindow)
		}), turtlewowPathLabel)),
	)

	return pathSelectionForm
}

// createPatchOperationsLayout creates the layout for patch operations
func createPatchOperationsLayout() fyne.CanvasObject {
	patchOperationsLayout := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(4,
			widget.NewLabel("Game Patch:"), turtlewowStatusLabel, patchTurtleWoWButton, unpatchTurtleWoWButton,
		),
		container.NewGridWithColumns(4,
			widget.NewLabel("CrossOver Patch:"), crossoverStatusLabel, patchCrossOverButton, unpatchCrossOverButton,
		),
		container.NewGridWithColumns(4,
			widget.NewLabel("RosettaX87 Service:"), serviceStatusLabel, startServiceButton, stopServiceButton,
		),
		widget.NewSeparator(),
	)

	return patchOperationsLayout
}

// createMainContent creates the main content area of the application
func createMainContent(myWindow fyne.Window) fyne.CanvasObject {
	logoContainer := createLogoContainer()
	pathSelectionForm := createPathSelectionForm(myWindow)
	patchOperationsLayout := createPatchOperationsLayout()

	// Create main content area with better spacing
	mainContent := container.NewVBox(
		logoContainer,
		pathSelectionForm,
		patchOperationsLayout,
	)

	return mainContent
}
