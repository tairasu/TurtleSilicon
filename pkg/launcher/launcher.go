package launcher

import (
	"fmt"
	"log"
	"path/filepath"

	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

var EnableMetalHud = true // Default to enabled
var CustomEnvVars = ""    // Custom environment variables

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

	log.Println("Preparing to launch TurtleSilicon...")

	rosettaInTurtlePath := filepath.Join(paths.TurtlewowPath, "rosettax87")
	rosettaExecutable := filepath.Join(rosettaInTurtlePath, "rosettax87")
	wineloader2Path := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")
	wowExePath := filepath.Join(paths.TurtlewowPath, "wow.exe") // Corrected to wow.exe

	if !utils.PathExists(rosettaExecutable) {
		dialog.ShowError(fmt.Errorf("rosetta executable not found at %s. Ensure TurtleWoW patching was successful", rosettaExecutable), myWindow)
		return
	}
	if !utils.PathExists(wineloader2Path) {
		dialog.ShowError(fmt.Errorf("patched wineloader2 not found at %s. Ensure CrossOver patching was successful", wineloader2Path), myWindow)
		return
	}
	if !utils.PathExists(wowExePath) {
		dialog.ShowError(fmt.Errorf("wow.exe not found at %s. Ensure your TurtleWoW directory is correct", wowExePath), myWindow)
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
	envVars := fmt.Sprintf(`WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s`, mtlHudValue)
	if CustomEnvVars != "" {
		envVars = CustomEnvVars + " " + envVars
	}

	shellCmd := fmt.Sprintf(`cd %s && %s %s %s %s`,
		utils.QuotePathForShell(paths.TurtlewowPath),
		envVars,
		utils.QuotePathForShell(rosettaExecutable),
		utils.QuotePathForShell(wineloader2Path),
		utils.QuotePathForShell(wowExePath))

	escapedShellCmd := utils.EscapeStringForAppleScript(shellCmd)
	cmd2Script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", escapedShellCmd)

	log.Println("Executing WoW launch command via AppleScript...")
	if !utils.RunOsascript(cmd2Script, myWindow) {
		return
	}

	log.Println("Launch command executed. Check the new terminal window.")
}
