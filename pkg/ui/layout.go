package ui

import (
	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createHeaderContainer creates the header with title and subtitle
func createHeaderContainer() fyne.CanvasObject {
	// Main title
	titleText := widget.NewRichTextFromMarkdown("# TurtleSilicon")
	titleText.Wrapping = fyne.TextWrapOff

	// Subtitle
	subtitleText := widget.NewLabel("A TurtleWoW launcher for Apple Silicon Macs")
	subtitleText.Alignment = fyne.TextAlignCenter

	// Create header container
	headerContainer := container.NewVBox(
		container.NewCenter(titleText),
		container.NewCenter(subtitleText),
	)

	return headerContainer
}

// createLogoContainer creates and returns the application logo container
func createLogoContainer() fyne.CanvasObject {
	// Load the application logo
	logoResource, err := fyne.LoadResourceFromPath("Icon.png")
	if err != nil {
		debug.Printf("Warning: could not load logo: %v", err)
	}

	// Create the logo image with a smaller fixed size since we have a header now
	var logoImage *canvas.Image
	if logoResource != nil {
		logoImage = canvas.NewImageFromResource(logoResource)
		logoImage.FillMode = canvas.ImageFillContain
		logoImage.SetMinSize(fyne.NewSize(80, 80))
	}

	// Create a container to center the logo
	var logoContainer fyne.CanvasObject
	if logoImage != nil {
		logoContainer = container.NewCenter(logoImage)
	} else {
		// If logo couldn't be loaded, add an empty space for consistent layout
		logoContainer = container.NewCenter(widget.NewLabel(""))
	}

	return logoContainer
}

// createPathSelectionForm creates the form for selecting CrossOver and TurtleWoW paths
func createPathSelectionForm(myWindow fyne.Window) *widget.Form {
	pathSelectionForm := widget.NewForm(
		widget.NewFormItem("CrossOver Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			paths.SelectCrossOverPath(myWindow, crossoverPathLabel, UpdateAllStatuses)
		}), crossoverPathLabel)),
		widget.NewFormItem("TurtleWoW Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", func() {
			paths.SelectTurtleWoWPath(myWindow, turtlewowPathLabel, UpdateAllStatuses)
		}), turtlewowPathLabel)),
	)

	return pathSelectionForm
}

// createPatchOperationsLayout creates the layout for patch operations
func createPatchOperationsLayout() fyne.CanvasObject {
	patchOperationsLayout := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(4,
			widget.NewLabel("TurtleWoW Patch:"), turtlewowStatusLabel, patchTurtleWoWButton, unpatchTurtleWoWButton,
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
