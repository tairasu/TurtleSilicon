package patching

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func PatchTurtleWoW(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Patch TurtleWoW clicked")
	if paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it first."), myWindow)
		return
	}

	targetWinerosettaDll := filepath.Join(paths.TurtlewowPath, "winerosetta.dll")
	targetD3d9Dll := filepath.Join(paths.TurtlewowPath, "d3d9.dll")
	targetLibSiliconPatchDll := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
	targetRosettaX87Dir := filepath.Join(paths.TurtlewowPath, "rosettax87")
	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")
	filesToCopy := map[string]string{
		"winerosetta/winerosetta.dll":     targetWinerosettaDll,
		"winerosetta/d3d9.dll":            targetD3d9Dll,
		"winerosetta/libSiliconPatch.dll": targetLibSiliconPatchDll,
	}

	for resourceName, destPath := range filesToCopy {
		log.Printf("Processing resource: %s to %s", resourceName, destPath)

		// Check if file already exists and has correct size
		if utils.PathExists(destPath) && utils.CompareFileWithBundledResource(destPath, resourceName) {
			log.Printf("File %s already exists with correct size, skipping copy", destPath)
			continue
		}

		if utils.PathExists(destPath) {
			log.Printf("File %s exists but has incorrect size, updating...", destPath)
		} else {
			log.Printf("File %s does not exist, creating...", destPath)
		}

		resource, err := fyne.LoadResourceFromPath(resourceName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		destinationFile, err := os.Create(destPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
		if err != nil {
			errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		log.Printf("Successfully copied %s to %s", resourceName, destPath)
	}

	log.Printf("Preparing rosettax87 directory at: %s", targetRosettaX87Dir)
	if err := os.RemoveAll(targetRosettaX87Dir); err != nil {
		log.Printf("Warning: could not remove existing rosettax87 folder '%s': %v", targetRosettaX87Dir, err)
	}
	if err := os.MkdirAll(targetRosettaX87Dir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create directory %s: %v", targetRosettaX87Dir, err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		log.Println(errMsg)
		paths.PatchesAppliedTurtleWoW = false
		updateAllStatuses()
		return
	}

	rosettaFilesToCopy := map[string]string{
		"rosettax87/rosettax87":           filepath.Join(targetRosettaX87Dir, "rosettax87"),
		"rosettax87/libRuntimeRosettax87": filepath.Join(targetRosettaX87Dir, "libRuntimeRosettax87"),
	}

	for resourceName, destPath := range rosettaFilesToCopy {
		log.Printf("Processing rosetta resource: %s to %s", resourceName, destPath)
		resource, err := fyne.LoadResourceFromPath(resourceName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		destinationFile, err := os.Create(destPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
		if err != nil {
			destinationFile.Close()
			errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		destinationFile.Close()

		if filepath.Base(destPath) == "rosettax87" {
			log.Printf("Setting execute permission for %s", destPath)
			if err := os.Chmod(destPath, 0755); err != nil {
				errMsg := fmt.Sprintf("failed to set execute permission for %s: %v", destPath, err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				log.Println(errMsg)
				paths.PatchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
		}
		log.Printf("Successfully copied %s to %s", resourceName, destPath)
	}

	log.Printf("Checking dlls.txt file at: %s", dllsTextFile)
	winerosettaEntry := "winerosetta.dll"
	libSiliconPatchEntry := "libSiliconPatch.dll"
	needsWinerosettaUpdate := true
	needsLibSiliconPatchUpdate := true

	if fileContentBytes, err := os.ReadFile(dllsTextFile); err == nil {
		fileContent := string(fileContentBytes)
		if strings.Contains(fileContent, winerosettaEntry) {
			log.Printf("dlls.txt already contains %s", winerosettaEntry)
			needsWinerosettaUpdate = false
		}
		if strings.Contains(fileContent, libSiliconPatchEntry) {
			log.Printf("dlls.txt already contains %s", libSiliconPatchEntry)
			needsLibSiliconPatchUpdate = false
		}
	} else {
		log.Printf("dlls.txt not found, will create a new one with both entries")
	}

	if needsWinerosettaUpdate || needsLibSiliconPatchUpdate {
		var fileContentBytes []byte
		var err error
		if utils.PathExists(dllsTextFile) {
			fileContentBytes, err = os.ReadFile(dllsTextFile)
			if err != nil {
				errMsg := fmt.Sprintf("failed to read dlls.txt for update: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				log.Println(errMsg)
			}
		}

		currentContent := string(fileContentBytes)
		updatedContent := currentContent

		if len(updatedContent) > 0 && !strings.HasSuffix(updatedContent, "\n") {
			updatedContent += "\n"
		}

		if needsWinerosettaUpdate {
			if !strings.Contains(updatedContent, winerosettaEntry+"\n") {
				updatedContent += winerosettaEntry + "\n"
				log.Printf("Adding %s to dlls.txt", winerosettaEntry)
			}
		}
		if needsLibSiliconPatchUpdate {
			if !strings.Contains(updatedContent, libSiliconPatchEntry+"\n") {
				updatedContent += libSiliconPatchEntry + "\n"
				log.Printf("Adding %s to dlls.txt", libSiliconPatchEntry)
			}
		}

		if err := os.WriteFile(dllsTextFile, []byte(updatedContent), 0644); err != nil {
			errMsg := fmt.Sprintf("failed to update dlls.txt: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
		} else {
			log.Printf("Successfully updated dlls.txt")
		}
	}

	log.Println("TurtleWoW patching with bundled resources completed successfully.")
	dialog.ShowInformation("Success", "TurtleWoW patching process completed using bundled resources.", myWindow)
	updateAllStatuses()
}

func PatchCrossOver(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Patch CrossOver clicked")
	if paths.CrossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it first."), myWindow)
		return
	}

	wineloaderBasePath := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application")
	wineloaderOrig := filepath.Join(wineloaderBasePath, "wineloader")
	wineloaderCopy := filepath.Join(wineloaderBasePath, "wineloader2")

	if !utils.PathExists(wineloaderOrig) {
		dialog.ShowError(fmt.Errorf("original wineloader not found at %s", wineloaderOrig), myWindow)
		paths.PatchesAppliedCrossOver = false
		updateAllStatuses()
		return
	}

	log.Printf("Copying %s to %s", wineloaderOrig, wineloaderCopy)
	if err := utils.CopyFile(wineloaderOrig, wineloaderCopy); err != nil {
		dialog.ShowError(fmt.Errorf("failed to copy wineloader: %w", err), myWindow)
		paths.PatchesAppliedCrossOver = false
		updateAllStatuses()
		return
	}

	log.Printf("Executing: codesign --remove-signature %s", wineloaderCopy)
	cmd := exec.Command("codesign", "--remove-signature", wineloaderCopy)
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		derrMsg := fmt.Sprintf("failed to remove signature from %s: %v\nOutput: %s", wineloaderCopy, err, string(combinedOutput))
		dialog.ShowError(errors.New(derrMsg), myWindow)
		log.Println(derrMsg)
		paths.PatchesAppliedCrossOver = false
		if err := os.Remove(wineloaderCopy); err != nil {
			log.Printf("Warning: failed to cleanup wineloader2 after codesign failure: %v", err)
		}
		updateAllStatuses()
		return
	}
	log.Printf("codesign output: %s", string(combinedOutput))

	log.Printf("Setting execute permissions for %s", wineloaderCopy)
	if err := os.Chmod(wineloaderCopy, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to set executable permissions for %s: %v", wineloaderCopy, err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		log.Println(errMsg)
		paths.PatchesAppliedCrossOver = false
		updateAllStatuses()
		return
	}

	log.Println("CrossOver patching completed successfully.")
	paths.PatchesAppliedCrossOver = true
	dialog.ShowInformation("Success", "CrossOver patching process completed.", myWindow)
	updateAllStatuses()
}

func UnpatchTurtleWoW(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Unpatch TurtleWoW clicked")
	if paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it first."), myWindow)
		return
	}

	// Files to remove
	winerosettaDllPath := filepath.Join(paths.TurtlewowPath, "winerosetta.dll")
	d3d9DllPath := filepath.Join(paths.TurtlewowPath, "d3d9.dll")
	libSiliconPatchDllPath := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
	rosettaX87DirPath := filepath.Join(paths.TurtlewowPath, "rosettax87")
	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")

	// Remove the rosettaX87 directory
	if utils.DirExists(rosettaX87DirPath) {
		log.Printf("Removing directory: %s", rosettaX87DirPath)
		if err := os.RemoveAll(rosettaX87DirPath); err != nil {
			errMsg := fmt.Sprintf("failed to remove directory %s: %v", rosettaX87DirPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
		} else {
			log.Printf("Successfully removed directory: %s", rosettaX87DirPath)
		}
	}

	// Remove DLL files
	filesToRemove := []string{winerosettaDllPath, d3d9DllPath, libSiliconPatchDllPath}
	for _, file := range filesToRemove {
		if utils.PathExists(file) {
			log.Printf("Removing file: %s", file)
			if err := os.Remove(file); err != nil {
				errMsg := fmt.Sprintf("failed to remove file %s: %v", file, err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				log.Println(errMsg)
			} else {
				log.Printf("Successfully removed file: %s", file)
			}
		}
	}

	// Update dlls.txt file - remove winerosetta.dll and libSiliconPatch.dll entries
	if utils.PathExists(dllsTextFile) {
		log.Printf("Updating dlls.txt file: %s", dllsTextFile)
		content, err := os.ReadFile(dllsTextFile)
		if err != nil {
			errMsg := fmt.Sprintf("failed to read dlls.txt file: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
		} else {
			lines := strings.Split(string(content), "\n")
			filteredLines := make([]string, 0, len(lines))

			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine != "winerosetta.dll" && trimmedLine != "libSiliconPatch.dll" {
					filteredLines = append(filteredLines, line)
				}
			}

			updatedContent := strings.Join(filteredLines, "\n")
			if err := os.WriteFile(dllsTextFile, []byte(updatedContent), 0644); err != nil {
				errMsg := fmt.Sprintf("failed to update dlls.txt file: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				log.Println(errMsg)
			} else {
				log.Printf("Successfully updated dlls.txt file")
			}
		}
	}

	log.Println("TurtleWoW unpatching completed successfully.")
	paths.PatchesAppliedTurtleWoW = false
	dialog.ShowInformation("Success", "TurtleWoW unpatching process completed.", myWindow)
	updateAllStatuses()
}

func UnpatchCrossOver(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Unpatch CrossOver clicked")
	if paths.CrossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it first."), myWindow)
		return
	}

	wineloaderCopy := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")

	if utils.PathExists(wineloaderCopy) {
		log.Printf("Removing file: %s", wineloaderCopy)
		if err := os.Remove(wineloaderCopy); err != nil {
			errMsg := fmt.Sprintf("failed to remove file %s: %v", wineloaderCopy, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			log.Println(errMsg)
			updateAllStatuses()
			return
		} else {
			log.Printf("Successfully removed file: %s", wineloaderCopy)
		}
	} else {
		log.Printf("File not found to remove: %s", wineloaderCopy)
	}

	log.Println("CrossOver unpatching completed successfully.")
	paths.PatchesAppliedCrossOver = false
	dialog.ShowInformation("Success", "CrossOver unpatching process completed.", myWindow)
	updateAllStatuses()
}
