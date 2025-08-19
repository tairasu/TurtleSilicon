package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// UI component variables - centralized for easy access across modules
var (
	// Status labels
	crossoverPathLabel   *widget.RichText
	turtlewowPathLabel   *widget.RichText
	turtlewowStatusLabel *widget.RichText
	crossoverStatusLabel *widget.RichText
	serviceStatusLabel   *widget.RichText

	// Version management
	VersionDropdown    *widget.Select
	VersionTitleButton *widget.Button
	VersionTitleText   *widget.RichText

	// Action buttons
	launchButton           *widget.Button
	playButton             *widget.Button
	playButtonText         *widget.RichText
	patchTurtleWoWButton   *widget.Button
	patchCrossOverButton   *widget.Button
	unpatchTurtleWoWButton *widget.Button
	unpatchCrossOverButton *widget.Button
	startServiceButton     *widget.Button
	stopServiceButton      *widget.Button

	// Button containers
	leftButtons *fyne.Container

	// Option checkboxes
	metalHudCheckbox      *widget.Check
	showTerminalCheckbox  *widget.Check
	vanillaTweaksCheckbox *widget.Check
	autoDeleteWdbCheckbox *widget.Check

	// Recommended settings button
	applyRecommendedSettingsButton *widget.Button
	recommendedSettingsHelpButton  *widget.Button

	// Wine registry buttons and status
	enableOptionAsAltButton  *widget.Button
	disableOptionAsAltButton *widget.Button
	optionAsAltStatusLabel   *widget.RichText

	// Environment variables entry
	envVarsEntry *widget.Entry

	// Graphics settings checkboxes
	reduceTerrainDistanceCheckbox *widget.Check
	setMultisampleTo2xCheckbox    *widget.Check
	setShadowLOD0Checkbox         *widget.Check
	applyGraphicsSettingsButton   *widget.Button

	// Graphics settings help buttons
	reduceTerrainDistanceHelpButton *widget.Button
	setMultisampleTo2xHelpButton    *widget.Button
	setShadowLOD0HelpButton         *widget.Button

	// Window reference for popup functionality
	currentWindow fyne.Window

	// Logo components
	logoImage     *canvas.Image
	logoContainer fyne.CanvasObject

	// State variables
	currentWineRegistryEnabled bool
	remapOperationInProgress   bool

	// Pulsing effect variables (pulsingActive is in status.go)
	pulsingTicker *time.Ticker

	// Troubleshooting popup and controls
	troubleshootingButton       *widget.Button
	troubleshootingPopupActive  bool
	crossoverVersionStatusLabel *widget.RichText
	wdbDeleteButton             *widget.Button
	wineDeleteButton            *widget.Button
	appMgmtPermissionButton     *widget.Button
	troubleshootingCloseButton  *widget.Button
)
