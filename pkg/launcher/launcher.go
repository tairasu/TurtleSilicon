package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/patching"
	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path
	"turtlesilicon/pkg/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

var EnableMetalHud = false // Default to disabled
var CustomEnvVars = ""     // Custom environment variables
var AutoDeleteWdb = true   // Default to enabled

// UI update callback for triggering status updates from launcher
var uiUpdateCallback func()

// Terminal state management
var (
	currentGameProcess *exec.Cmd
	isGameRunning      bool
	gameMutex          sync.Mutex
)

// runGameIntegrated runs the game with integrated terminal output
func runGameIntegrated(parentWindow fyne.Window, shellCmd string) error {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	if isGameRunning {
		return fmt.Errorf("game is already running")
	}

	isGameRunning = true

	// Parse the shell command to extract components
	// The shellCmd format is: cd <path> && <envVars> <rosettaExec> <wineloader> <wowExe>
	debug.Printf("Parsing shell command: %s", shellCmd)

	// Create the command without context cancellation
	cmd := exec.Command("sh", "-c", shellCmd)

	// Set up stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		isGameRunning = false
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		isGameRunning = false
		return err
	}

	currentGameProcess = cmd

	// Start the process
	if err := cmd.Start(); err != nil {
		isGameRunning = false
		return err
	}

	// Monitor output in goroutines
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			debug.Printf("GAME STDOUT: %s", line)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			debug.Printf("GAME STDERR: %s", line)
		}
	}()

	// Wait for the process to complete in a goroutine
	go func() {
		defer func() {
			gameMutex.Lock()
			isGameRunning = false
			currentGameProcess = nil
			gameMutex.Unlock()

			// Verify patch status after game closes to detect if TurtleWoW client deleted dlls.txt
			debug.Println("Game closed, verifying patch status...")
			verifyPatchStatusAfterGameClose()
		}()

		if err := cmd.Wait(); err != nil {
			debug.Printf("Game process ended with error: %v", err)
		} else {
			debug.Println("Game process ended successfully")
		}
	}()

	return nil
}

func LaunchGame(myWindow fyne.Window) {
	debug.Println("Launch Game button clicked")

	if paths.CrossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it in the patcher."), myWindow)
		return
	}
	if paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("game path not set. Please set it in the patcher."), myWindow)
		return
	}

	// Verify patch status before launching - this catches when TurtleWoW client has updated
	debug.Println("Verifying patch status before launch...")
	if !verifyPatchStatusBeforeLaunch(myWindow) {
		return // Don't launch if patches are invalid
	}

	if !paths.PatchesAppliedTurtleWoW || !paths.PatchesAppliedCrossOver {
		confirmed := false
		dialog.ShowConfirm("Warning", "Not all patches confirmed applied. Continue with launch?", func(c bool) {
			confirmed = c
		}, myWindow)
		if !confirmed {
			return
		}
	}

	// Check if game is already running
	gameMutex.Lock()
	if isGameRunning {
		gameMutex.Unlock()
		dialog.ShowInformation("Game Already Running", "The game is already running.", myWindow)
		return
	}
	gameMutex.Unlock()

	debug.Println("Preparing to launch TurtleSilicon...")

	// Use the standard WoW.exe executable
	wowExePath := filepath.Join(paths.TurtlewowPath, "WoW.exe")

	// Continue with normal launch process
	continueLaunch(myWindow, wowExePath)
}

// continueLaunch continues the game launch process with the specified executable
func continueLaunch(myWindow fyne.Window, wowExePath string) {
	rosettaInTurtlePath := filepath.Join(paths.TurtlewowPath, "rosettax87")
	rosettaExecutable := filepath.Join(rosettaInTurtlePath, "rosettax87")
	wineloader2Path := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")

	if !utils.PathExists(rosettaExecutable) {
		dialog.ShowError(fmt.Errorf("rosetta executable not found at %s. Ensure TurtleWoW patching was successful", rosettaExecutable), myWindow)
		return
	}
	if !utils.PathExists(wineloader2Path) {
		dialog.ShowError(fmt.Errorf("patched wineloader2 not found at %s. Ensure CrossOver patching was successful", wineloader2Path), myWindow)
		return
	}
	if !utils.PathExists(wowExePath) {
		dialog.ShowError(fmt.Errorf("WoW executable not found at %s. Ensure your TurtleWoW directory is correct", wowExePath), myWindow)
		return
	}

	// Auto-delete WDB directory if enabled
	if AutoDeleteWdb {
		deleteLegacyWDBDirectories(paths.TurtlewowPath)
	}

	// Launch WoW using direct rosettax87 execution (no service required)
	debug.Println("Launching WoW with direct rosettax87 execution.")

	if paths.CrossoverPath == "" || paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path or game path is not set. Cannot launch WoW."), myWindow)
		return
	}

	mtlHudValue := "0"
	if EnableMetalHud {
		mtlHudValue = "1"
	}

	// Prepare environment variables
	envVars := fmt.Sprintf(`WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s MVK_CONFIG_SYNCHRONOUS_QUEUE_SUBMITS=1 DXVK_ASYNC=1`, mtlHudValue)
	if CustomEnvVars != "" {
		envVars = CustomEnvVars + " " + envVars
	}

	shellCmd := fmt.Sprintf(`cd %s && %s %s %s %s`,
		utils.QuotePathForShell(paths.TurtlewowPath),
		envVars,
		utils.QuotePathForShell(rosettaExecutable),
		utils.QuotePathForShell(wineloader2Path),
		utils.QuotePathForShell(wowExePath))

	// Check user preference for terminal display
	prefs, _ := utils.LoadPrefs()

	if prefs.ShowTerminalNormally {
		// Use the old method with external Terminal.app
		escapedShellCmd := utils.EscapeStringForAppleScript(shellCmd)
		cmd2Script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", escapedShellCmd)

		debug.Println("Executing WoW launch command via AppleScript...")
		if !utils.RunOsascript(cmd2Script, myWindow) {
			return
		}

		debug.Println("Launch command executed. Check the new terminal window.")
	} else {
		// Use integrated terminal
		debug.Printf("Shell command for integrated terminal: %s", shellCmd)
		debug.Println("Executing WoW launch command with integrated terminal...")
		if err := runGameIntegrated(myWindow, shellCmd); err != nil {
			dialog.ShowError(fmt.Errorf("failed to launch game: %v", err), myWindow)
			return
		}
		debug.Println("Game launched with integrated terminal. Check the application logs for output.")
	}
}

// IsGameRunning returns true if the game is currently running
func IsGameRunning() bool {
	gameMutex.Lock()
	defer gameMutex.Unlock()
	return isGameRunning
}

// StopGame forcefully stops the running game
func StopGame() error {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	if !isGameRunning || currentGameProcess == nil {
		return fmt.Errorf("no game process is running")
	}

	// Try to terminate gracefully first
	if err := currentGameProcess.Process.Signal(os.Interrupt); err != nil {
		// If that fails, force kill
		return currentGameProcess.Process.Kill()
	}

	return nil
}

// deleteLegacyWDBDirectories deletes WDB directories for legacy launcher
func deleteLegacyWDBDirectories(gamePath string) {
	// Check for WDB in root directory
	wdbPath := filepath.Join(gamePath, "WDB")
	if utils.DirExists(wdbPath) {
		debug.Printf("Auto-deleting WDB directory: %s", wdbPath)
		if err := os.RemoveAll(wdbPath); err != nil {
			debug.Printf("Warning: failed to auto-delete WDB directory: %v", err)
		} else {
			debug.Printf("Successfully auto-deleted WDB directory")
		}
	}

	// Check for WDB in Cache subdirectory
	cacheWdbPath := filepath.Join(gamePath, "Cache", "WDB")
	if utils.DirExists(cacheWdbPath) {
		debug.Printf("Auto-deleting Cache/WDB directory: %s", cacheWdbPath)
		if err := os.RemoveAll(cacheWdbPath); err != nil {
			debug.Printf("Warning: failed to auto-delete Cache/WDB directory: %v", err)
		} else {
			debug.Printf("Successfully auto-deleted Cache/WDB directory")
		}
	}

	// If neither was found, log it
	if !utils.DirExists(wdbPath) && !utils.DirExists(cacheWdbPath) {
		debug.Printf("WDB directory not found, nothing to delete")
	}
}

// SetUIUpdateCallback sets the callback function for triggering UI updates
func SetUIUpdateCallback(callback func()) {
	uiUpdateCallback = callback
}

// verifyPatchStatusBeforeLaunch checks patch status before launching and updates UI if needed
// Returns true if launch should proceed, false if patches are invalid
func verifyPatchStatusBeforeLaunch(myWindow fyne.Window) bool {
	debug.Println("Pre-launch patch verification...")

	// Load current version
	vm, err := version.LoadVersionManager()
	if err != nil {
		debug.Printf("Failed to load version manager for pre-launch verification: %v", err)
		return true // Continue launch if we can't verify
	}

	currentVer, err := vm.GetCurrentVersion()
	if err != nil {
		debug.Printf("Failed to get current version for pre-launch verification: %v", err)
		return true // Continue launch if we can't verify
	}

	if currentVer.GamePath == "" {
		debug.Println("No game path set, skipping pre-launch verification")
		return true
	}

	// Check if patches are still applied
	patchesStillValid := patching.CheckVersionPatchingStatus(currentVer.GamePath, currentVer.UsesRosettaPatching, currentVer.UsesDivxDecoderPatch, currentVer.ID)

	if !patchesStillValid {
		debug.Println("⚠️ Pre-launch check: Patches are no longer valid!")

		// Reset patch status so UI reflects reality
		_, crossoverPatched := paths.GetVersionPatchingStatus(currentVer.ID)
		paths.SetVersionPatchingStatus(currentVer.ID, false, crossoverPatched)
		paths.PatchesAppliedTurtleWoW = false

		// Trigger UI update immediately
		if uiUpdateCallback != nil {
			uiUpdateCallback()
		}

		// Show user-friendly dialog
		dialog.ShowError(fmt.Errorf("Game patches are no longer valid!\n\nThe TurtleWoW client may have updated and reset your patches.\nPlease re-patch the game before launching."), myWindow)

		return false // Don't launch
	}

	debug.Println("✓ Pre-launch verification passed")
	return true // Launch is OK
}

// verifyPatchStatusAfterGameClose checks if patches are still valid after game closes
// This detects when TurtleWoW client updates itself and deletes dlls.txt content
func verifyPatchStatusAfterGameClose() {
	debug.Println("Verifying patch status after game close...")

	// Add a small delay to ensure the game process has fully terminated
	time.Sleep(1 * time.Second)

	// Load current version
	vm, err := version.LoadVersionManager()
	if err != nil {
		debug.Printf("Failed to load version manager for patch verification: %v", err)
		return
	}

	currentVer, err := vm.GetCurrentVersion()
	if err != nil {
		debug.Printf("Failed to get current version for patch verification: %v", err)
		return
	}

	if currentVer.GamePath == "" {
		debug.Println("No game path set, skipping patch verification")
		return
	}

	// Check if patches are still applied
	patchesStillValid := patching.CheckVersionPatchingStatus(currentVer.GamePath, currentVer.UsesRosettaPatching, currentVer.UsesDivxDecoderPatch, currentVer.ID)

	if !patchesStillValid {
		debug.Println("⚠️ Patches are no longer valid! TurtleWoW client may have updated and reset dlls.txt")

		// Reset patch status so user needs to re-patch
		_, crossoverPatched := paths.GetVersionPatchingStatus(currentVer.ID)
		paths.SetVersionPatchingStatus(currentVer.ID, false, crossoverPatched)

		// For legacy system compatibility, also reset the global flags
		paths.PatchesAppliedTurtleWoW = false

		debug.Println("Patch status reset. User will need to re-patch the game.")

		// Trigger UI update to reflect the changed patch status
		if uiUpdateCallback != nil {
			uiUpdateCallback()
		}
	} else {
		debug.Println("✓ Patches verified successfully - still valid")
	}
}
