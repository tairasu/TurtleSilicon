package ui

import (
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

	// Environment variables entry
	envVarsEntry *widget.Entry

	// Window reference for popup functionality
	currentWindow fyne.Window
)
