package main

import (
	"bytes" // Added import
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const defaultCrossOverPath = "/Applications/CrossOver.app"

var (
	crossoverPath           string
	turtlewowPath           string
	patchesAppliedTurtleWoW = false
	patchesAppliedCrossOver = false
	enableMetalHud          = true // Default to enabled
)

// Helper function to check if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper function to check if a path exists and is a directory
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// Helper function to copy a directory recursively
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	dir, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range dir {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Helper function to run an AppleScript command using osascript
func runOsascript(scriptString string, myWindow fyne.Window) bool {
	log.Printf("Executing AppleScript: %s", scriptString)
	cmd := exec.Command("osascript", "-e", scriptString)
	output, err := cmd.CombinedOutput() // Changed variable name to avoid conflict if 'output' is used later
	if err != nil {
		errMsg := fmt.Sprintf("AppleScript failed: %v\\nOutput: %s", err, string(output))
		dialog.ShowError(fmt.Errorf(errMsg), myWindow)
		log.Println(errMsg)
		return false
	}
	log.Printf("osascript output: %s", string(output))
	return true
}

// escapeStringForAppleScript escapes a string to be safely embedded in an AppleScript double-quoted string.
func escapeStringForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\") // Escape backslashes first: \ -> \\
	s = strings.ReplaceAll(s, "\"", "\\\"")  // Escape double quotes: " -> \"
	return s
}

func main() {
	myApp := app.NewWithID("com.example.turtlesilicon")
	myWindow := myApp.NewWindow("TurtleSilicon")
	myWindow.Resize(fyne.NewSize(650, 450)) // Slightly wider for clarity
	myWindow.SetFixedSize(true)

	// --- Path Labels ---
	crossoverPathLabel := widget.NewRichText() // Changed to RichText
	turtlewowPathLabel := widget.NewRichText() // Changed to RichText

	// --- Status Labels (Changed to RichText for color) ---
	var turtlewowStatusLabel *widget.RichText
	var crossoverStatusLabel *widget.RichText

	turtlewowStatusLabel = widget.NewRichText()
	crossoverStatusLabel = widget.NewRichText()

	// --- Checkbox for Metal HUD ---
	metalHudCheckbox := widget.NewCheck("Enable Metal Hud (show FPS)", func(checked bool) {
		enableMetalHud = checked
		log.Printf("Metal HUD enabled: %v", enableMetalHud)
	})
	metalHudCheckbox.SetChecked(enableMetalHud) // Set initial state

	// --- Buttons (declared here to be accessible in updateAllStatuses) ---
	var launchButton *widget.Button
	var patchTurtleWoWButton *widget.Button
	var patchCrossOverButton *widget.Button

	// --- Helper to update all statuses and button states ---
	updateAllStatuses := func() {
		// Update Crossover Path and Status
		if crossoverPath == "" {
			crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			crossoverPathLabel.Refresh()
			patchesAppliedCrossOver = false // Reset if path is cleared
		} else {
			crossoverPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: crossoverPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
			crossoverPathLabel.Refresh()
			wineloader2Path := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
			if pathExists(wineloader2Path) {
				patchesAppliedCrossOver = true
			} else {
				// patchesAppliedCrossOver = false // Only set to false if not already true from a patch action this session
			}
		}
		if patchesAppliedCrossOver {
			crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}} // Changed to ColorNameSuccess
			crossoverStatusLabel.Refresh()
			if patchCrossOverButton != nil {
				patchCrossOverButton.Disable()
			}
		} else {
			crossoverStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			crossoverStatusLabel.Refresh()
			if patchCrossOverButton != nil {
				if crossoverPath != "" {
					patchCrossOverButton.Enable()
				} else {
					patchCrossOverButton.Disable()
				}
			}
		}

		// Update TurtleWoW Path and Status
		if turtlewowPath == "" {
			turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not set", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			turtlewowPathLabel.Refresh()
			patchesAppliedTurtleWoW = false // Reset if path is cleared
		} else {
			turtlewowPathLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: turtlewowPath, Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}}
			turtlewowPathLabel.Refresh()
			winerosettaDllPath := filepath.Join(turtlewowPath, "winerosetta.dll")
			d3d9DllPath := filepath.Join(turtlewowPath, "d3d9.dll")
			rosettaX87DirPath := filepath.Join(turtlewowPath, "rosettax87")
			// Check for libRuntimeRosettax87 as well
			rosettaX87ExePath := filepath.Join(rosettaX87DirPath, "rosettax87")
			libRuntimeRosettaX87Path := filepath.Join(rosettaX87DirPath, "libRuntimeRosettax87")

			if pathExists(winerosettaDllPath) && pathExists(d3d9DllPath) && dirExists(rosettaX87DirPath) && pathExists(rosettaX87ExePath) && pathExists(libRuntimeRosettaX87Path) {
				patchesAppliedTurtleWoW = true
			} else {
				// patchesAppliedTurtleWoW = false
			}
		}
		if patchesAppliedTurtleWoW {
			turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameSuccess}}} // Changed to ColorNameSuccess
			turtlewowStatusLabel.Refresh()
			if patchTurtleWoWButton != nil {
				patchTurtleWoWButton.Disable()
			}
		} else {
			turtlewowStatusLabel.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: "Not patched", Style: widget.RichTextStyle{ColorName: theme.ColorNameError}}}
			turtlewowStatusLabel.Refresh()
			if patchTurtleWoWButton != nil {
				if turtlewowPath != "" {
					patchTurtleWoWButton.Enable()
				} else {
					patchTurtleWoWButton.Disable()
				}
			}
		}

		// Update Launch Button State
		if launchButton != nil {
			if patchesAppliedTurtleWoW && patchesAppliedCrossOver && turtlewowPath != "" && crossoverPath != "" {
				launchButton.Enable()
			} else {
				launchButton.Disable()
			}
		}
	}

	// --- Path Selection Functions ---
	selectCrossOverPath := func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if uri == nil { // User cancelled
				log.Println("CrossOver path selection cancelled.")
				// Do not reset crossoverPath if user cancels, keep previous valid path
				updateAllStatuses() // Re-evaluate with existing path
				return
			}
			selectedPath := uri.Path()
			if filepath.Ext(selectedPath) == ".app" && dirExists(selectedPath) { // Check if it's a directory too
				crossoverPath = selectedPath
				patchesAppliedCrossOver = false // Reset patch status on new path, updateAllStatuses will re-check
				log.Println("CrossOver path set to:", crossoverPath)
			} else {
				// Don't reset crossoverPath, show error and keep old one if any
				dialog.ShowError(fmt.Errorf("invalid selection: '%s'. Please select a valid .app bundle", selectedPath), myWindow)
				log.Println("Invalid CrossOver path selected:", selectedPath)
			}
			updateAllStatuses()
		}, myWindow)
	}

	selectTurtleWoWPath := func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if uri == nil { // User cancelled
				log.Println("TurtleWoW path selection cancelled.")
				updateAllStatuses() // Re-evaluate
				return
			}
			selectedPath := uri.Path()
			if dirExists(selectedPath) { // Basic check for directory
				turtlewowPath = selectedPath
				patchesAppliedTurtleWoW = false // Reset patch status on new path, updateAllStatuses will re-check
				log.Println("TurtleWoW path set to:", turtlewowPath)
			} else {
				dialog.ShowError(fmt.Errorf("invalid selection: '%s' is not a valid directory", selectedPath), myWindow)
				log.Println("Invalid TurtleWoW path selected:", selectedPath)
			}
			updateAllStatuses()
		}, myWindow)
	}

	// --- Patching Functions ---
	patchTurtleWoWFunc := func() {
		log.Println("Patch TurtleWoW clicked")
		if turtlewowPath == "" {
			dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it first."), myWindow)
			return
		}

		// Target paths
		targetWinerosettaDll := filepath.Join(turtlewowPath, "winerosetta.dll")
		targetD3d9Dll := filepath.Join(turtlewowPath, "d3d9.dll")
		targetRosettaX87Dir := filepath.Join(turtlewowPath, "rosettax87")

		// Files to copy directly into turtlewowPath
		filesToCopy := map[string]string{
			"winerosetta/winerosetta.dll": targetWinerosettaDll, // Adjusted path
			"winerosetta/d3d9.dll":        targetD3d9Dll,        // Adjusted path
		}

		for resourceName, destPath := range filesToCopy {
			log.Printf("Processing resource: %s to %s", resourceName, destPath)

			resource, err := fyne.LoadResourceFromPath(resourceName)
			if err != nil {
				errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}

			destinationFile, err := os.Create(destPath)
			if err != nil {
				errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
			defer destinationFile.Close()

			_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content())) // Changed to use bytes.NewReader(resource.Content())
			if err != nil {
				errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
			log.Printf("Successfully copied %s to %s", resourceName, destPath)
		}

		// Handle rosettax87 folder and its contents
		log.Printf("Preparing rosettax87 directory at: %s", targetRosettaX87Dir)
		if err := os.RemoveAll(targetRosettaX87Dir); err != nil {
			log.Printf("Warning: could not remove existing rosettax87 folder '%s': %v", targetRosettaX87Dir, err)
			// Not necessarily fatal, MkdirAll will handle creation.
		}
		if err := os.MkdirAll(targetRosettaX87Dir, 0755); err != nil {
			errMsg := fmt.Sprintf("failed to create directory %s: %v", targetRosettaX87Dir, err)
			dialog.ShowError(fmt.Errorf(errMsg), myWindow)
			log.Println(errMsg)
			patchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		rosettaFilesToCopy := map[string]string{
			"rosettax87/rosettax87":           filepath.Join(targetRosettaX87Dir, "rosettax87"),           // Adjusted path
			"rosettax87/libRuntimeRosettax87": filepath.Join(targetRosettaX87Dir, "libRuntimeRosettax87"), // Added libRuntimeRosettax87
		}

		for resourceName, destPath := range rosettaFilesToCopy {
			log.Printf("Processing rosetta resource: %s to %s", resourceName, destPath)
			resource, err := fyne.LoadResourceFromPath(resourceName)
			if err != nil {
				errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}

			destinationFile, err := os.Create(destPath)
			if err != nil {
				errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
			// No defer destinationFile.Close() here, because we chmod after copy and then it can be closed.

			_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content())) // Changed to use bytes.NewReader(resource.Content())
			if err != nil {
				destinationFile.Close() // Close before erroring out
				errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
				dialog.ShowError(fmt.Errorf(errMsg), myWindow)
				log.Println(errMsg)
				patchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
			destinationFile.Close() // Close after successful copy

			// Set execute permissions for the rosettax87 executable
			if filepath.Base(destPath) == "rosettax87" { // Corrected condition
				log.Printf("Setting execute permission for %s", destPath)
				if err := os.Chmod(destPath, 0755); err != nil {
					errMsg := fmt.Sprintf("failed to set execute permission for %s: %v", destPath, err)
					dialog.ShowError(fmt.Errorf(errMsg), myWindow)
					log.Println(errMsg)
					// Decide if this is fatal or a warning. For now, treat as fatal for patching.
					patchesAppliedTurtleWoW = false
					updateAllStatuses()
					return
				}
			}
			log.Printf("Successfully copied %s to %s", resourceName, destPath)
		}

		log.Println("TurtleWoW patching with bundled resources completed successfully.")
		patchesAppliedTurtleWoW = true
		dialog.ShowInformation("Success", "TurtleWoW patching process completed using bundled resources.", myWindow)
		updateAllStatuses()
	}

	patchCrossOverFunc := func() {
		log.Println("Patch CrossOver clicked")
		if crossoverPath == "" {
			dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it first."), myWindow)
			return
		}

		wineloaderBasePath := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application")
		wineloaderOrig := filepath.Join(wineloaderBasePath, "wineloader")
		wineloaderCopy := filepath.Join(wineloaderBasePath, "wineloader2")

		if !pathExists(wineloaderOrig) {
			dialog.ShowError(fmt.Errorf("original wineloader not found at %s", wineloaderOrig), myWindow)
			patchesAppliedCrossOver = false
			updateAllStatuses()
			return
		}

		// 1. Make a copy of wineloader
		log.Printf("Copying %s to %s", wineloaderOrig, wineloaderCopy)
		if err := copyFile(wineloaderOrig, wineloaderCopy); err != nil {
			dialog.ShowError(fmt.Errorf("failed to copy wineloader: %w", err), myWindow)
			patchesAppliedCrossOver = false
			updateAllStatuses()
			return
		}

		// 2. Execute codesign --remove-signature
		log.Printf("Executing: codesign --remove-signature %s", wineloaderCopy)
		cmd := exec.Command("codesign", "--remove-signature", wineloaderCopy)
		combinedOutput, err := cmd.CombinedOutput() // Store result in a new variable
		if err != nil {
			// Corrected variable name for error message and use combinedOutput
			derrMsg := fmt.Sprintf("failed to remove signature from %s: %v\\nOutput: %s", wineloaderCopy, err, string(combinedOutput))
			dialog.ShowError(fmt.Errorf(derrMsg), myWindow)
			log.Println(derrMsg)
			patchesAppliedCrossOver = false
			// Attempt to clean up the copied file if codesign fails
			if err := os.Remove(wineloaderCopy); err != nil {
				log.Printf("Warning: failed to cleanup wineloader2 after codesign failure: %v", err)
			}
			updateAllStatuses()
			return
		}
		log.Printf("codesign output: %s", string(combinedOutput)) // Use combinedOutput here

		log.Println("CrossOver patching completed successfully.")
		patchesAppliedCrossOver = true
		dialog.ShowInformation("Success", "CrossOver patching process completed.", myWindow)
		updateAllStatuses()
	}

	// --- Launch Function ---
	launchGameFunc := func(myWindow fyne.Window) {
		log.Println("Launch Game button clicked")

		// crossoverBinPath := filepath.Join(crossoverPath, "Contents", "MacOS", "CrossOver") // No longer used
		turtlewowExePath := filepath.Join(turtlewowPath, "WoW.exe")

		// Pre-launch checks
		if crossoverPath == "" {
			dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it in the patcher."), myWindow)
			return
		}
		if turtlewowPath == "" {
			dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it in the patcher."), myWindow)
			return
		}
		if !patchesAppliedTurtleWoW || !patchesAppliedCrossOver {
			confirmed := false
			dialog.ShowConfirm("Warning", "Not all patches confirmed applied. Continue with launch?", func(c bool) {
				confirmed = c
			}, myWindow)
			if !confirmed {
				return
			}
		}

		log.Println("Preparing to launch TurtleSilicon...")

		rosettaInTurtlePath := filepath.Join(turtlewowPath, "rosettax87")
		rosettaExecutable := filepath.Join(rosettaInTurtlePath, "rosettax87")
		wineloader2Path := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
		wowExePath := filepath.Join(turtlewowPath, "wow.exe")

		if !pathExists(rosettaExecutable) {
			dialog.ShowError(fmt.Errorf("rosetta executable not found at %s. Ensure TurtleWoW patching was successful", rosettaExecutable), myWindow)
			return
		}
		if !pathExists(wineloader2Path) {
			dialog.ShowError(fmt.Errorf("patched wineloader2 not found at %s. Ensure CrossOver patching was successful", wineloader2Path), myWindow)
			return
		}
		if !pathExists(wowExePath) {
			dialog.ShowError(fmt.Errorf("wow.exe not found at %s. Ensure your TurtleWoW directory is correct", wowExePath), myWindow)
			return
		}

		// Command 1: Launch rosettax87
		// The path itself needs to be an AppleScript string literal, so we use escapeStringForAppleScript.
		// Corrected: cd into the rosettax87 directory itself.
		appleScriptSafeRosettaDir := escapeStringForAppleScript(rosettaInTurtlePath)
		// Construct the AppleScript command using a regular Go string literal for the format string.
		// This ensures that \" correctly escapes double quotes for Go, producing literal quotes in the AppleScript command.
		cmd1Script := fmt.Sprintf("tell application \"Terminal\" to do script \"cd \" & quoted form of \"%s\" & \" && sudo ./rosettax87\"", appleScriptSafeRosettaDir)

		log.Println("Launching rosettax87 (requires sudo password in new terminal)...")
		if !runOsascript(cmd1Script, myWindow) {
			// Error already shown by runOsascript
			return
		}

		dialog.ShowConfirm("Action Required", // Changed from ShowInformation to ShowConfirm
			"The rosetta x87 terminal has been initiated.\n\n"+
				"1. Please enter your sudo password in that new terminal window.\n"+
				"2. Wait for rosetta x87 to fully start.\n\n"+
				"Click Yes once rosetta x87 is running and you have entered the password.\n"+
				"Click No to abort launching WoW.",
			func(confirmed bool) {
				if confirmed {
					log.Println("User confirmed rosetta x87 is running. Proceeding to launch WoW.")
					// Command 2: Launch WoW.exe via rosettax87
					if crossoverPath == "" || turtlewowPath == "" {
						dialog.ShowError(fmt.Errorf("CrossOver path or TurtleWoW path is not set. Cannot launch WoW."), myWindow)
						return
					}

					// Determine MTL_HUD_ENABLED value based on checkbox
					mtlHudValue := "0"
					if enableMetalHud {
						mtlHudValue = "1"
					}

					// Construct the new shell command
					shellCmd := fmt.Sprintf(`cd %s && WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s %s %s %s`,
						quotePathForShell(turtlewowPath), // Added cd to turtlewowPath
						mtlHudValue,                      // Use dynamic value
						quotePathForShell(rosettaExecutable),
						quotePathForShell(wineloader2Path),
						quotePathForShell(turtlewowExePath)) // turtlewowExePath is defined at the start of launchGameFunc

					// Escape the entire shell command for AppleScript
					escapedShellCmd := escapeStringForAppleScript(shellCmd)

					cmd2Script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", escapedShellCmd)

					log.Println("Executing updated WoW launch command via AppleScript...")
					if !runOsascript(cmd2Script, myWindow) {
						// Error already shown by runOsascript
						return
					}

					log.Println("Launch commands executed. Check the new terminal windows.")
					dialog.ShowInformation("Launched", "World of Warcraft is starting. Enjoy.", myWindow)
				} else {
					log.Println("User cancelled WoW launch after rosetta x87 initiation.")
					dialog.ShowInformation("Cancelled", "WoW launch was cancelled.", myWindow)
				}
			}, myWindow)

		// The following lines are now moved inside the dialog.ShowConfirm callback
		// // Command 2: Launch WoW.exe via CrossOver
		// ... (rest of the original cmd2 logic was here)
	}

	// --- Button Definitions ---
	patchTurtleWoWButton = widget.NewButton("Patch TurtleWoW", patchTurtleWoWFunc)
	patchCrossOverButton = widget.NewButton("Patch CrossOver", patchCrossOverFunc)
	launchButton = widget.NewButton("Launch Game", func() { // Wrap the call in an anonymous function
		launchGameFunc(myWindow) // Pass myWindow to the actual handler
	})

	// --- Initial Check for Default CrossOver Path ---
	if info, err := os.Stat(defaultCrossOverPath); err == nil && info.IsDir() {
		crossoverPath = defaultCrossOverPath
		log.Println("Pre-set CrossOver to default:", defaultCrossOverPath)
		// No need to reset patchesAppliedCrossOver here, updateAllStatuses will check
	}

	// --- UI Layout ---
	// Using Form layout for better alignment of labels and controls
	pathSelectionForm := widget.NewForm(
		widget.NewFormItem("CrossOver Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", selectCrossOverPath), crossoverPathLabel)),
		widget.NewFormItem("TurtleWoW Path:", container.NewBorder(nil, nil, nil, widget.NewButton("Set/Change", selectTurtleWoWPath), turtlewowPathLabel)),
	)

	patchOperationsLayout := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(3, // Label, Status, Button
			widget.NewLabel("TurtleWoW Patch:"), turtlewowStatusLabel, patchTurtleWoWButton,
		),
		container.NewGridWithColumns(3,
			widget.NewLabel("CrossOver Patch:"), crossoverStatusLabel, patchCrossOverButton,
		),
		widget.NewSeparator(),
	)

	myWindow.SetContent(container.NewVBox(
		pathSelectionForm,
		patchOperationsLayout,
		metalHudCheckbox, // Added Metal HUD checkbox
		container.NewPadded(launchButton), // Added padding to launchButton
	))

	updateAllStatuses() // Initial UI state update, including button states
	myWindow.ShowAndRun()
}

// Helper to quote paths for shell commands if they contain spaces or special chars
func quotePathForShell(path string) string {
	// A simple approach: always quote. More robust parsing might be needed for complex paths.
	return fmt.Sprintf(`"%s"`, path)
}
