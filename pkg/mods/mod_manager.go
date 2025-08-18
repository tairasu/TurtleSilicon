package mods

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Mod struct {
	Name        string
	Path        string
	Enabled     bool
	Required    bool // For winerosetta.dll which is required
	FileSize    int64
	LastMod     time.Time
	Description string
}

type ModManager struct {
	mods             []Mod
	window           fyne.Window
	currentPopup     *widget.PopUp
	modsList         *container.Scroll
	contentContainer *fyne.Container
	versionManager   *version.VersionManager
}

// getModDescription returns a description for known mods
func getModDescription(modName string) string {
	switch strings.ToLower(modName) {
	case "libsiliconpatch.dll":
		return "Hooks into the WoW process and replaces slow X87 instructions with SSE2 instructions that Rosetta can translate much quicker, resulting in an increase in FPS (2x or more). This mod is enabled by default for new users as it provides significant performance improvements. May potentially cause rare graphical bugs in some situations."
	case "winerosetta.dll":
		return "Core Wine compatibility layer required for running 32-bit World of Warcraft executables on Apple Silicon. This mod is required and cannot be disabled."
	case "d3d9.dll":
		return "Direct3D 9 graphics wrapper that provides compatibility and performance optimizations for DirectX applications."
	default:
		return "No description available for this mod."
	}
}

func NewModManager(window fyne.Window, vm *version.VersionManager) *ModManager {
	return &ModManager{
		window:         window,
		versionManager: vm,
	}
}

// IsModsSupported checks if the current version supports mods
func (mm *ModManager) IsModsSupported() bool {
	if mm.versionManager == nil {
		debug.Printf("IsModsSupported: versionManager is nil")
		return false
	}

	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err != nil || currentVer == nil {
		debug.Printf("IsModsSupported: failed to get current version: %v", err)
		return false
	}

	debug.Printf("IsModsSupported: Current version ID: %s, DisplayName: %s, SupportsDLLLoading: %v",
		currentVer.ID, currentVer.DisplayName, currentVer.SupportsDLLLoading)

	// Use the SupportsDLLLoading field from the version configuration
	return currentVer.SupportsDLLLoading
}

// ensureModsDirectory creates the mods directory if it doesn't exist
func (mm *ModManager) ensureModsDirectory() error {
	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err != nil || currentVer == nil || currentVer.GamePath == "" {
		return fmt.Errorf("game path not set")
	}

	modsPath := filepath.Join(currentVer.GamePath, "mods")
	if _, err := os.Stat(modsPath); os.IsNotExist(err) {
		debug.Printf("Creating mods directory: %s", modsPath)
		if err := os.MkdirAll(modsPath, 0755); err != nil {
			return fmt.Errorf("failed to create mods directory: %v", err)
		}
	}
	return nil
}

// getDllsFilePath returns the path to dlls.txt for the current game
func (mm *ModManager) getDllsFilePath() string {
	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err != nil || currentVer == nil {
		return ""
	}
	return filepath.Join(currentVer.GamePath, "dlls.txt")
}

// ScanMods scans for mods in the mods directory and parses dlls.txt
func (mm *ModManager) ScanMods() error {
	return mm.ScanModsWithProgress(nil)
}

func (mm *ModManager) ScanModsWithProgress(updateProgress func(string)) error {
	if !mm.IsModsSupported() {
		return fmt.Errorf("mods are not supported for this game version")
	}

	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err != nil || currentVer == nil || currentVer.GamePath == "" {
		return fmt.Errorf("game path not set")
	}

	if updateProgress != nil {
		updateProgress("Ensuring mods directory exists...")
	}

	// Ensure mods directory exists
	if err := mm.ensureModsDirectory(); err != nil {
		return err
	}

	if updateProgress != nil {
		updateProgress("Reading dlls.txt...")
	}

	// Parse dlls.txt to see which mods are enabled
	enabledMods, err := mm.parseDllsFile()
	if err != nil {
		debug.Printf("Warning: failed to parse dlls.txt: %v", err)
		enabledMods = make(map[string]bool)
	}

	if updateProgress != nil {
		updateProgress("Scanning mods directory...")
	}

	// Scan mods directory for DLL files
	modsPath := filepath.Join(currentVer.GamePath, "mods")
	entries, err := os.ReadDir(modsPath)
	if err != nil {
		return fmt.Errorf("failed to read mods directory: %v", err)
	}

	mm.mods = []Mod{}

	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .dll files
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".dll") {
			continue
		}

		if updateProgress != nil {
			updateProgress(fmt.Sprintf("Processing %s (%d/%d)...", entry.Name(), i+1, len(entries)))
		}

		modPath := filepath.Join(modsPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			debug.Printf("Warning: failed to get info for %s: %v", entry.Name(), err)
			continue
		}

		mod := Mod{
			Name:        entry.Name(),
			Path:        modPath,
			Enabled:     enabledMods["mods/"+entry.Name()],
			Required:    entry.Name() == "winerosetta.dll",
			FileSize:    info.Size(),
			LastMod:     info.ModTime(),
			Description: getModDescription(entry.Name()),
		}

		mm.mods = append(mm.mods, mod)
	}

	if updateProgress != nil {
		updateProgress("Finalizing...")
	}

	debug.Printf("Found %d mods (%d enabled)", len(mm.mods), mm.countEnabledMods())
	return nil
}

// parseDllsFile parses the dlls.txt file and returns enabled mods
func (mm *ModManager) parseDllsFile() (map[string]bool, error) {
	dllsPath := mm.getDllsFilePath()
	enabledMods := make(map[string]bool)

	file, err := os.Open(dllsPath)
	if err != nil {
		// If file doesn't exist, that's okay - no mods are enabled
		if os.IsNotExist(err) {
			return enabledMods, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if line is commented out (disabled)
		if strings.HasPrefix(line, "#") {
			// This mod is disabled
			dllName := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			enabledMods[dllName] = false
		} else {
			// This mod is enabled
			enabledMods[line] = true
		}
	}

	return enabledMods, scanner.Err()
}

// updateDllsFile updates the dlls.txt file with current mod states
func (mm *ModManager) updateDllsFile() error {
	dllsPath := mm.getDllsFilePath()

	// Read existing file content to preserve non-mod entries
	var existingLines []string
	if file, err := os.Open(dllsPath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Skip lines that are mod-related (start with "mods/" or "#mods/")
			cleanLine := strings.TrimPrefix(line, "#")
			if !strings.HasPrefix(cleanLine, "mods/") && line != "" {
				existingLines = append(existingLines, line)
			}
		}
		file.Close()
	}

	// Create new file content
	var lines []string
	lines = append(lines, existingLines...)

	// Add mod entries
	for _, mod := range mm.mods {
		modEntry := "mods/" + mod.Name
		if mod.Enabled {
			lines = append(lines, modEntry)
		} else {
			lines = append(lines, "#"+modEntry)
		}
	}

	// Write file
	file, err := os.Create(dllsPath)
	if err != nil {
		return fmt.Errorf("failed to create dlls.txt: %v", err)
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write dlls.txt: %v", err)
		}
	}

	debug.Printf("Updated dlls.txt with %d mod entries", len(mm.mods))
	return nil
}

func (mm *ModManager) countEnabledMods() int {
	count := 0
	for _, mod := range mm.mods {
		if mod.Enabled {
			count++
		}
	}
	return count
}

// ToggleMod toggles the enabled state of a mod
func (mm *ModManager) ToggleMod(modIndex int) error {
	if modIndex < 0 || modIndex >= len(mm.mods) {
		return fmt.Errorf("invalid mod index")
	}

	mod := &mm.mods[modIndex]

	// Don't allow disabling required mods
	if mod.Required && mod.Enabled {
		return fmt.Errorf("cannot disable required mod: %s", mod.Name)
	}

	mod.Enabled = !mod.Enabled
	debug.Printf("Toggled mod %s: enabled=%v", mod.Name, mod.Enabled)

	// Special handling for libSiliconPatch.dll - update version settings
	if strings.ToLower(mod.Name) == "libsiliconpatch.dll" {
		if currentVer, err := mm.versionManager.GetCurrentVersion(); err == nil {
			if mod.Enabled {
				// User enabled libSiliconPatch - clear the disabled flag
				currentVer.Settings.EnableLibSiliconPatch = true
				currentVer.Settings.UserDisabledLibSiliconPatch = false
				debug.Printf("libSiliconPatch enabled via mods manager - clearing user disabled flag")
			} else {
				// User disabled libSiliconPatch - set the disabled flag
				currentVer.Settings.EnableLibSiliconPatch = false
				currentVer.Settings.UserDisabledLibSiliconPatch = true
				debug.Printf("libSiliconPatch disabled via mods manager - setting user disabled flag")
			}
			mm.versionManager.UpdateVersion(currentVer)
		}
	}

	// Update dlls.txt file
	return mm.updateDllsFile()
}

// DeleteMod removes a mod file and updates dlls.txt
func (mm *ModManager) DeleteMod(modIndex int) error {
	if modIndex < 0 || modIndex >= len(mm.mods) {
		return fmt.Errorf("invalid mod index")
	}

	mod := &mm.mods[modIndex]

	// Don't allow deleting required mods
	if mod.Required {
		return fmt.Errorf("cannot delete required mod: %s", mod.Name)
	}

	debug.Printf("Deleting mod: %s", mod.Name)

	// Remove the file
	if err := os.Remove(mod.Path); err != nil {
		return fmt.Errorf("failed to delete mod file: %v", err)
	}

	// Remove from our list
	mm.mods = append(mm.mods[:modIndex], mm.mods[modIndex+1:]...)

	// Update dlls.txt
	return mm.updateDllsFile()
}

// ShowModManager shows the mod manager popup
func (mm *ModManager) ShowModManager() {
	debug.Printf("ShowModManager called")

	// Get version info for debugging
	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err != nil {
		debug.Printf("ERROR: Failed to get current version: %v", err)
		dialog.ShowError(fmt.Errorf("Failed to get current version: %v", err), mm.window)
		return
	}

	debug.Printf("Current version details: ID='%s', DisplayName='%s', SupportsDLLLoading=%v",
		currentVer.ID, currentVer.DisplayName, currentVer.SupportsDLLLoading)

	if !mm.IsModsSupported() {
		debug.Printf("Mods not supported for current version")
		// Show detailed error message with current version info for debugging
		errorMsg := fmt.Sprintf("Mods are not supported for the current version.\n\n"+
			"Current Version: %s (ID: %s)\n"+
			"SupportsDLLLoading: %v\n\n"+
			"Mods are only supported for TurtleWoW, EpochSilicon, and WrathSilicon.\n"+
			"VanillaSilicon and BurningSilicon use the DivX patch method which does not support DLL injection.",
			currentVer.DisplayName, currentVer.ID, currentVer.SupportsDLLLoading)

		dialog.ShowInformation("Mods Not Supported", errorMsg, mm.window)
		return
	}

	if mm.currentPopup != nil {
		mm.currentPopup.Hide()
		mm.currentPopup = nil
	}

	loadingText := widget.NewLabel("Initializing...")
	loadingText.Alignment = fyne.TextAlignCenter

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	loadingContent := container.NewVBox(
		widget.NewLabel("Loading Mods"),
		widget.NewSeparator(),
		loadingText,
		progressBar,
	)

	loadingPopup := widget.NewModalPopUp(container.NewPadded(loadingContent), mm.window.Canvas())
	loadingPopup.Resize(fyne.NewSize(300, 150))
	loadingPopup.Show()

	go func() {
		updateProgress := func(message string) {
			fyne.Do(func() {
				loadingText.SetText(message)
			})
		}

		if err := mm.ScanModsWithProgress(updateProgress); err != nil {
			debug.Printf("Error scanning mods: %v", err)
			fyne.Do(func() {
				progressBar.Stop()
				loadingPopup.Hide()
				dialog.ShowError(fmt.Errorf("failed to scan mods: %v", err), mm.window)
			})
			return
		}

		debug.Printf("Successfully scanned mods")

		fyne.Do(func() {
			progressBar.Stop()
			loadingPopup.Hide()
			mm.createModManagerPopup()
		})
	}()
}

func (mm *ModManager) createModManagerPopup() {
	titleText := widget.NewLabel("Mod Manager")
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	summaryText := widget.NewLabel(fmt.Sprintf("Found %d mods (%d enabled)", len(mm.mods), mm.countEnabledMods()))
	summaryText.TextStyle = fyne.TextStyle{Italic: true}

	refreshButton := widget.NewButton("Refresh", func() {
		mm.refreshModManager()
	})
	refreshButton.Importance = widget.MediumImportance

	addButton := widget.NewButton("Add Mod", func() {
		mm.showAddModDialog()
	})
	addButton.Importance = widget.HighImportance

	leftSide := container.NewHBox(summaryText)
	rightSide := container.NewHBox(addButton, refreshButton)

	headerContainer := container.NewBorder(
		nil,
		widget.NewSeparator(),
		leftSide,
		rightSide,
		nil,
	)

	mm.modsList = mm.createModsList()

	mm.contentContainer = container.NewBorder(
		headerContainer,
		nil,
		nil,
		nil,
		mm.modsList,
	)

	closeButton := widget.NewButton("✕", func() {})
	closeButton.Importance = widget.LowImportance
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8))
	closeButton.Resize(fyne.NewSize(30, 30))

	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		container.NewCenter(titleText),
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		container.NewPadded(mm.contentContainer),
	)

	popup := widget.NewModalPopUp(mainContainer, mm.window.Canvas())

	canvasSize := mm.window.Canvas().Size()
	popup.Resize(canvasSize)

	mm.currentPopup = popup

	canvas := mm.window.Canvas()
	originalOnTypedKey := canvas.OnTypedKey()

	closeAction := func() {
		canvas.SetOnTypedKey(originalOnTypedKey)
		popup.Hide()
		mm.currentPopup = nil
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

func (mm *ModManager) createModsList() *container.Scroll {
	list := container.NewVBox()

	for i := range mm.mods {
		modCopy := &mm.mods[i]
		modIndex := i
		modCard := mm.createModCard(modCopy, modIndex)
		list.Add(modCard)
		if i < len(mm.mods)-1 {
			list.Add(widget.NewSeparator())
		}
	}

	if len(mm.mods) == 0 {
		emptyLabel := widget.NewLabel("No mods found in mods directory")
		emptyLabel.Alignment = fyne.TextAlignCenter
		list.Add(emptyLabel)
	}

	return container.NewScroll(list)
}

func (mm *ModManager) createModCard(mod *Mod, modIndex int) *fyne.Container {
	nameText := fmt.Sprintf("**%s**", mod.Name)
	nameLabel := widget.NewRichTextFromMarkdown(nameText)
	nameLabel.Wrapping = fyne.TextWrapOff

	// Show file size and last modified
	sizeText := fmt.Sprintf("Size: %.1f KB", float64(mod.FileSize)/1024)
	lastModText := fmt.Sprintf("Modified: %s", mod.LastMod.Format("2006-01-02 15:04"))
	infoLabel := widget.NewLabel(fmt.Sprintf("%s | %s", sizeText, lastModText))
	infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	nameContainer := container.NewHBox(nameLabel)
	textGroup := container.NewWithoutLayout(nameContainer, infoLabel)
	nameContainer.Move(fyne.NewPos(0, 0))
	nameContainer.Resize(fyne.NewSize(400, 18))
	infoLabel.Move(fyne.NewPos(0, 16))
	infoLabel.Resize(fyne.NewSize(400, 14))
	textGroup.Resize(fyne.NewSize(400, 30))

	infoContainer := container.NewWithoutLayout(textGroup)
	textGroup.Move(fyne.NewPos(0, -5))
	infoContainer.Resize(fyne.NewSize(400, 50))

	// Enable/Disable checkbox
	enabledCheck := widget.NewCheck("Enabled", nil)
	enabledCheck.SetChecked(mod.Enabled)

	// Gray out and disable checkbox for required mods
	if mod.Required {
		enabledCheck.Disable()
		enabledCheck.SetChecked(true)
	} else {
		// Only add callback for non-required mods
		enabledCheck.OnChanged = func(checked bool) {
			debug.Printf("Checkbox changed for mod %s: %v", mod.Name, checked)
			if err := mm.ToggleMod(modIndex); err != nil {
				debug.Printf("Error toggling mod: %v", err)
				// Revert checkbox state on error
				enabledCheck.SetChecked(!checked)
				dialog.ShowError(err, mm.window)
			}
		}
	}

	// Required indicator
	var requiredLabel *widget.Label
	if mod.Required {
		requiredLabel = widget.NewLabel("(Required)")
		requiredLabel.TextStyle = fyne.TextStyle{Italic: true}
	}

	// Info button
	infoButton := widget.NewButton("Info", func() {
		mm.showModInfoPopup(mod.Name, mod.Description)
	})
	infoButton.Importance = widget.LowImportance

	deleteButton := widget.NewButton("Delete", func() {
		mm.confirmDeleteMod(modIndex)
	})
	deleteButton.Importance = widget.MediumImportance

	// Don't allow deleting required mods
	if mod.Required {
		deleteButton.Disable()
	}

	var buttonsContainer *fyne.Container
	if requiredLabel != nil {
		buttonsContainer = container.NewHBox(enabledCheck, requiredLabel, infoButton, deleteButton)
	} else {
		buttonsContainer = container.NewHBox(enabledCheck, infoButton, deleteButton)
	}

	buttonsWithMargin := container.NewPadded(buttonsContainer)

	cardContainer := container.NewBorder(
		nil,
		nil,
		infoContainer,
		buttonsWithMargin,
		nil,
	)

	cardContainer.Resize(fyne.NewSize(600, 70))

	return container.NewPadded(cardContainer)
}

func (mm *ModManager) refreshModsList() {
	if mm.currentPopup == nil || mm.modsList == nil {
		return
	}

	newModsList := mm.createModsList()
	mm.modsList.Content = newModsList.Content
	mm.modsList.Refresh()
}

func (mm *ModManager) refreshModManager() {
	if mm.currentPopup == nil {
		// No popup is currently open, so just show a new one
		mm.ShowModManager()
		return
	}

	// Update the existing popup content
	go func() {
		// Re-scan mods
		if err := mm.ScanMods(); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to refresh mods: %v", err), mm.window)
			})
			return
		}

		// Update the popup content on main thread
		fyne.Do(func() {
			// Get current popup size and position
			popupSize := mm.currentPopup.Size()

			// Hide current popup
			mm.currentPopup.Hide()

			// Create new popup with updated content
			mm.createModManagerPopup()

			// Maintain the same size
			if mm.currentPopup != nil {
				mm.currentPopup.Resize(popupSize)
			}
		})
	}()
}

func (mm *ModManager) confirmDeleteMod(modIndex int) {
	if modIndex < 0 || modIndex >= len(mm.mods) {
		return
	}

	mod := &mm.mods[modIndex]
	message := fmt.Sprintf("Are you sure you want to delete the mod '%s'?\n\nThis action cannot be undone.", mod.Name)

	confirmDialog := dialog.NewConfirm("Confirm Deletion", message, func(confirmed bool) {
		if confirmed {
			if err := mm.DeleteMod(modIndex); err != nil {
				dialog.ShowError(err, mm.window)
			} else {
				dialog.ShowInformation("Success", fmt.Sprintf("Successfully deleted %s", mod.Name), mm.window)
				mm.refreshModManager()
			}
		}
	}, mm.window)

	confirmDialog.Show()
}

func (mm *ModManager) showModInfoPopup(modName, description string) {
	titleText := widget.NewRichTextFromMarkdown(fmt.Sprintf("# %s", modName))
	titleText.Wrapping = fyne.TextWrapOff

	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	contentContainer := container.NewVBox(
		container.NewCenter(titleText),
		widget.NewSeparator(),
		descriptionLabel,
	)

	windowSize := mm.window.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 2

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge

	// Create top bar with close button
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		nil,
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		contentContainer,
	)

	popup := widget.NewModalPopUp(mainContainer, mm.window.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}

func (mm *ModManager) showAddModDialog() {
	// Create content for the dialog
	titleLabel := widget.NewLabel("Add Mods")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	instructionText := widget.NewLabel("To add mods:\n\n" +
		"1. Place your .dll files in the 'mods' directory inside your game folder\n" +
		"2. Click 'Refresh' to reload the mod list\n\n" +
		"The mods directory will be created automatically if it doesn't exist.\n" +
		"Note: d3d9.dll should remain in the root game directory, not in mods/")
	instructionText.Wrapping = fyne.TextWrapWord

	okButton := widget.NewButton("OK", func() {
		// This will be set when the popup is created
	})
	okButton.Importance = widget.HighImportance

	// Only show "Where do I get mods?" button for turtlesilicon
	var buttonsContainer *fyne.Container
	currentVer, err := mm.versionManager.GetCurrentVersion()
	if err == nil && currentVer.ID == "turtlesilicon" {
		whereToGetButton := widget.NewButton("Where do I get mods?", func() {
			modURL := "https://turtle-wow.fandom.com/wiki/Client_Fixes_and_Tweaks"
			// Open URL in browser
			cmd := exec.Command("open", modURL)
			if err := cmd.Start(); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to open mods URL: %v", err), mm.window)
			}
		})
		whereToGetButton.Importance = widget.MediumImportance
		buttonsContainer = container.NewHBox(whereToGetButton, okButton)
	} else {
		buttonsContainer = container.NewHBox(okButton)
	}

	content := container.NewVBox(
		container.NewCenter(titleLabel),
		widget.NewSeparator(),
		instructionText,
		widget.NewSeparator(),
		container.NewCenter(buttonsContainer),
	)

	// Calculate popup size
	windowSize := mm.window.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 2

	popup := widget.NewModalPopUp(container.NewPadded(content), mm.window.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	okButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}
