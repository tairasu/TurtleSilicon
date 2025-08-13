package patching

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// PatchVersionGame patches a game version based on its configuration
func PatchVersionGame(myWindow fyne.Window, updateAllStatuses func(), gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool) {
	debug.Printf("Patching game at path: %s, rosetta=%v, divx=%v", gamePath, usesRosettaPatching, usesDivxDecoderPatch)

	if gamePath == "" {
		dialog.ShowError(fmt.Errorf("game path not set. Please set it first."), myWindow)
		return
	}

	if usesDivxDecoderPatch {
		// For non-TurtleSilicon versions, only apply DivxDecoder replacement
		patchWithDivxDecoderMethod(myWindow, updateAllStatuses, gamePath)
	} else {
		// For TurtleSilicon, use the full rosettax87 patching
		// Temporarily set paths.TurtlewowPath so existing patching works
		originalPath := paths.TurtlewowPath
		paths.TurtlewowPath = gamePath
		defer func() {
			paths.TurtlewowPath = originalPath
		}()

		PatchTurtleWoW(myWindow, updateAllStatuses)
	}
}

// patchWithDivxDecoderMethod implements the new patching method for other versions
func patchWithDivxDecoderMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Applying DivxDecoder patching method")

	divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")
	divxDecoderBackupPath := filepath.Join(gamePath, "DivxDecoder.dll.backup")
	d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

	// Step 1: Create backup of existing DivxDecoder.dll if it exists
	if utils.PathExists(divxDecoderPath) {
		debug.Printf("Creating backup of existing DivxDecoder.dll at: %s", divxDecoderBackupPath)

		// Read the original file
		originalData, err := os.ReadFile(divxDecoderPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to read existing DivxDecoder.dll: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}

		// Write the backup
		if err := os.WriteFile(divxDecoderBackupPath, originalData, 0644); err != nil {
			errMsg := fmt.Sprintf("failed to create backup of DivxDecoder.dll: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully created backup of DivxDecoder.dll")

		// Remove the original file
		if err := os.Remove(divxDecoderPath); err != nil {
			errMsg := fmt.Sprintf("failed to remove existing DivxDecoder.dll: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully removed original DivxDecoder.dll")
	}

	// Step 2: Copy winerosetta.dll to the game directory as DivxDecoder.dll
	debug.Printf("Copying winerosetta.dll as DivxDecoder.dll to: %s", divxDecoderPath)

	resource, err := fyne.LoadResourceFromPath("winerosetta/winerosetta.dll")
	if err != nil {
		errMsg := fmt.Sprintf("failed to open bundled winerosetta.dll resource: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}

	destinationFile, err := os.Create(divxDecoderPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create DivxDecoder.dll file: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
	if err != nil {
		errMsg := fmt.Sprintf("failed to copy winerosetta.dll as DivxDecoder.dll: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}

	debug.Printf("Successfully copied winerosetta.dll as DivxDecoder.dll")

	// Step 3: Copy d3d9.dll for graphics
	debug.Printf("Copying d3d9.dll to: %s", d3d9DllPath)

	d3d9Resource, err := fyne.LoadResourceFromPath("winerosetta/d3d9.dll")
	if err != nil {
		errMsg := fmt.Sprintf("failed to open bundled d3d9.dll resource: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}

	d3d9File, err := os.Create(d3d9DllPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create d3d9.dll file: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}
	defer d3d9File.Close()

	_, err = io.Copy(d3d9File, bytes.NewReader(d3d9Resource.Content()))
	if err != nil {
		errMsg := fmt.Sprintf("failed to copy d3d9.dll: %v", err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		updateAllStatuses()
		return
	}

	debug.Printf("Successfully copied d3d9.dll")

	// Step 4: Copy rosettax87 service files (each version gets its own)
	rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
	if !utils.DirExists(rosettaX87Dir) {
		if err := os.MkdirAll(rosettaX87Dir, 0755); err != nil {
			errMsg := fmt.Sprintf("failed to create rosettax87 directory: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Created rosettax87 directory: %s", rosettaX87Dir)
	}

	// Copy rosettax87 executable files
	rosettaFilesToCopy := map[string]string{
		"rosettax87/rosettax87":           filepath.Join(rosettaX87Dir, "rosettax87"),
		"rosettax87/libRuntimeRosettax87": filepath.Join(rosettaX87Dir, "libRuntimeRosettax87"),
	}

	for resourceName, destPath := range rosettaFilesToCopy {
		debug.Printf("Processing rosetta resource: %s to %s", resourceName, destPath)

		// Check if file already exists and has correct size and hash
		if utils.PathExists(destPath) && utils.CompareFileWithBundledResource(destPath, resourceName) {
			debug.Printf("File %s already exists with correct size and hash, skipping copy", destPath)

			// Ensure executable permission for all rosettax87 files
			if err := os.Chmod(destPath, 0755); err != nil {
				debug.Printf("Warning: failed to set execute permission for existing %s: %v", destPath, err)
			}
			continue
		}

		if utils.PathExists(destPath) {
			debug.Printf("File %s exists but has incorrect size/hash, updating...", destPath)
		} else {
			debug.Printf("File %s does not exist, creating...", destPath)
		}

		resource, err := fyne.LoadResourceFromPath(resourceName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}

		destinationFile, err := os.Create(destPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}

		_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
		if err != nil {
			errMsg := fmt.Sprintf("failed to copy bundled resource to %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			destinationFile.Close()
			updateAllStatuses()
			return
		}
		destinationFile.Close()

		// Make rosettax87 executable
		if err := os.Chmod(destPath, 0755); err != nil {
			errMsg := fmt.Sprintf("failed to make %s executable: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully copied and made executable: %s to %s", resourceName, destPath)
	}

	dialog.ShowInformation("Success", "Game patching completed successfully.", myWindow)
	updateAllStatuses()
}

// patchWithRosettaMethod implements the existing TurtleWoW patching method
func patchWithRosettaMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Applying Rosetta patching method (legacy TurtleWoW method)")

	// This essentially calls the existing PatchTurtleWoW logic but with a different path
	// For now, we'll redirect to the existing function by temporarily setting the path
	// This is a temporary solution - ideally we'd refactor the existing patching code

	dialog.ShowInformation("Info", "Using legacy TurtleWoW patching method. Please use the existing TurtleWoW patch button.", myWindow)
	updateAllStatuses()
}

// UnpatchVersionGame unpatches a game version based on its configuration
func UnpatchVersionGame(myWindow fyne.Window, updateAllStatuses func(), gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool) {
	debug.Printf("Unpatching game at path: %s, rosetta=%v, divx=%v", gamePath, usesRosettaPatching, usesDivxDecoderPatch)

	if gamePath == "" {
		dialog.ShowError(fmt.Errorf("game path not set. Please set it first."), myWindow)
		return
	}

	if usesDivxDecoderPatch {
		// For non-TurtleSilicon versions, only remove DivxDecoder replacement
		unpatchWithDivxDecoderMethod(myWindow, updateAllStatuses, gamePath)
	} else {
		// For TurtleSilicon, use the full rosettax87 unpatching
		// Temporarily set paths.TurtlewowPath so existing unpatching works
		originalPath := paths.TurtlewowPath
		paths.TurtlewowPath = gamePath
		defer func() {
			paths.TurtlewowPath = originalPath
		}()

		UnpatchTurtleWoW(myWindow, updateAllStatuses)
	}
}

// unpatchWithDivxDecoderMethod removes the DivxDecoder.dll file and restores backup if available
func unpatchWithDivxDecoderMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Removing DivxDecoder patching")

	divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")
	divxDecoderBackupPath := filepath.Join(gamePath, "DivxDecoder.dll.backup")
	d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

	// Remove the patched DivxDecoder.dll
	if utils.PathExists(divxDecoderPath) {
		debug.Printf("Removing patched DivxDecoder.dll at: %s", divxDecoderPath)
		if err := os.Remove(divxDecoderPath); err != nil {
			errMsg := fmt.Sprintf("failed to remove DivxDecoder.dll: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully removed patched DivxDecoder.dll")
	}

	// Remove d3d9.dll
	if utils.PathExists(d3d9DllPath) {
		debug.Printf("Removing d3d9.dll at: %s", d3d9DllPath)
		if err := os.Remove(d3d9DllPath); err != nil {
			debug.Printf("Warning: failed to remove d3d9.dll: %v", err)
			// Don't fail the operation for this
		} else {
			debug.Printf("Successfully removed d3d9.dll")
		}
	}

	// Remove rosettax87 directory
	rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
	if utils.DirExists(rosettaX87Dir) {
		debug.Printf("Removing rosettax87 directory at: %s", rosettaX87Dir)
		if err := os.RemoveAll(rosettaX87Dir); err != nil {
			debug.Printf("Warning: failed to remove rosettax87 directory: %v", err)
			// Don't fail the operation for this
		} else {
			debug.Printf("Successfully removed rosettax87 directory")
		}
	}

	// Restore the original DivxDecoder.dll from backup if it exists
	if utils.PathExists(divxDecoderBackupPath) {
		debug.Printf("Restoring original DivxDecoder.dll from backup: %s", divxDecoderBackupPath)

		// Read the backup file
		backupData, err := os.ReadFile(divxDecoderBackupPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to read DivxDecoder.dll backup: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}

		// Write it back as the original file
		if err := os.WriteFile(divxDecoderPath, backupData, 0644); err != nil {
			errMsg := fmt.Sprintf("failed to restore DivxDecoder.dll from backup: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}

		// Remove the backup file
		if err := os.Remove(divxDecoderBackupPath); err != nil {
			debug.Printf("Warning: failed to remove backup file %s: %v", divxDecoderBackupPath, err)
			// Don't fail the operation for this
		}

		debug.Printf("Successfully restored original DivxDecoder.dll from backup")
	} else {
		debug.Printf("No DivxDecoder.dll backup found, original file was not present")
	}

	dialog.ShowInformation("Success", "Game unpatching completed successfully.", myWindow)
	updateAllStatuses()
}

// unpatchWithRosettaMethod implements the existing TurtleWoW unpatching method
func unpatchWithRosettaMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Removing Rosetta patching (legacy TurtleWoW method)")

	dialog.ShowInformation("Info", "Using legacy TurtleWoW unpatching method. Please use the existing TurtleWoW unpatch button.", myWindow)
	updateAllStatuses()
}

// CheckVersionPatchingStatus checks if a version is properly patched
func CheckVersionPatchingStatus(gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool) bool {
	if gamePath == "" {
		return false
	}

	if usesDivxDecoderPatch {
		// For non-TurtleSilicon versions, check for DivxDecoder.dll (winerosetta), d3d9.dll, and rosettax87 directory with verification
		divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")
		d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")
		rosettaX87Dir := filepath.Join(gamePath, "rosettax87")

		// Check that main files exist and rosettax87 directory exists
		if !utils.PathExists(divxDecoderPath) || !utils.PathExists(d3d9DllPath) || !utils.DirExists(rosettaX87Dir) {
			return false
		}

		// Verify rosettax87 binary files with hash/size verification
		rosettax87Path := filepath.Join(rosettaX87Dir, "rosettax87")
		libRuntimeRosettax87Path := filepath.Join(rosettaX87Dir, "libRuntimeRosettax87")

		rosettax87Valid := utils.PathExists(rosettax87Path) && utils.CompareFileWithBundledResource(rosettax87Path, "rosettax87/rosettax87")
		libRuntimeValid := utils.PathExists(libRuntimeRosettax87Path) && utils.CompareFileWithBundledResource(libRuntimeRosettax87Path, "rosettax87/libRuntimeRosettax87")

		return rosettax87Valid && libRuntimeValid
	}

	// For TurtleSilicon, check full rosettax87 patches (winerosetta.dll, d3d9.dll, etc.)
	winerosettaDll := filepath.Join(gamePath, "winerosetta.dll")
	d3d9Dll := filepath.Join(gamePath, "d3d9.dll")
	rosettaX87Dir := filepath.Join(gamePath, "rosettax87")

	// Check that main files exist and rosettax87 directory exists
	if !utils.PathExists(winerosettaDll) || !utils.PathExists(d3d9Dll) || !utils.DirExists(rosettaX87Dir) {
		return false
	}

	// Verify rosettax87 binary files with hash/size verification
	rosettax87Path := filepath.Join(rosettaX87Dir, "rosettax87")
	libRuntimeRosettax87Path := filepath.Join(rosettaX87Dir, "libRuntimeRosettax87")

	rosettax87Valid := utils.PathExists(rosettax87Path) && utils.CompareFileWithBundledResource(rosettax87Path, "rosettax87/rosettax87")
	libRuntimeValid := utils.PathExists(libRuntimeRosettax87Path) && utils.CompareFileWithBundledResource(libRuntimeRosettax87Path, "rosettax87/libRuntimeRosettax87")

	return rosettax87Valid && libRuntimeValid
}
