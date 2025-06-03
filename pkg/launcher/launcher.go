package launcher

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

var EnableMetalHud = true       // Default to enabled
var CustomEnvVars = ""          // Custom environment variables
var EnableVanillaTweaks = false // Default to disabled

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
	log.Printf("Parsing shell command: %s", shellCmd)

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
			log.Printf("GAME STDOUT: %s", line)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("GAME STDERR: %s", line)
		}
	}()

	// Wait for the process to complete in a goroutine
	go func() {
		defer func() {
			gameMutex.Lock()
			isGameRunning = false
			currentGameProcess = nil
			gameMutex.Unlock()
		}()

		if err := cmd.Wait(); err != nil {
			log.Printf("Game process ended with error: %v", err)
		} else {
			log.Println("Game process ended successfully")
		}
	}()

	return nil
}

func LaunchGame(myWindow fyne.Window) {
	log.Println("Launch Game button clicked")

	if paths.CrossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it in the patcher."), myWindow)
		return
	}
	if paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it in the patcher."), myWindow)
		return
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

	log.Println("Preparing to launch TurtleSilicon...")

	// Determine which WoW executable to use based on vanilla-tweaks preference
	var wowExePath string
	if EnableVanillaTweaks {
		if !CheckForWoWTweakedExecutable() {
			// Show dialog asking if user wants us to apply vanilla-tweaks
			HandleVanillaTweaksRequest(myWindow, func() {
				// After successful patching, continue with launch using the tweaked executable
				wowTweakedExePath := GetWoWTweakedExecutablePath()
				if wowTweakedExePath != "" {
					continueLaunch(myWindow, wowTweakedExePath)
				} else {
					dialog.ShowError(fmt.Errorf("failed to find WoW-tweaked.exe after patching"), myWindow)
				}
			})
			return // Exit early since dialog will handle the continuation
		}
		wowExePath = GetWoWTweakedExecutablePath()
	} else {
		wowExePath = filepath.Join(paths.TurtlewowPath, "WoW.exe")
	}

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

	// Since RosettaX87 service is already running, we can directly launch WoW
	log.Println("RosettaX87 service is running. Proceeding to launch WoW.")

	if paths.CrossoverPath == "" || paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path or TurtleWoW path is not set. Cannot launch WoW."), myWindow)
		return
	}

	mtlHudValue := "0"
	if EnableMetalHud {
		mtlHudValue = "1"
	}

	// Prepare environment variables
	envVars := fmt.Sprintf(`WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s MVK_CONFIG_SYNCHRONOUS_QUEUE_SUBMITS=1`, mtlHudValue)
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

		log.Println("Executing WoW launch command via AppleScript...")
		if !utils.RunOsascript(cmd2Script, myWindow) {
			return
		}

		log.Println("Launch command executed. Check the new terminal window.")
	} else {
		// Use integrated terminal
		log.Printf("Shell command for integrated terminal: %s", shellCmd)
		log.Println("Executing WoW launch command with integrated terminal...")
		if err := runGameIntegrated(myWindow, shellCmd); err != nil {
			dialog.ShowError(fmt.Errorf("failed to launch game: %v", err), myWindow)
			return
		}
		log.Println("Game launched with integrated terminal. Check the application logs for output.")
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
