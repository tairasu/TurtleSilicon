package ui

import (
	"time"

	"fyne.io/fyne/v2"
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

	// Option checkboxes
	metalHudCheckbox      *widget.Check
	showTerminalCheckbox  *widget.Check
	vanillaTweaksCheckbox *widget.Check

	// Recommended settings button
	applyRecommendedSettingsButton *widget.Button
	recommendedSettingsHelpButton  *widget.Button

	// Wine registry buttons and status
	enableOptionAsAltButton  *widget.Button
	disableOptionAsAltButton *widget.Button
	optionAsAltStatusLabel   *widget.RichText

	// Environment variables entry
	envVarsEntry *widget.Entry

	// Window reference for popup functionality
	currentWindow fyne.Window

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
