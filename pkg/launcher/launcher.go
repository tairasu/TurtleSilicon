package launcher

import (
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path
)

var EnableMetalHud = true // Default to enabled

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

	appleScriptSafeRosettaDir := utils.EscapeStringForAppleScript(rosettaInTurtlePath)
	cmd1Script := fmt.Sprintf("tell application \"Terminal\" to do script \"cd \" & quoted form of \"%s\" & \" && sudo ./rosettax87\"", appleScriptSafeRosettaDir)

	log.Println("Launching rosettax87 (requires sudo password in new terminal)...")
	if !utils.RunOsascript(cmd1Script, myWindow) {
		return
	}

	dialog.ShowConfirm("Action Required",
		"The rosetta x87 terminal has been initiated.\n\n"+
			"1. Please enter your sudo password in that new terminal window.\n"+
			"2. Wait for rosetta x87 to fully start.\n\n"+
			"Click Yes once rosetta x87 is running and you have entered the password.\n"+
			"Click No to abort launching WoW.",
		func(confirmed bool) {
			if confirmed {
				log.Println("User confirmed rosetta x87 is running. Proceeding to launch WoW.")
				if paths.CrossoverPath == "" || paths.TurtlewowPath == "" {
					dialog.ShowError(fmt.Errorf("CrossOver path or TurtleWoW path is not set. Cannot launch WoW."), myWindow)
					return
				}

				mtlHudValue := "0"
				if EnableMetalHud {
					mtlHudValue = "1"
				}

				shellCmd := fmt.Sprintf(`cd %s && WINEDLLOVERRIDES="d3d9=n,b" MTL_HUD_ENABLED=%s %s %s %s`,
					utils.QuotePathForShell(paths.TurtlewowPath),
					mtlHudValue,
					utils.QuotePathForShell(rosettaExecutable),
					utils.QuotePathForShell(wineloader2Path),
					utils.QuotePathForShell(wowExePath))

				escapedShellCmd := utils.EscapeStringForAppleScript(shellCmd)
				cmd2Script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", escapedShellCmd)

				log.Println("Executing updated WoW launch command via AppleScript...")
				if !utils.RunOsascript(cmd2Script, myWindow) {
					return
				}

				log.Println("Launch commands executed. Check the new terminal windows.")
				dialog.ShowInformation("Launched", "World of Warcraft is starting. Enjoy.", myWindow)
			} else {
				log.Println("User cancelled WoW launch after rosetta x87 initiation.")
				dialog.ShowInformation("Cancelled", "WoW launch was cancelled.", myWindow)
			}
		}, myWindow)
}
