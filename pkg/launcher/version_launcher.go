package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"
	"turtlesilicon/pkg/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// Version-specific launcher state management
var (
	versionGameProcesses = make(map[string]*exec.Cmd)
	versionGameRunning   = make(map[string]bool)
	versionGameMutex     sync.Mutex
)

// getCurrentVersionFromManager gets the version settings for a specific version ID
func getCurrentVersionFromManager(versionID string) *version.GameVersion {
	vm, err := version.LoadVersionManager()
	if err != nil {
		debug.Printf("Failed to load version manager: %v", err)
		return nil
	}

	ver, err := vm.GetVersion(versionID)
	if err != nil {
		debug.Printf("Failed to get version %s: %v", versionID, err)
		return nil
	}

	return ver
}

// LaunchVersionGame launches a specific version of the game
func LaunchVersionGame(myWindow fyne.Window, versionID string, gamePath string, crossoverPath string, executableName string, enableMetalHud bool, customEnvVars string, autoDeleteWdb bool) {
	debug.Printf("Launch Game button clicked for version: %s", versionID)

	if crossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set for version %s. Please set it in the patcher.", versionID), myWindow)
		return
	}
	if gamePath == "" {
		dialog.ShowError(fmt.Errorf("Game path not set for version %s. Please set it in the patcher.", versionID), myWindow)
		return
	}

	// Check if game is already running for this version
	versionGameMutex.Lock()
	if versionGameRunning[versionID] {
		versionGameMutex.Unlock()
		dialog.ShowInformation("Game Already Running", fmt.Sprintf("The game is already running for version %s.", versionID), myWindow)
		return
	}
	versionGameMutex.Unlock()

	debug.Printf("Preparing to launch %s...", versionID)

	// Determine the executable path
	gameExePath := filepath.Join(gamePath, executableName)
	if !utils.PathExists(gameExePath) {
		dialog.ShowError(fmt.Errorf("game executable not found at %s. Ensure your game directory is correct", gameExePath), myWindow)
		return
	}

	// Auto-delete WDB directory if enabled
	if autoDeleteWdb {
		deleteWDBDirectories(gamePath, versionID)
	}

	// For non-TurtleSilicon versions, we launch differently
	if versionID == "turtlesilicon" {
		// Use existing TurtleSilicon launch logic
		launchTurtleSiliconVersion(myWindow, gamePath, crossoverPath, gameExePath, enableMetalHud, customEnvVars)
	} else {
		// Use new launch method for other versions
		launchOtherVersion(myWindow, versionID, gamePath, crossoverPath, gameExePath, enableMetalHud, customEnvVars)
	}
}

// launchTurtleSiliconVersion launches using the existing TurtleSilicon method
func launchTurtleSiliconVersion(myWindow fyne.Window, gamePath string, crossoverPath string, gameExePath string, enableMetalHud bool, customEnvVars string) {
	debug.Println("Using TurtleSilicon launch method")

	// Get the current TurtleSilicon version settings
	currentVer := getCurrentVersionFromManager("turtlesilicon")

	// Temporarily set the legacy paths and settings for the existing launch function
	originalTurtlewowPath := paths.TurtlewowPath
	originalCrossoverPath := paths.CrossoverPath
	originalEnableMetalHud := EnableMetalHud
	originalCustomEnvVars := CustomEnvVars
	originalPatchesAppliedTurtleWoW := paths.PatchesAppliedTurtleWoW
	originalPatchesAppliedCrossOver := paths.PatchesAppliedCrossOver

	// Also temporarily set user preferences for terminal setting
	originalPrefs, _ := utils.LoadPrefs()
	tempPrefs := *originalPrefs // Copy the preferences
	if currentVer != nil {
		tempPrefs.ShowTerminalNormally = currentVer.Settings.ShowTerminalNormally
		utils.SavePrefs(&tempPrefs)
	}

	// Set the paths and settings for this version
	paths.TurtlewowPath = gamePath
	paths.CrossoverPath = crossoverPath
	EnableMetalHud = enableMetalHud
	CustomEnvVars = customEnvVars

	// Set patch status based on version-aware checking
	paths.PatchesAppliedTurtleWoW = true // We know patches are applied if we got this far
	// Check CrossOver patch status
	wineloader2Path := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
	paths.PatchesAppliedCrossOver = utils.PathExists(wineloader2Path)

	// Restore original values after launch
	defer func() {
		paths.TurtlewowPath = originalTurtlewowPath
		paths.CrossoverPath = originalCrossoverPath
		EnableMetalHud = originalEnableMetalHud
		CustomEnvVars = originalCustomEnvVars
		paths.PatchesAppliedTurtleWoW = originalPatchesAppliedTurtleWoW
		paths.PatchesAppliedCrossOver = originalPatchesAppliedCrossOver

		// Restore original preferences
		utils.SavePrefs(originalPrefs)
	}()

	// Call the existing launch function
	LaunchGame(myWindow)
}

// launchOtherVersion launches other versions using rosettax87 direct execution
func launchOtherVersion(myWindow fyne.Window, versionID string, gamePath string, crossoverPath string, gameExePath string, enableMetalHud bool, customEnvVars string) {
	debug.Printf("Launching %s using rosettax87 direct execution", versionID)

	wineloader2Path := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
	rosettaX87ExePath := filepath.Join(gamePath, "rosettax87", "rosettax87")

	if !utils.PathExists(wineloader2Path) {
		dialog.ShowError(fmt.Errorf("patched wineloader2 not found at %s. Ensure CrossOver patching was successful", wineloader2Path), myWindow)
		return
	}

	if !utils.PathExists(rosettaX87ExePath) {
		dialog.ShowError(fmt.Errorf("rosettax87 binary not found at %s. Ensure game patching was successful", rosettaX87ExePath), myWindow)
		return
	}

	mtlHudValue := "0"
	if enableMetalHud {
		mtlHudValue = "1"
	}

	// Prepare environment variables for other versions (include WINEDLLOVERRIDES for d3d9.dll graphics)
	envVars := fmt.Sprintf(`WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s MVK_CONFIG_SYNCHRONOUS_QUEUE_SUBMITS=1 DXVK_ASYNC=1`, mtlHudValue)
	if customEnvVars != "" {
		envVars = customEnvVars + " " + envVars
	}

	// Direct execution without service dependency
	shellCmd := fmt.Sprintf(`cd %s && %s %s %s %s`,
		utils.QuotePathForShell(gamePath),
		envVars,
		utils.QuotePathForShell(rosettaX87ExePath),
		utils.QuotePathForShell(wineloader2Path),
		utils.QuotePathForShell(gameExePath))

	// Check version-specific preference for terminal display
	// Get the current version to access its settings
	currentVer := getCurrentVersionFromManager(versionID)
	showTerminal := false
	if currentVer != nil {
		showTerminal = currentVer.Settings.ShowTerminalNormally
	}

	if showTerminal {
		// Use external Terminal.app
		escapedShellCmd := utils.EscapeStringForAppleScript(shellCmd)
		cmd2Script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", escapedShellCmd)

		debug.Printf("Executing %s launch command via AppleScript...", versionID)
		if !utils.RunOsascript(cmd2Script, myWindow) {
			return
		}

		debug.Printf("Launch command executed for %s. Check the new terminal window.", versionID)
	} else {
		// Use integrated terminal
		debug.Printf("Shell command for %s integrated terminal: %s", versionID, shellCmd)
		debug.Printf("Executing %s launch command with integrated terminal...", versionID)
		if err := runVersionGameIntegrated(myWindow, versionID, shellCmd); err != nil {
			dialog.ShowError(fmt.Errorf("failed to launch %s: %v", versionID, err), myWindow)
			return
		}
		debug.Printf("%s launched with integrated terminal. Check the application logs for output.", versionID)
	}
}

// runVersionGameIntegrated runs a version-specific game with integrated terminal output
func runVersionGameIntegrated(parentWindow fyne.Window, versionID string, shellCmd string) error {
	versionGameMutex.Lock()
	defer versionGameMutex.Unlock()

	if versionGameRunning[versionID] {
		return fmt.Errorf("game is already running for version %s", versionID)
	}

	versionGameRunning[versionID] = true

	// Parse the shell command to extract components
	debug.Printf("Parsing shell command for %s: %s", versionID, shellCmd)

	// Create the command without context cancellation
	cmd := exec.Command("sh", "-c", shellCmd)

	// Set up stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		versionGameRunning[versionID] = false
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		versionGameRunning[versionID] = false
		return err
	}

	versionGameProcesses[versionID] = cmd

	// Start the process
	if err := cmd.Start(); err != nil {
		versionGameRunning[versionID] = false
		return err
	}

	// Monitor output in goroutines
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			debug.Printf("%s STDOUT: %s", versionID, line)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			debug.Printf("%s STDERR: %s", versionID, line)
		}
	}()

	// Wait for the process to complete in a goroutine
	go func() {
		defer func() {
			versionGameMutex.Lock()
			versionGameRunning[versionID] = false
			delete(versionGameProcesses, versionID)
			versionGameMutex.Unlock()
		}()

		if err := cmd.Wait(); err != nil {
			debug.Printf("%s process ended with error: %v", versionID, err)
		} else {
			debug.Printf("%s process ended successfully", versionID)
		}
	}()

	return nil
}

// IsVersionGameRunning returns true if the game is currently running for a specific version
func IsVersionGameRunning(versionID string) bool {
	versionGameMutex.Lock()
	defer versionGameMutex.Unlock()
	return versionGameRunning[versionID]
}

// StopVersionGame forcefully stops the running game for a specific version
func StopVersionGame(versionID string) error {
	versionGameMutex.Lock()
	defer versionGameMutex.Unlock()

	if !versionGameRunning[versionID] {
		return fmt.Errorf("no game process is running for version %s", versionID)
	}

	process, exists := versionGameProcesses[versionID]
	if !exists || process == nil {
		return fmt.Errorf("no game process found for version %s", versionID)
	}

	// Try to terminate gracefully first
	if err := process.Process.Signal(os.Interrupt); err != nil {
		// If that fails, force kill
		return process.Process.Kill()
	}

	return nil
}

// deleteWDBDirectories deletes WDB directories, checking both direct and Cache subdirectory
func deleteWDBDirectories(gamePath string, versionID string) {
	// Check for WDB in root directory
	wdbPath := filepath.Join(gamePath, "WDB")
	if utils.DirExists(wdbPath) {
		debug.Printf("Auto-deleting WDB directory for %s: %s", versionID, wdbPath)
		if err := os.RemoveAll(wdbPath); err != nil {
			debug.Printf("Warning: failed to auto-delete WDB directory for %s: %v", versionID, err)
		} else {
			debug.Printf("Successfully auto-deleted WDB directory for %s", versionID)
		}
	}

	// Check for WDB in Cache subdirectory (other versions might have it there)
	cacheWdbPath := filepath.Join(gamePath, "Cache", "WDB")
	if utils.DirExists(cacheWdbPath) {
		debug.Printf("Auto-deleting Cache/WDB directory for %s: %s", versionID, cacheWdbPath)
		if err := os.RemoveAll(cacheWdbPath); err != nil {
			debug.Printf("Warning: failed to auto-delete Cache/WDB directory for %s: %v", versionID, err)
		} else {
			debug.Printf("Successfully auto-deleted Cache/WDB directory for %s", versionID)
		}
	}

	// If neither was found, log it
	if !utils.DirExists(wdbPath) && !utils.DirExists(cacheWdbPath) {
		debug.Printf("WDB directory not found for %s, nothing to delete", versionID)
	}
}
