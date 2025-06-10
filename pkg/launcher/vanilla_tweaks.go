package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// ApplyVanillaTweaks applies vanilla-tweaks to WoW.exe to create WoW-tweaked.exe
func ApplyVanillaTweaks(myWindow fyne.Window) error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}
	if paths.CrossoverPath == "" {
		return fmt.Errorf("CrossOver path not set")
	}

	// Get the current working directory (where the app executable is located)
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	appDir := filepath.Dir(execPath)

	// Check if we're in development mode
	vanillaTweaksPath := filepath.Join(appDir, "winerosetta", "vanilla-tweaks.exe")
	if !utils.PathExists(vanillaTweaksPath) {
		// Try relative path from current working directory (for development)
		workingDir, _ := os.Getwd()
		vanillaTweaksPath = filepath.Join(workingDir, "winerosetta", "vanilla-tweaks.exe")
		if !utils.PathExists(vanillaTweaksPath) {
			return fmt.Errorf("vanilla-tweaks.exe not found")
		}
	}

	wowExePath := filepath.Join(paths.TurtlewowPath, "WoW.exe")
	wineloader2Path := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")

	if !utils.PathExists(wowExePath) {
		return fmt.Errorf("WoW.exe not found at %s", wowExePath)
	}
	if !utils.PathExists(wineloader2Path) {
		return fmt.Errorf("wineloader2 not found at %s", wineloader2Path)
	}

	// First, copy vanilla-tweaks.exe to the TurtleWoW directory temporarily
	tempVanillaTweaksPath := filepath.Join(paths.TurtlewowPath, "vanilla-tweaks.exe")

	// Copy vanilla-tweaks.exe to TurtleWoW directory
	debug.Printf("Copying vanilla-tweaks.exe from %s to %s", vanillaTweaksPath, tempVanillaTweaksPath)
	sourceFile, err := os.Open(vanillaTweaksPath)
	if err != nil {
		return fmt.Errorf("failed to open vanilla-tweaks.exe: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(tempVanillaTweaksPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary vanilla-tweaks.exe: %v", err)
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy vanilla-tweaks.exe: %v", err)
	}

	// Ensure the copied file is executable
	if err := os.Chmod(tempVanillaTweaksPath, 0755); err != nil {
		debug.Printf("Warning: failed to set executable permission on vanilla-tweaks.exe: %v", err)
	}

	// Build the command to apply vanilla-tweaks using the correct format:
	// cd "path" && "wineloader2" ./vanilla-tweaks.exe --no-frilldistance -no-farclip ./WoW.exe
	shellCmd := fmt.Sprintf(`cd %s && %s ./vanilla-tweaks.exe --no-frilldistance --no-farclip ./WoW.exe`,
		utils.QuotePathForShell(paths.TurtlewowPath),
		utils.QuotePathForShell(wineloader2Path))

	debug.Printf("Applying vanilla-tweaks with command: %s", shellCmd)

	// Execute the command
	cmd := exec.Command("sh", "-c", shellCmd)
	output, err := cmd.CombinedOutput()

	debug.Printf("vanilla-tweaks command output: %s", string(output))

	// Clean up the temporary vanilla-tweaks.exe file
	if cleanupErr := os.Remove(tempVanillaTweaksPath); cleanupErr != nil {
		debug.Printf("Warning: failed to clean up temporary vanilla-tweaks.exe: %v", cleanupErr)
	}

	// Always check if the output file was created, regardless of exit code
	// Some Wine programs report error exit codes even when they succeed
	foundPath := GetWoWTweakedExecutablePath()
	if foundPath == "" {
		// Only report error if no output file was created
		if err != nil {
			debug.Printf("vanilla-tweaks command failed: %v", err)
			return fmt.Errorf("failed to apply vanilla-tweaks: %v\nOutput: %s", err, string(output))
		} else {
			return fmt.Errorf("vanilla-tweaks completed but WoW-tweaked.exe was not created\nOutput: %s", string(output))
		}
	}

	// If we found the file but there was an error code, log it as a warning
	if err != nil {
		debug.Printf("vanilla-tweaks reported error but output file was created: %v", err)
	}

	debug.Println("vanilla-tweaks applied successfully")
	return nil
}

// CheckForVanillaTweaksExecutable checks if vanilla-tweaks.exe exists and is accessible
func CheckForVanillaTweaksExecutable() bool {
	// Get the current working directory (where the app executable is located)
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	appDir := filepath.Dir(execPath)

	// Check if we're in development mode (running from VSCode)
	vanillaTweaksPath := filepath.Join(appDir, "winerosetta", "vanilla-tweaks.exe")
	if utils.PathExists(vanillaTweaksPath) {
		return true
	}

	// Try relative path from current working directory (for development)
	workingDir, _ := os.Getwd()
	vanillaTweaksPath = filepath.Join(workingDir, "winerosetta", "vanilla-tweaks.exe")
	return utils.PathExists(vanillaTweaksPath)
}

// GetVanillaTweaksExecutablePath returns the path to vanilla-tweaks.exe if it exists
func GetVanillaTweaksExecutablePath() (string, error) {
	// Get the current working directory (where the app executable is located)
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	appDir := filepath.Dir(execPath)

	// Check if we're in development mode (running from VSCode)
	vanillaTweaksPath := filepath.Join(appDir, "winerosetta", "vanilla-tweaks.exe")
	if utils.PathExists(vanillaTweaksPath) {
		return vanillaTweaksPath, nil
	}

	// Try relative path from current working directory (for development)
	workingDir, _ := os.Getwd()
	vanillaTweaksPath = filepath.Join(workingDir, "winerosetta", "vanilla-tweaks.exe")
	if utils.PathExists(vanillaTweaksPath) {
		return vanillaTweaksPath, nil
	}

	return "", fmt.Errorf("vanilla-tweaks.exe not found")
}

// CheckForWoWTweakedExecutable checks if WoW_tweaked.exe exists in the TurtleWoW directory
func CheckForWoWTweakedExecutable() bool {
	if paths.TurtlewowPath == "" {
		return false
	}

	testPath := filepath.Join(paths.TurtlewowPath, "WoW_tweaked.exe")
	return utils.PathExists(testPath)
}

// GetWoWTweakedExecutablePath returns the path to the WoW_tweaked.exe file if it exists
func GetWoWTweakedExecutablePath() string {
	if paths.TurtlewowPath == "" {
		return ""
	}

	testPath := filepath.Join(paths.TurtlewowPath, "WoW_tweaked.exe")
	if utils.PathExists(testPath) {
		return testPath
	}
	return ""
}

// HandleVanillaTweaksRequest handles the case when vanilla-tweaks is enabled but WoW-tweaked.exe doesn't exist
func HandleVanillaTweaksRequest(myWindow fyne.Window, callback func()) {
	dialog.ShowConfirm("Vanilla-tweaks not found",
		"WoW-tweaked.exe was not found in your TurtleWoW directory.\n\nWould you like TurtleSilicon to automatically apply vanilla-tweaks for you?",
		func(confirmed bool) {
			if confirmed {
				if err := ApplyVanillaTweaks(myWindow); err != nil {
					dialog.ShowError(fmt.Errorf("failed to apply vanilla-tweaks: %v", err), myWindow)
					return
				}
				// After successful patching, execute the callback
				callback()
			}
		}, myWindow)
}
