package patching

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths" // Corrected import path
	"turtlesilicon/pkg/utils" // Corrected import path

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func PatchTurtleWoW(myWindow fyne.Window, updateAllStatuses func()) {
	debug.Println("Patch TurtleWoW clicked")
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
		debug.Printf("Processing resource: %s to %s", resourceName, destPath)

		// Check if file already exists and has correct size
		if utils.PathExists(destPath) && utils.CompareFileWithBundledResource(destPath, resourceName) {
			debug.Printf("File %s already exists with correct size, skipping copy", destPath)
			continue
		}

		if utils.PathExists(destPath) {
			debug.Printf("File %s exists but has incorrect size, updating...", destPath)
		} else {
			debug.Printf("File %s does not exist, creating...", destPath)
		}

		resource, err := fyne.LoadResourceFromPath(resourceName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		destinationFile, err := os.Create(destPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
		if err != nil {
			errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		debug.Printf("Successfully copied %s to %s", resourceName, destPath)
	}

	debug.Printf("Preparing rosettax87 directory at: %s", targetRosettaX87Dir)
	if err := os.RemoveAll(targetRosettaX87Dir); err != nil {
		debug.Printf("Warning: could not remove existing rosettax87 folder '%s': %v", targetRosettaX87Dir, err)
	}
	if err := os.MkdirAll(targetRosettaX87Dir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create directory %s: %v", targetRosettaX87Dir, err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		paths.PatchesAppliedTurtleWoW = false
		updateAllStatuses()
		return
	}

	rosettaFilesToCopy := map[string]string{
		"rosettax87/rosettax87":           filepath.Join(targetRosettaX87Dir, "rosettax87"),
		"rosettax87/libRuntimeRosettax87": filepath.Join(targetRosettaX87Dir, "libRuntimeRosettax87"),
	}

	for resourceName, destPath := range rosettaFilesToCopy {
		debug.Printf("Processing rosetta resource: %s to %s", resourceName, destPath)
		resource, err := fyne.LoadResourceFromPath(resourceName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to open bundled resource %s: %v", resourceName, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		destinationFile, err := os.Create(destPath)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create destination file %s: %v", destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}

		_, err = io.Copy(destinationFile, bytes.NewReader(resource.Content()))
		if err != nil {
			destinationFile.Close()
			errMsg := fmt.Sprintf("failed to copy bundled resource %s to %s: %v", resourceName, destPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			paths.PatchesAppliedTurtleWoW = false
			updateAllStatuses()
			return
		}
		destinationFile.Close()

		if filepath.Base(destPath) == "rosettax87" {
			debug.Printf("Setting execute permission for %s", destPath)
			if err := os.Chmod(destPath, 0755); err != nil {
				errMsg := fmt.Sprintf("failed to set execute permission for %s: %v", destPath, err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
				paths.PatchesAppliedTurtleWoW = false
				updateAllStatuses()
				return
			}
		}
		debug.Printf("Successfully copied %s to %s", resourceName, destPath)
	}

	debug.Printf("Checking dlls.txt file at: %s", dllsTextFile)
	winerosettaEntry := "winerosetta.dll"
	libSiliconPatchEntry := "libSiliconPatch.dll"
	needsWinerosettaUpdate := true
	needsLibSiliconPatchUpdate := true

	if fileContentBytes, err := os.ReadFile(dllsTextFile); err == nil {
		fileContent := string(fileContentBytes)
		if strings.Contains(fileContent, winerosettaEntry) {
			debug.Printf("dlls.txt already contains %s", winerosettaEntry)
			needsWinerosettaUpdate = false
		}
		if strings.Contains(fileContent, libSiliconPatchEntry) {
			debug.Printf("dlls.txt already contains %s", libSiliconPatchEntry)
			needsLibSiliconPatchUpdate = false
		}
	} else {
		debug.Printf("dlls.txt not found, will create a new one with both entries")
	}

	if needsWinerosettaUpdate || needsLibSiliconPatchUpdate {
		var fileContentBytes []byte
		var err error
		if utils.PathExists(dllsTextFile) {
			fileContentBytes, err = os.ReadFile(dllsTextFile)
			if err != nil {
				errMsg := fmt.Sprintf("failed to read dlls.txt for update: %v", err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
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
				debug.Printf("Adding %s to dlls.txt", winerosettaEntry)
			}
		}
		if needsLibSiliconPatchUpdate {
			if !strings.Contains(updatedContent, libSiliconPatchEntry+"\n") {
				updatedContent += libSiliconPatchEntry + "\n"
				debug.Printf("Adding %s to dlls.txt", libSiliconPatchEntry)
			}
		}

		if err := os.WriteFile(dllsTextFile, []byte(updatedContent), 0644); err != nil {
			errMsg := fmt.Sprintf("failed to update dlls.txt: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
		} else {
			debug.Printf("Successfully updated dlls.txt")
		}
	}

	// Apply shadowLOD setting to Config.wtf for FPS optimization
	if err := applyShadowLODSetting(); err != nil {
		debug.Printf("Warning: failed to apply shadowLOD setting to Config.wtf: %v", err)
		// Continue with patching even if Config.wtf update fails
	}

	debug.Println("TurtleWoW patching with bundled resources completed successfully.")
	dialog.ShowInformation("Success", "TurtleWoW patching process completed using bundled resources.", myWindow)
	updateAllStatuses()
}

func PatchCrossOver(myWindow fyne.Window, updateAllStatuses func()) {
	debug.Println("Patch CrossOver clicked")
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

	debug.Printf("Copying %s to %s", wineloaderOrig, wineloaderCopy)
	if err := utils.CopyFile(wineloaderOrig, wineloaderCopy); err != nil {
		errMsg := fmt.Sprintf("failed to copy wineloader: %v", err)
		if strings.Contains(err.Error(), "operation not permitted") {
			errMsg += "\n\nSolution: Open System Settings, go to Privacy & Security > App Management, and enable TurtleSilicon."
		}
		dialog.ShowError(fmt.Errorf(errMsg), myWindow)
		paths.PatchesAppliedCrossOver = false
		updateAllStatuses()
		return
	}

	debug.Printf("Executing: codesign --remove-signature %s", wineloaderCopy)
	cmd := exec.Command("codesign", "--remove-signature", wineloaderCopy)
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		derrMsg := fmt.Sprintf("failed to remove signature from %s: %v\nOutput: %s", wineloaderCopy, err, string(combinedOutput))
		dialog.ShowError(errors.New(derrMsg), myWindow)
		debug.Println(derrMsg)
		paths.PatchesAppliedCrossOver = false
		if err := os.Remove(wineloaderCopy); err != nil {
			debug.Printf("Warning: failed to cleanup wineloader2 after codesign failure: %v", err)
		}
		updateAllStatuses()
		return
	}
	debug.Printf("codesign output: %s", string(combinedOutput))

	debug.Printf("Setting execute permissions for %s", wineloaderCopy)
	if err := os.Chmod(wineloaderCopy, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to set executable permissions for %s: %v", wineloaderCopy, err)
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		paths.PatchesAppliedCrossOver = false
		updateAllStatuses()
		return
	}

	debug.Println("CrossOver patching completed successfully.")
	paths.PatchesAppliedCrossOver = true
	dialog.ShowInformation("Success", "CrossOver patching process completed.", myWindow)
	updateAllStatuses()
}

func UnpatchTurtleWoW(myWindow fyne.Window, updateAllStatuses func()) {
	debug.Println("Unpatch TurtleWoW clicked")
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
		debug.Printf("Removing directory: %s", rosettaX87DirPath)
		if err := os.RemoveAll(rosettaX87DirPath); err != nil {
			errMsg := fmt.Sprintf("failed to remove directory %s: %v", rosettaX87DirPath, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
		} else {
			debug.Printf("Successfully removed directory: %s", rosettaX87DirPath)
		}
	}

	// Remove DLL files
	filesToRemove := []string{winerosettaDllPath, d3d9DllPath, libSiliconPatchDllPath}
	for _, file := range filesToRemove {
		if utils.PathExists(file) {
			debug.Printf("Removing file: %s", file)
			if err := os.Remove(file); err != nil {
				errMsg := fmt.Sprintf("failed to remove file %s: %v", file, err)
				dialog.ShowError(errors.New(errMsg), myWindow)
				debug.Println(errMsg)
			} else {
				debug.Printf("Successfully removed file: %s", file)
			}
		}
	}

	// Update dlls.txt file - remove winerosetta.dll and libSiliconPatch.dll entries
	if utils.PathExists(dllsTextFile) {
		debug.Printf("Updating dlls.txt file: %s", dllsTextFile)
		content, err := os.ReadFile(dllsTextFile)
		if err != nil {
			errMsg := fmt.Sprintf("failed to read dlls.txt file: %v", err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
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
				debug.Println(errMsg)
			} else {
				debug.Printf("Successfully updated dlls.txt file")
			}
		}
	}

	// Remove shadowLOD setting from Config.wtf
	if err := removeShadowLODSetting(); err != nil {
		debug.Printf("Warning: failed to remove shadowLOD setting from Config.wtf: %v", err)
		// Continue with unpatching even if Config.wtf update fails
	}

	debug.Println("TurtleWoW unpatching completed successfully.")
	paths.PatchesAppliedTurtleWoW = false
	dialog.ShowInformation("Success", "TurtleWoW unpatching process completed.", myWindow)
	updateAllStatuses()
}

func UnpatchCrossOver(myWindow fyne.Window, updateAllStatuses func()) {
	debug.Println("Unpatch CrossOver clicked")
	if paths.CrossoverPath == "" {
		dialog.ShowError(fmt.Errorf("CrossOver path not set. Please set it first."), myWindow)
		return
	}

	wineloaderCopy := filepath.Join(paths.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "CrossOver-Hosted Application", "wineloader2")

	if utils.PathExists(wineloaderCopy) {
		debug.Printf("Removing file: %s", wineloaderCopy)
		if err := os.Remove(wineloaderCopy); err != nil {
			errMsg := fmt.Sprintf("failed to remove file %s: %v", wineloaderCopy, err)
			dialog.ShowError(errors.New(errMsg), myWindow)
			debug.Println(errMsg)
			updateAllStatuses()
			return
		} else {
			debug.Printf("Successfully removed file: %s", wineloaderCopy)
		}
	} else {
		debug.Printf("File not found to remove: %s", wineloaderCopy)
	}

	debug.Println("CrossOver unpatching completed successfully.")
	paths.PatchesAppliedCrossOver = false
	dialog.ShowInformation("Success", "CrossOver unpatching process completed.", myWindow)
	updateAllStatuses()
}

// applyShadowLODSetting applies the shadowLOD setting to Config.wtf for FPS optimization
func applyShadowLODSetting() error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	// Create WTF directory if it doesn't exist
	wtfDir := filepath.Dir(configPath)
	if err := os.MkdirAll(wtfDir, 0755); err != nil {
		return fmt.Errorf("failed to create WTF directory: %v", err)
	}

	var configText string

	// Read existing config if it exists
	if content, err := os.ReadFile(configPath); err == nil {
		configText = string(content)
	} else {
		debug.Printf("Config.wtf not found, creating new file")
		configText = ""
	}

	// Apply shadowLOD setting
	configText = updateOrAddConfigSetting(configText, "shadowLOD", "0")

	// Write the updated config back to file
	if err := os.WriteFile(configPath, []byte(configText), 0644); err != nil {
		return fmt.Errorf("failed to write Config.wtf: %v", err)
	}

	debug.Printf("Successfully applied shadowLOD setting to Config.wtf")
	return nil
}

// updateOrAddConfigSetting updates an existing setting or adds a new one if it doesn't exist
func updateOrAddConfigSetting(configText, setting, value string) string {
	// Create regex pattern to match the setting
	pattern := fmt.Sprintf(`SET\s+%s\s+"[^"]*"`, regexp.QuoteMeta(setting))
	re := regexp.MustCompile(pattern)

	newSetting := fmt.Sprintf(`SET %s "%s"`, setting, value)

	if re.MatchString(configText) {
		// Replace existing setting
		configText = re.ReplaceAllString(configText, newSetting)
		debug.Printf("Updated setting %s to %s", setting, value)
	} else {
		// Add new setting
		if configText != "" && !strings.HasSuffix(configText, "\n") {
			configText += "\n"
		}
		configText += newSetting + "\n"
		debug.Printf("Added new setting %s with value %s", setting, value)
	}

	return configText
}

// CheckShadowLODSetting checks if the shadowLOD setting is correctly applied in Config.wtf
func CheckShadowLODSetting() bool {
	if paths.TurtlewowPath == "" {
		return false
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	configText := string(content)
	return isConfigSettingCorrect(configText, "shadowLOD", "0")
}

// isConfigSettingCorrect checks if a specific setting has the correct value in the config text
func isConfigSettingCorrect(configText, setting, expectedValue string) bool {
	// Create regex pattern to match the setting
	pattern := fmt.Sprintf(`SET\s+%s\s+"([^"]*)"`, regexp.QuoteMeta(setting))
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(configText)
	if len(matches) < 2 {
		return false
	}

	currentValue := matches[1]
	return currentValue == expectedValue
}

// removeShadowLODSetting removes the shadowLOD setting from Config.wtf
func removeShadowLODSetting() error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		debug.Printf("Config.wtf not found, nothing to remove")
		return nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read Config.wtf: %v", err)
	}

	configText := string(content)

	// Remove shadowLOD setting if it exists
	pattern := fmt.Sprintf(`SET\s+%s\s+"[^"]*"[\r\n]*`, regexp.QuoteMeta("shadowLOD"))
	re := regexp.MustCompile(pattern)

	if re.MatchString(configText) {
		configText = re.ReplaceAllString(configText, "")
		debug.Printf("Removed shadowLOD setting from Config.wtf")

		// Write the updated config back to file
		if err := os.WriteFile(configPath, []byte(configText), 0644); err != nil {
			return fmt.Errorf("failed to write Config.wtf: %v", err)
		}
		debug.Printf("Successfully updated Config.wtf")
	} else {
		debug.Printf("shadowLOD setting not found in Config.wtf, nothing to remove")
	}

	return nil
}
