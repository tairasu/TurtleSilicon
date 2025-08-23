package patching

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// PatchVersionGame patches a game version based on its configuration
func PatchVersionGame(myWindow fyne.Window, updateAllStatuses func(), gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool, executableName string, versionID string) {
	debug.Printf("=== PATCHING DEBUG START ===")
	debug.Printf("Version ID: %s", versionID)
	debug.Printf("Game Path: %s", gamePath)
	debug.Printf("Executable: %s", executableName)
	debug.Printf("Uses Rosetta Patching: %v", usesRosettaPatching)
	debug.Printf("Uses DivX Decoder Patch: %v", usesDivxDecoderPatch)
	debug.Printf("=== PATCHING DEBUG END ===")

	if gamePath == "" {
		dialog.ShowError(fmt.Errorf("game path not set. Please set it first."), myWindow)
		return
	}

	if usesRosettaPatching {
		// TurtleSilicon uses the full rosettax87 patching (includes libSiliconPatch.dll)
		debug.Printf("Using full rosettax87 patching for %s (includes libSiliconPatch.dll)", versionID)
		// Temporarily set paths.TurtlewowPath so existing patching works
		originalPath := paths.TurtlewowPath
		paths.TurtlewowPath = gamePath
		defer func() {
			paths.TurtlewowPath = originalPath
		}()

		PatchTurtleWoW(myWindow, updateAllStatuses)
	} else if usesDivxDecoderPatch {
		// BurningSilicon and VanillaSilicon use the original DivX decoder approach
		if versionID == "burningsilicon" || versionID == "vanillasilicon" {
			debug.Printf("Using original DivX decoder method for %s", versionID)
			patchWithOriginalDivxDecoderMethod(myWindow, updateAllStatuses, gamePath)
		} else {
			// Other DivX versions use the new libDllLdr approach
			debug.Printf("Using libDllLdr method for %s (no libSiliconPatch.dll)", versionID)
			patchWithLibDllLdrMethod(myWindow, updateAllStatuses, gamePath, executableName, true) // Apply movie setting for DivX versions
		}
	} else {
		// Versions with both flags false (EpochSilicon, WrathSilicon) use libDllLdr approach
		debug.Printf("Using libDllLdr method for %s (no libSiliconPatch.dll)", versionID)
		patchWithLibDllLdrMethod(myWindow, updateAllStatuses, gamePath, executableName, false) // Don't apply movie setting
	}
}

// patchWithOriginalDivxDecoderMethod implements the original DivX decoder patching method for BurningSilicon
func patchWithOriginalDivxDecoderMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Applying original DivX decoder patching method")

	// Show progress popup and run entire patching process in background
	debug.Printf("Starting DivX decoder patching process")

	// Create and show progress popup
	progressPopup := createPatchingProgressPopup(myWindow)
	progressPopup.Show()

	// Run the patching process in a goroutine to keep UI responsive
	go func() {
		// Step 1: Create backup of existing DivxDecoder.dll if it exists
		divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")
		divxDecoderBackupPath := filepath.Join(gamePath, "DivxDecoder.dll.backup")

		if utils.PathExists(divxDecoderPath) {
			debug.Printf("Creating backup of existing DivxDecoder.dll at: %s", divxDecoderBackupPath)

			originalData, err := os.ReadFile(divxDecoderPath)
			if err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to read existing DivxDecoder.dll: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}

			if err := os.WriteFile(divxDecoderBackupPath, originalData, 0644); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create backup of DivxDecoder.dll: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
			debug.Printf("Successfully created backup of DivxDecoder.dll")

			if err := os.Remove(divxDecoderPath); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to remove existing DivxDecoder.dll: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
			debug.Printf("Successfully removed original DivxDecoder.dll")
		}

		// Step 2: Copy winerosetta.dll as DivxDecoder.dll
		debug.Printf("Copying winerosetta.dll as DivxDecoder.dll to: %s", divxDecoderPath)

		winerosettaResource, err := fyne.LoadResourceFromPath("winerosetta/winerosetta.dll")
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to open bundled winerosetta.dll resource: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		divxFile, err := os.Create(divxDecoderPath)
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to create DivxDecoder.dll file: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer divxFile.Close()

		_, err = io.Copy(divxFile, bytes.NewReader(winerosettaResource.Content()))
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to copy winerosetta.dll as DivxDecoder.dll: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Successfully copied winerosetta.dll as DivxDecoder.dll")

		// Step 3: Copy d3d9.dll
		d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")
		debug.Printf("Copying d3d9.dll to: %s", d3d9DllPath)

		d3d9Resource, err := fyne.LoadResourceFromPath("winerosetta/d3d9.dll")
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to open bundled d3d9.dll resource: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		d3d9File, err := os.Create(d3d9DllPath)
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to create d3d9.dll file: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer d3d9File.Close()

		_, err = io.Copy(d3d9File, bytes.NewReader(d3d9Resource.Content()))
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to copy d3d9.dll: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Successfully copied d3d9.dll")

		// Step 4: Copy rosettax87 service files (same as libDllLdr method)
		rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
		if !utils.DirExists(rosettaX87Dir) {
			if err := os.MkdirAll(rosettaX87Dir, 0755); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create rosettax87 directory: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
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
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}

			destinationFile, err := os.Create(destPath)
			if err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}

			_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
			if err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to copy bundled resource to %s: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					destinationFile.Close()
					updateAllStatuses()
				})
				return
			}
			destinationFile.Close()

			if err := os.Chmod(destPath, 0755); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to make %s executable: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
			debug.Printf("Successfully copied and made executable: %s to %s", resourceName, destPath)
		}

		// Step 9: Apply movie setting to Config.wtf for versions that use divx decoder patch
		if err := EnsureMovieSetting(gamePath); err != nil {
			debug.Printf("Warning: failed to apply movie setting to Config.wtf: %v", err)
		} else {
			debug.Printf("Successfully applied movie setting to Config.wtf")
		}

		// Success - update UI on main thread
		fyne.DoAndWait(func() {
			progressPopup.Hide()
			dialog.ShowInformation("Success", "Game patching completed successfully.", myWindow)
			updateAllStatuses()
		})
	}()
}

// patchWithLibDllLdrMethod implements the new libDllLdr.dll patching method for other versions
func patchWithLibDllLdrMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string, executableName string, applyMovieSetting bool) {
	debug.Println("Applying libDllLdr.dll patching method")

	// Show progress popup and run entire patching process in background
	debug.Printf("Starting libDllLdr.dll patching process")

	// Create and show progress popup
	progressPopup := createPatchingProgressPopup(myWindow)
	progressPopup.Show()

	// Run the patching process in a goroutine to keep UI responsive
	go func() {
		// Determine patched executable name based on the provided executable
		var patchedExecutableName string
		if executableName == "Ascension.exe" {
			patchedExecutableName = "Ascension_patched.exe"
			debug.Printf("Using Ascension.exe for EpochSilicon")
		} else {
			// Default to Wow.exe for all other versions
			executableName = "Wow.exe"
			patchedExecutableName = "Wow_patched.exe"
			debug.Printf("Using Wow.exe for standard WoW game")
		}

		// Step 1: Verify game directory exists and copy libDllLdr.dll from winerosetta directory to game path
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		debug.Printf("Copying libDllLdr.dll to: %s", libDllLdrPath)

		// First verify that the game directory exists
		if !utils.DirExists(gamePath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("Game directory does not exist: %s", gamePath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Game directory verified: %s", gamePath)

		// Test write permissions by creating a temporary file
		testFilePath := filepath.Join(gamePath, "test_write_permissions.tmp")
		if testFile, err := os.Create(testFilePath); err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("Cannot write to game directory (permission denied): %s\nError: %v", gamePath, err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		} else {
			testFile.Close()
			os.Remove(testFilePath) // Clean up test file
			debug.Printf("Write permissions verified for game directory")
		}

		libDllLdrResource, err := fyne.LoadResourceFromPath("winerosetta/libDllLdr.dll")
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to open bundled libDllLdr.dll resource: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		libDllLdrFile, err := os.Create(libDllLdrPath)
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to create libDllLdr.dll file at %s: %v\nGame directory: %s\nDirectory exists: %v", libDllLdrPath, err, gamePath, utils.DirExists(gamePath))
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer libDllLdrFile.Close()

		_, err = io.Copy(libDllLdrFile, bytes.NewReader(libDllLdrResource.Content()))
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to copy libDllLdr.dll: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Successfully copied libDllLdr.dll")

		// Step 2: Create mods directory if it doesn't exist and copy winerosetta.dll there
		modsDir := filepath.Join(gamePath, "mods")
		if !utils.DirExists(modsDir) {
			debug.Printf("Creating mods directory: %s", modsDir)
			if err := os.MkdirAll(modsDir, 0755); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create mods directory: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
		}

		winerosettaDllPath := filepath.Join(modsDir, "winerosetta.dll")
		debug.Printf("Copying winerosetta.dll to: %s", winerosettaDllPath)

		winerosettaResource, err := fyne.LoadResourceFromPath("winerosetta/winerosetta.dll")
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to open bundled winerosetta.dll resource: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		winerosettaFile, err := os.Create(winerosettaDllPath)
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to create winerosetta.dll file: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer winerosettaFile.Close()

		_, err = io.Copy(winerosettaFile, bytes.NewReader(winerosettaResource.Content()))
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to copy winerosetta.dll: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Successfully copied winerosetta.dll")

		// Step 3: Copy d3d9.dll from bundled resources to game path
		d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")
		debug.Printf("Copying d3d9.dll to: %s", d3d9DllPath)

		d3d9Resource, err := fyne.LoadResourceFromPath("winerosetta/d3d9.dll")
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to open bundled d3d9.dll resource: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		d3d9File, err := os.Create(d3d9DllPath)
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to create d3d9.dll file: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer d3d9File.Close()

		_, err = io.Copy(d3d9File, bytes.NewReader(d3d9Resource.Content()))
		if err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("failed to copy d3d9.dll: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Successfully copied d3d9.dll")

		// Step 4: Create/update dlls.txt with winerosetta.dll entry (moved inside goroutine)
		dllsPath := filepath.Join(gamePath, "dlls.txt")
		debug.Printf("Creating/updating dlls.txt at: %s", dllsPath)

		// Check if dlls.txt already contains mods/winerosetta.dll
		if !isDllRegisteredInDllsTxt(gamePath, "mods/winerosetta.dll") {
			// Read existing content or start with empty
			var existingContent string
			if utils.PathExists(dllsPath) {
				if content, err := os.ReadFile(dllsPath); err == nil {
					existingContent = string(content)
				}
			}

			// Ensure content ends with newline if it's not empty
			if existingContent != "" && !strings.HasSuffix(existingContent, "\n") {
				existingContent += "\n"
			}

			// Add mods/winerosetta.dll entry
			newContent := existingContent + "mods/winerosetta.dll\n"

			if err := os.WriteFile(dllsPath, []byte(newContent), 0644); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to write dlls.txt: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
			debug.Printf("Successfully updated dlls.txt with mods/winerosetta.dll entry")
		} else {
			debug.Printf("mods/winerosetta.dll already registered in dlls.txt")
		}

		// Step 5: Copy rosettax87 service files (required for the patching process)
		rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
		if !utils.DirExists(rosettaX87Dir) {
			if err := os.MkdirAll(rosettaX87Dir, 0755); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create rosettax87 directory: %v", err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
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
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}

			destinationFile, err := os.Create(destPath)
			if err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}

			_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
			if err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to copy bundled resource to %s: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					destinationFile.Close()
					updateAllStatuses()
				})
				return
			}
			destinationFile.Close()

			// Make rosettax87 executable
			if err := os.Chmod(destPath, 0755); err != nil {
				fyne.DoAndWait(func() {
					progressPopup.Hide()
					errMsg := fmt.Sprintf("failed to make %s executable: %v", destPath, err)
					dialog.ShowError(errors.New(errMsg), myWindow)
					debug.Println(errMsg)
					updateAllStatuses()
				})
				return
			}
			debug.Printf("Successfully copied and made executable: %s to %s", resourceName, destPath)
		}

		// Step 6: Verify libDllLdr.dll, winerosetta.dll, and d3d9.dll were created successfully before proceeding
		if !utils.PathExists(libDllLdrPath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("libDllLdr.dll was not created successfully at: %s", libDllLdrPath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Verified libDllLdr.dll exists at: %s", libDllLdrPath)

		if !utils.PathExists(winerosettaDllPath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("winerosetta.dll was not created successfully at: %s", winerosettaDllPath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Verified winerosetta.dll exists at: %s", winerosettaDllPath)

		if !utils.PathExists(d3d9DllPath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide()
				errMsg := fmt.Sprintf("d3d9.dll was not created successfully at: %s", d3d9DllPath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		debug.Printf("Verified d3d9.dll exists at: %s", d3d9DllPath)

		// Step 7: Check if patched executable already exists - skip rundll32 if it does
		patchedExePath := filepath.Join(gamePath, patchedExecutableName)
		if utils.PathExists(patchedExePath) {
			debug.Printf("Patched executable already exists, skipping rundll32 command: %s", patchedExePath)

			fyne.DoAndWait(func() {
				progressPopup.Hide()
				dialog.ShowInformation("Success", "Game patching completed successfully.", myWindow)
				updateAllStatuses()
			})
			return
		}
		// Step 8: Run wine rundll32 command to generate patched executable
		debug.Printf("Patched executable not found, generating it with wine rundll32")

		// Get CrossOver wineloader path (same directory as wineloader2, not the /bin/wine)
		crossoverPath := paths.CrossoverPath
		if crossoverPath == "" {
			fyne.DoAndWait(func() {
				progressPopup.Hide() // Hide progress popup on error
				errMsg := "CrossOver path not set. Cannot run wine command."
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		wineloaderPath := filepath.Join(crossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader")
		if !utils.PathExists(wineloaderPath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide() // Hide progress popup on error
				errMsg := fmt.Sprintf("Wine loader not found at: %s", wineloaderPath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		// Change to game directory and run the wine command
		originalDir, _ := os.Getwd()
		if err := os.Chdir(gamePath); err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide() // Hide progress popup on error
				errMsg := fmt.Sprintf("failed to change to game directory: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}
		defer os.Chdir(originalDir)

		// Run: wine rundll32 libDllLdr.dll,RunDll32Entry Wow.exe (or Ascension.exe)
		// Using wineloader (original wine) without any bottles
		cmd := []string{wineloaderPath, "rundll32", "libDllLdr.dll,RunDll32Entry", executableName}
		debug.Printf("Running command: %v", cmd)

		// Execute the command with environment variables to avoid bottles
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		execCmd.Dir = gamePath

		// Create a temporary wine prefix to avoid using bottles
		tempDir := filepath.Join(os.TempDir(), "turtlesilicon_wine_temp")
		os.RemoveAll(tempDir)       // Clean up any existing temp directory
		defer os.RemoveAll(tempDir) // Clean up after we're done

		// Set environment variables with temporary wine prefix
		execCmd.Env = append(os.Environ(),
			"WINEPREFIX="+tempDir, // Use temporary directory instead of bottles
			"WINEARCH=win64",      // Set architecture
			"WINEDLLOVERRIDES=",   // Clear any DLL overrides
		)

		if output, err := execCmd.CombinedOutput(); err != nil {
			fyne.DoAndWait(func() {
				progressPopup.Hide() // Hide progress popup on error
				errMsg := fmt.Sprintf("failed to run wine rundll32 command: %v\nOutput: %s", err, string(output))
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		} else {
			debug.Printf("Wine rundll32 command output: %s", string(output))
		}

		// Verify that the patched executable was created by rundll32
		if !utils.PathExists(patchedExePath) {
			fyne.DoAndWait(func() {
				progressPopup.Hide() // Hide progress popup on error
				errMsg := fmt.Sprintf("Patched executable was not created by rundll32: %s", patchedExePath)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				updateAllStatuses()
			})
			return
		}

		debug.Printf("Successfully created patched executable with rundll32: %s", patchedExecutableName)

		// Step 9: Apply movie setting to Config.wtf only for versions that use divx decoder patch
		if applyMovieSetting {
			if err := EnsureMovieSetting(gamePath); err != nil {
				debug.Printf("Warning: failed to apply movie setting to Config.wtf: %v", err)
			} else {
				debug.Printf("Successfully applied movie setting to Config.wtf")
			}
		} else {
			debug.Printf("Skipping movie setting for this version (not needed for non-DivX versions)")
		}

		// Success - update UI on main thread
		fyne.DoAndWait(func() {
			progressPopup.Hide() // Hide progress popup on success
			dialog.ShowInformation("Success", "Game patching completed successfully.", myWindow)
			updateAllStatuses()
		})
	}()
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
func UnpatchVersionGame(myWindow fyne.Window, updateAllStatuses func(), gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool, versionID string) {
	debug.Printf("Unpatching game at path: %s, rosetta=%v, divx=%v, version=%s", gamePath, usesRosettaPatching, usesDivxDecoderPatch, versionID)

	if gamePath == "" {
		dialog.ShowError(fmt.Errorf("game path not set. Please set it first."), myWindow)
		return
	}

	if usesRosettaPatching {
		// TurtleSilicon uses the full rosettax87 unpatching
		// Temporarily set paths.TurtlewowPath so existing unpatching works
		originalPath := paths.TurtlewowPath
		paths.TurtlewowPath = gamePath
		defer func() {
			paths.TurtlewowPath = originalPath
		}()

		UnpatchTurtleWoW(myWindow, updateAllStatuses)
	} else if usesDivxDecoderPatch {
		// For DivX decoder versions, determine which approach to unpatch
		// Check which files exist to determine the method used
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")

		if utils.PathExists(libDllLdrPath) {
			// libDllLdr approach was used
			unpatchWithLibDllLdrMethod(myWindow, updateAllStatuses, gamePath)
		} else if utils.PathExists(divxDecoderPath) {
			// Original DivX decoder approach was used
			unpatchWithOriginalDivxDecoderMethod(myWindow, updateAllStatuses, gamePath)
		} else {
			dialog.ShowInformation("Info", "No patches found to remove.", myWindow)
			updateAllStatuses()
		}
	} else {
		// Versions with both flags false (EpochSilicon, WrathSilicon) use libDllLdr approach
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		if utils.PathExists(libDllLdrPath) {
			unpatchWithLibDllLdrMethod(myWindow, updateAllStatuses, gamePath)
		} else {
			dialog.ShowInformation("Info", "No patches found to remove.", myWindow)
			updateAllStatuses()
		}
	}
}

// unpatchWithLibDllLdrMethod removes the libDllLdr.dll and patched executables
func unpatchWithLibDllLdrMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Removing libDllLdr.dll patching")

	// Remove libDllLdr.dll
	libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
	if utils.PathExists(libDllLdrPath) {
		debug.Printf("Removing libDllLdr.dll at: %s", libDllLdrPath)
		if err := os.Remove(libDllLdrPath); err != nil {
			errMsg := fmt.Sprintf("failed to remove libDllLdr.dll: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully removed libDllLdr.dll")
	}

	// Remove winerosetta.dll from mods directory
	winerosettaDllPath := filepath.Join(gamePath, "mods", "winerosetta.dll")
	if utils.PathExists(winerosettaDllPath) {
		debug.Printf("Removing winerosetta.dll at: %s", winerosettaDllPath)
		if err := os.Remove(winerosettaDllPath); err != nil {
			debug.Printf("Warning: failed to remove winerosetta.dll: %v", err)
			// Don't fail the operation for this
		} else {
			debug.Printf("Successfully removed winerosetta.dll")
		}
	}

	// Remove d3d9.dll
	d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")
	if utils.PathExists(d3d9DllPath) {
		debug.Printf("Removing d3d9.dll at: %s", d3d9DllPath)
		if err := os.Remove(d3d9DllPath); err != nil {
			debug.Printf("Warning: failed to remove d3d9.dll: %v", err)
			// Don't fail the operation for this
		} else {
			debug.Printf("Successfully removed d3d9.dll")
		}
	}

	// Remove patched executables - check for multiple possible names
	patchedExecutables := []string{
		"Wow_patched.exe",
		"Project-Epoch_patched.exe", // Legacy name
		"Ascension_patched.exe",     // New name for EpochSilicon
	}

	for _, execName := range patchedExecutables {
		patchedPath := filepath.Join(gamePath, execName)
		if utils.PathExists(patchedPath) {
			debug.Printf("Removing %s at: %s", execName, patchedPath)
			if err := os.Remove(patchedPath); err != nil {
				debug.Printf("Warning: failed to remove %s: %v", execName, err)
				// Don't fail the operation for this
			} else {
				debug.Printf("Successfully removed %s", execName)
			}
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

	// Remove winerosetta.dll entry from dlls.txt
	dllsPath := filepath.Join(gamePath, "dlls.txt")
	if utils.PathExists(dllsPath) {
		debug.Printf("Removing winerosetta.dll entry from dlls.txt")

		content, err := os.ReadFile(dllsPath)
		if err != nil {
			debug.Printf("Warning: failed to read dlls.txt: %v", err)
		} else {
			contentStr := string(content)

			// Remove "mods/winerosetta.dll\n" or "mods/winerosetta.dll" at end
			newContent := strings.ReplaceAll(contentStr, "mods/winerosetta.dll\n", "")
			newContent = strings.TrimSuffix(newContent, "mods/winerosetta.dll")

			if newContent != contentStr {
				if err := os.WriteFile(dllsPath, []byte(newContent), 0644); err != nil {
					debug.Printf("Warning: failed to update dlls.txt: %v", err)
				} else {
					debug.Printf("Successfully removed mods/winerosetta.dll entry from dlls.txt")
				}
			} else {
				debug.Printf("mods/winerosetta.dll entry not found in dlls.txt")
			}
		}
	}

	dialog.ShowInformation("Success", "Game unpatching completed successfully.", myWindow)
	updateAllStatuses()
}

// unpatchWithOriginalDivxDecoderMethod removes the original DivX decoder patching for BurningSilicon
func unpatchWithOriginalDivxDecoderMethod(myWindow fyne.Window, updateAllStatuses func(), gamePath string) {
	debug.Println("Removing original DivX decoder patching")

	// Remove DivxDecoder.dll and restore backup if it exists
	divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")
	divxDecoderBackupPath := filepath.Join(gamePath, "DivxDecoder.dll.backup")

	if utils.PathExists(divxDecoderPath) {
		debug.Printf("Removing DivxDecoder.dll at: %s", divxDecoderPath)
		if err := os.Remove(divxDecoderPath); err != nil {
			debug.Printf("Warning: failed to remove DivxDecoder.dll: %v", err)
			// Don't fail the operation for this
		} else {
			debug.Printf("Successfully removed DivxDecoder.dll")
		}
	}

	// Restore backup if it exists
	if utils.PathExists(divxDecoderBackupPath) {
		debug.Printf("Restoring backup DivxDecoder.dll from: %s", divxDecoderBackupPath)
		backupData, err := os.ReadFile(divxDecoderBackupPath)
		if err != nil {
			debug.Printf("Warning: failed to read backup DivxDecoder.dll: %v", err)
		} else {
			if err := os.WriteFile(divxDecoderPath, backupData, 0644); err != nil {
				debug.Printf("Warning: failed to restore backup DivxDecoder.dll: %v", err)
			} else {
				debug.Printf("Successfully restored backup DivxDecoder.dll")
				// Remove the backup file after successful restore
				if err := os.Remove(divxDecoderBackupPath); err != nil {
					debug.Printf("Warning: failed to remove backup file: %v", err)
				}
			}
		}
	}

	// Remove d3d9.dll
	d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")
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
func CheckVersionPatchingStatus(gamePath string, usesRosettaPatching bool, usesDivxDecoderPatch bool, versionID string) bool {
	if gamePath == "" {
		return false
	}

	// Add EpochSilicon executable migration logic
	if versionID == "epochsilicon" {
		migrateEpochSiliconExecutables(gamePath)
	}

	if usesRosettaPatching {
		// TurtleSilicon check (same as current "else" case)
		// For non-TurtleSilicon versions, determine which approach was used
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")

		if utils.PathExists(libDllLdrPath) {
			// libDllLdr approach - check for libDllLdr.dll, winerosetta.dll in mods/, d3d9.dll, and patched executables
			winerosettaDllPath := filepath.Join(gamePath, "mods", "winerosetta.dll")
			d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

			// Check that winerosetta.dll exists
			if !utils.PathExists(winerosettaDllPath) {
				debug.Printf("Patch verification failed: winerosetta.dll not found at %s", winerosettaDllPath)
				return false
			}

			// Check that d3d9.dll exists
			if !utils.PathExists(d3d9DllPath) {
				debug.Printf("Patch verification failed: d3d9.dll not found at %s", d3d9DllPath)
				return false
			}

			// Check for patched executable based on version
			var patchedFound bool
			if versionID == "epochsilicon" {
				// For EpochSilicon, look for Ascension_patched.exe
				ascensionPatchedPath := filepath.Join(gamePath, "Ascension_patched.exe")
				patchedFound = utils.PathExists(ascensionPatchedPath)
				if !patchedFound {
					debug.Printf("Patch verification failed: Ascension_patched.exe not found for EpochSilicon")
				}
			} else {
				// For other versions, look for Wow_patched.exe
				wowPatchedPath := filepath.Join(gamePath, "Wow_patched.exe")
				patchedFound = utils.PathExists(wowPatchedPath)
				if !patchedFound {
					debug.Printf("Patch verification failed: Wow_patched.exe not found")
				}
			}

			if !patchedFound {
				return false
			}

			// Check dlls.txt to ensure mods/winerosetta.dll is properly registered
			if !isDllRegisteredInDllsTxt(gamePath, "mods/winerosetta.dll") {
				debug.Printf("Patch verification failed: mods/winerosetta.dll not found in dlls.txt for %s", gamePath)
				return false
			}

			debug.Printf("✓ libDllLdr.dll patch verification passed for %s", gamePath)
		} else if utils.PathExists(divxDecoderPath) {
			// Original DivX decoder approach - check for DivxDecoder.dll and d3d9.dll
			d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

			// Check that d3d9.dll exists
			if !utils.PathExists(d3d9DllPath) {
				debug.Printf("Patch verification failed: d3d9.dll not found at %s", d3d9DllPath)
				return false
			}

			debug.Printf("✓ Original DivX decoder patch verification passed for %s", gamePath)
		} else {
			// No patches found
			debug.Printf("Patch verification failed: no patching method detected for %s", gamePath)
			return false
		}

		// Check for rosettax87 directory and files (common to both approaches)
		rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
		if !utils.DirExists(rosettaX87Dir) {
			debug.Printf("Patch verification failed: rosettax87 directory not found at %s", rosettaX87Dir)
			return false
		}

		// Verify rosettax87 binary files with hash/size verification
		rosettax87Path := filepath.Join(rosettaX87Dir, "rosettax87")
		libRuntimeRosettax87Path := filepath.Join(rosettaX87Dir, "libRuntimeRosettax87")

		rosettax87Valid := utils.PathExists(rosettax87Path) && utils.CompareFileWithBundledResource(rosettax87Path, "rosettax87/rosettax87")
		libRuntimeValid := utils.PathExists(libRuntimeRosettax87Path) && utils.CompareFileWithBundledResource(libRuntimeRosettax87Path, "rosettax87/libRuntimeRosettax87")

		if !rosettax87Valid {
			debug.Printf("Patch verification failed: rosettax87 binary invalid or missing at %s", rosettax87Path)
			return false
		}

		if !libRuntimeValid {
			debug.Printf("Patch verification failed: libRuntimeRosettax87 invalid or missing at %s", libRuntimeRosettax87Path)
			return false
		}

		// For versions that use divx decoder patch, also check that movie setting is applied
		if usesDivxDecoderPatch {
			if !CheckMovieSetting(gamePath) {
				debug.Printf("Patch verification failed: movie setting not applied in Config.wtf for %s", gamePath)
				return false
			}
		}

		return true
	} else if usesDivxDecoderPatch {
		// Handle versions that use DivX decoder patch (VanillaSilicon, BurningSilicon)
		// Determine which approach was used based on what files exist
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		divxDecoderPath := filepath.Join(gamePath, "DivxDecoder.dll")

		if utils.PathExists(libDllLdrPath) {
			// Uses libDllLdr approach - check for libDllLdr.dll, winerosetta.dll in mods/, d3d9.dll
			winerosettaDllPath := filepath.Join(gamePath, "mods", "winerosetta.dll")
			d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

			if !utils.PathExists(winerosettaDllPath) {
				debug.Printf("Patch verification failed: winerosetta.dll not found at %s", winerosettaDllPath)
				return false
			}

			if !utils.PathExists(d3d9DllPath) {
				debug.Printf("Patch verification failed: d3d9.dll not found at %s", d3d9DllPath)
				return false
			}

			// Check dlls.txt registration
			if !isDllRegisteredInDllsTxt(gamePath, "mods/winerosetta.dll") {
				debug.Printf("Patch verification failed: mods/winerosetta.dll not found in dlls.txt for %s", gamePath)
				return false
			}

			// Check for Wow_patched.exe
			wowPatchedPath := filepath.Join(gamePath, "Wow_patched.exe")
			if !utils.PathExists(wowPatchedPath) {
				debug.Printf("Patch verification failed: Wow_patched.exe not found")
				return false
			}

			debug.Printf("✓ libDllLdr DivX patch verification passed for %s", gamePath)
		} else if utils.PathExists(divxDecoderPath) {
			// Uses original DivX decoder approach - check for DivxDecoder.dll and d3d9.dll
			d3d9DllPath := filepath.Join(gamePath, "d3d9.dll")

			if !utils.PathExists(d3d9DllPath) {
				debug.Printf("Patch verification failed: d3d9.dll not found at %s", d3d9DllPath)
				return false
			}

			debug.Printf("✓ Original DivX decoder patch verification passed for %s", gamePath)
		} else {
			// No patches found
			debug.Printf("Patch verification failed: no DivX decoder patches detected for %s", gamePath)
			return false
		}

		// Check for rosettax87 directory and files (common to both approaches)
		rosettaX87Dir := filepath.Join(gamePath, "rosettax87")
		if !utils.DirExists(rosettaX87Dir) {
			debug.Printf("Patch verification failed: rosettax87 directory not found at %s", rosettaX87Dir)
			return false
		}

		// Verify rosettax87 binary files with hash/size verification
		rosettax87Path := filepath.Join(rosettaX87Dir, "rosettax87")
		libRuntimeRosettax87Path := filepath.Join(rosettaX87Dir, "libRuntimeRosettax87")

		rosettax87Valid := utils.PathExists(rosettax87Path) && utils.CompareFileWithBundledResource(rosettax87Path, "rosettax87/rosettax87")
		libRuntimeValid := utils.PathExists(libRuntimeRosettax87Path) && utils.CompareFileWithBundledResource(libRuntimeRosettax87Path, "rosettax87/libRuntimeRosettax87")

		if !rosettax87Valid {
			debug.Printf("Patch verification failed: rosettax87 binary invalid or missing at %s", rosettax87Path)
			return false
		}

		if !libRuntimeValid {
			debug.Printf("Patch verification failed: libRuntimeRosettax87 invalid or missing at %s", libRuntimeRosettax87Path)
			return false
		}

		// For versions that use divx decoder patch, also check that movie setting is applied
		if !CheckMovieSetting(gamePath) {
			debug.Printf("Patch verification failed: movie setting not applied in Config.wtf for %s", gamePath)
			return false
		}

		debug.Printf("✓ DivX decoder patch verification completed for %s", gamePath)
		return true
	}

	// For TurtleSilicon, check full rosettax87 patches (winerosetta.dll in mods/, d3d9.dll in root, etc.)
	winerosettaDll := filepath.Join(gamePath, "mods", "winerosetta.dll")
	d3d9Dll := filepath.Join(gamePath, "d3d9.dll")
	rosettaX87Dir := filepath.Join(gamePath, "rosettax87")

	// Check that main files exist and rosettax87 directory exists
	if !utils.PathExists(winerosettaDll) || !utils.PathExists(d3d9Dll) || !utils.DirExists(rosettaX87Dir) {
		return false
	}

	// Check dlls.txt to ensure winerosetta.dll is properly registered (check for both old and new format)
	if !isDllRegisteredInDllsTxt(gamePath, "mods/winerosetta.dll") && !isDllRegisteredInDllsTxt(gamePath, "winerosetta.dll") {
		debug.Printf("Patch verification failed: winerosetta.dll not found in dlls.txt for %s", gamePath)
		return false
	}
	debug.Printf("✓ dlls.txt verification passed for %s", gamePath)

	// Verify rosettax87 binary files with hash/size verification
	rosettax87Path := filepath.Join(rosettaX87Dir, "rosettax87")
	libRuntimeRosettax87Path := filepath.Join(rosettaX87Dir, "libRuntimeRosettax87")

	rosettax87Valid := utils.PathExists(rosettax87Path) && utils.CompareFileWithBundledResource(rosettax87Path, "rosettax87/rosettax87")
	libRuntimeValid := utils.PathExists(libRuntimeRosettax87Path) && utils.CompareFileWithBundledResource(libRuntimeRosettax87Path, "rosettax87/libRuntimeRosettax87")

	return rosettax87Valid && libRuntimeValid
}

// isDllRegisteredInDllsTxt checks if a specific DLL is registered in dlls.txt
func isDllRegisteredInDllsTxt(gamePath string, dllName string) bool {
	dllsTextFile := filepath.Join(gamePath, "dlls.txt")

	// If dlls.txt doesn't exist, consider it as not registered
	if !utils.PathExists(dllsTextFile) {
		debug.Printf("dlls.txt not found at: %s", dllsTextFile)
		return false
	}

	content, err := os.ReadFile(dllsTextFile)
	if err != nil {
		debug.Printf("Failed to read dlls.txt: %v", err)
		return false
	}

	contentStr := string(content)

	// Check if the DLL is registered (look for "winerosetta.dll" on its own line)
	// The format is just "winerosetta.dll\n", not "winerosetta.dll=native"
	dllEntry := dllName + "\n"
	found := strings.Contains(contentStr, dllEntry)

	if !found {
		// Also check if it's at the end without a newline
		found = strings.HasSuffix(contentStr, dllName)
	}

	if !found {
		debug.Printf("'%s' not found in dlls.txt", dllName)
	}

	return found
}

// migrateEpochSiliconExecutables handles migration from old Project-Epoch executable names to new Ascension names
func migrateEpochSiliconExecutables(gamePath string) {
	debug.Printf("Checking for EpochSilicon executable migration at: %s", gamePath)

	// Check and migrate Project-Epoch.exe to Ascension.exe
	oldExePath := filepath.Join(gamePath, "Project-Epoch.exe")
	newExePath := filepath.Join(gamePath, "Ascension.exe")

	if utils.PathExists(oldExePath) && !utils.PathExists(newExePath) {
		debug.Printf("Migrating Project-Epoch.exe to Ascension.exe")
		if err := os.Rename(oldExePath, newExePath); err != nil {
			debug.Printf("Warning: failed to rename Project-Epoch.exe to Ascension.exe: %v", err)
		} else {
			debug.Printf("Successfully migrated Project-Epoch.exe to Ascension.exe")
		}
	}

	// Check and migrate Project-Epoch_patched.exe to Ascension_patched.exe
	oldPatchedExePath := filepath.Join(gamePath, "Project-Epoch_patched.exe")
	newPatchedExePath := filepath.Join(gamePath, "Ascension_patched.exe")

	if utils.PathExists(oldPatchedExePath) && !utils.PathExists(newPatchedExePath) {
		debug.Printf("Migrating Project-Epoch_patched.exe to Ascension_patched.exe")
		if err := os.Rename(oldPatchedExePath, newPatchedExePath); err != nil {
			debug.Printf("Warning: failed to rename Project-Epoch_patched.exe to Ascension_patched.exe: %v", err)
		} else {
			debug.Printf("Successfully migrated Project-Epoch_patched.exe to Ascension_patched.exe")
		}
	}
}

// createPatchingProgressPopup creates a modal popup to show patching progress
func createPatchingProgressPopup(myWindow fyne.Window) *widget.PopUp {
	// Create progress message
	titleLabel := widget.NewLabel("Patching Game")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	messageLabel := widget.NewLabel("Please wait while the game is being patched...")
	messageLabel.Wrapping = fyne.TextWrapWord

	// Create a progress bar (indeterminate)
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	// Create content container
	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		messageLabel,
		widget.NewSeparator(),
		progressBar,
	)

	// Create the popup
	popup := widget.NewModalPopUp(
		container.NewPadded(content),
		myWindow.Canvas(),
	)

	// Set popup size
	popup.Resize(fyne.NewSize(300, 150))

	return popup
}
