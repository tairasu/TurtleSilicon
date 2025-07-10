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

	// Check user's preference for libSiliconPatch and shadowLOD
	prefs, _ := utils.LoadPrefs()

	// Enable by default unless user has explicitly disabled them
	shouldEnableLibSiliconPatch := true
	shouldEnableShadowLOD := true

	// If user has manually disabled these settings, respect their choice
	if prefs.UserDisabledLibSiliconPatch {
		shouldEnableLibSiliconPatch = false
		debug.Printf("libSiliconPatch disabled by user choice")
	} else {
		// Enable by default and update preferences
		prefs.EnableLibSiliconPatch = true
	}

	if prefs.UserDisabledShadowLOD {
		shouldEnableShadowLOD = false
		debug.Printf("shadowLOD disabled by user choice")
	} else {
		// Enable by default and update preferences
		prefs.SetShadowLOD0 = true
	}

	// Save updated preferences
	utils.SavePrefs(prefs)

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
		debug.Printf("dlls.txt not found, will create a new one")
	}

	// Only add libSiliconPatch if user wants it enabled
	if !shouldEnableLibSiliconPatch {
		needsLibSiliconPatchUpdate = false
		debug.Printf("libSiliconPatch disabled by user preference, will not add to dlls.txt")
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
		if needsLibSiliconPatchUpdate && shouldEnableLibSiliconPatch {
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

	// If user has disabled libSiliconPatch, make sure it's removed from dlls.txt
	if !shouldEnableLibSiliconPatch {
		if err := disableLibSiliconPatchInDlls(); err != nil {
			debug.Printf("Warning: failed to remove libSiliconPatch from dlls.txt: %v", err)
		}
	}

	// Always apply vertex animation shaders setting to Config.wtf
	if err := applyVertexAnimShadersSetting(); err != nil {
		debug.Printf("Warning: failed to apply vertex animation shaders setting to Config.wtf: %v", err)
		// Continue with patching even if Config.wtf update fails
	}

	// Apply shadowLOD setting to Config.wtf for FPS optimization
	// Use shouldEnableShadowLOD which accounts for first-time patching
	if shouldEnableShadowLOD {
		if err := applyShadowLODSetting(); err != nil {
			debug.Printf("Warning: failed to apply shadowLOD setting to Config.wtf: %v", err)
			// Continue with patching even if Config.wtf update fails
		}
	} else {
		// If user has disabled shadowLOD, make sure it's removed from Config.wtf
		if err := removeShadowLODSetting(); err != nil {
			debug.Printf("Warning: failed to remove shadowLOD setting from Config.wtf: %v", err)
		}
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

	// Remove shadowLOD setting from Config.wtf - only if it was applied via graphics settings
	prefs, _ := utils.LoadPrefs()
	if prefs.SetShadowLOD0 {
		if err := removeShadowLODSetting(); err != nil {
			debug.Printf("Warning: failed to remove shadowLOD setting from Config.wtf: %v", err)
			// Continue with unpatching even if Config.wtf update fails
		}
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

// removeConfigSetting removes a setting from the config text
func removeConfigSetting(configText, setting string) string {
	// Create regex pattern to match the setting
	pattern := fmt.Sprintf(`SET\s+%s\s+"[^"]*"[\r\n]*`, regexp.QuoteMeta(setting))
	re := regexp.MustCompile(pattern)

	if re.MatchString(configText) {
		configText = re.ReplaceAllString(configText, "")
		debug.Printf("Removed setting %s from config", setting)
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

// applyVertexAnimShadersSetting applies the vertex animation shaders setting to Config.wtf
func applyVertexAnimShadersSetting() error {
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

	// Apply vertex animation shaders setting
	configText = updateOrAddConfigSetting(configText, "M2UseShaders", "1")

	// Write the updated config back to file
	if err := os.WriteFile(configPath, []byte(configText), 0644); err != nil {
		return fmt.Errorf("failed to write Config.wtf: %v", err)
	}

	debug.Printf("Successfully applied vertex animation shaders setting to Config.wtf")
	return nil
}

// ApplyGraphicsSettings applies the selected graphics settings to Config.wtf
func ApplyGraphicsSettings(myWindow fyne.Window) error {
	prefs, err := utils.LoadPrefs()
	if err != nil {
		return fmt.Errorf("failed to load preferences: %v", err)
	}

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

	// Apply or remove graphics settings based on user preferences
	if prefs.ReduceTerrainDistance {
		configText = updateOrAddConfigSetting(configText, "farclip", "177")
	} else {
		configText = removeConfigSetting(configText, "farclip")
	}

	if prefs.SetMultisampleTo2x {
		configText = updateOrAddConfigSetting(configText, "gxMultisample", "2")
	} else {
		configText = removeConfigSetting(configText, "gxMultisample")
	}

	if prefs.SetShadowLOD0 {
		configText = updateOrAddConfigSetting(configText, "shadowLOD", "0")
	} else {
		configText = removeConfigSetting(configText, "shadowLOD")
	}

	// Handle libSiliconPatch.dll in dlls.txt (only if DLL exists)
	libSiliconPatchPath := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
	if utils.PathExists(libSiliconPatchPath) {
		if prefs.EnableLibSiliconPatch {
			if err := enableLibSiliconPatchInDlls(); err != nil {
				debug.Printf("Warning: failed to enable libSiliconPatch in dlls.txt: %v", err)
			}
		} else {
			if err := disableLibSiliconPatchInDlls(); err != nil {
				debug.Printf("Warning: failed to disable libSiliconPatch in dlls.txt: %v", err)
			}
		}
	}

	// Write the updated config back to file
	if err := os.WriteFile(configPath, []byte(configText), 0644); err != nil {
		return fmt.Errorf("failed to write Config.wtf: %v", err)
	}

	debug.Printf("Successfully applied graphics settings to Config.wtf")
	return nil
}

// CheckGraphicsSettings checks if the graphics settings are correctly applied in Config.wtf
func CheckGraphicsSettings() (bool, bool, bool) {
	prefs, _ := utils.LoadPrefs()

	if paths.TurtlewowPath == "" {
		return false, false, false
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, false, false
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return false, false, false
	}

	configText := string(content)

	terrainCorrect := !prefs.ReduceTerrainDistance || isConfigSettingCorrect(configText, "farclip", "177")
	multisampleCorrect := !prefs.SetMultisampleTo2x || isConfigSettingCorrect(configText, "gxMultisample", "2")
	shadowCorrect := !prefs.SetShadowLOD0 || isConfigSettingCorrect(configText, "shadowLOD", "0")

	return terrainCorrect, multisampleCorrect, shadowCorrect
}

// LoadGraphicsSettingsFromConfig reads Config.wtf and updates preferences with current settings
func LoadGraphicsSettingsFromConfig() error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	// If Config.wtf doesn't exist, nothing to load
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		debug.Printf("Config.wtf not found, using default graphics settings")
		return nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read Config.wtf: %v", err)
	}

	configText := string(content)

	// Load current preferences
	prefs, err := utils.LoadPrefs()
	if err != nil {
		return fmt.Errorf("failed to load preferences: %v", err)
	}

	// Check each graphics setting and update preferences
	prefs.ReduceTerrainDistance = isConfigSettingCorrect(configText, "farclip", "177")
	prefs.SetMultisampleTo2x = isConfigSettingCorrect(configText, "gxMultisample", "2")
	prefs.SetShadowLOD0 = isConfigSettingCorrect(configText, "shadowLOD", "0")

	// Check libSiliconPatch status (DLL exists and enabled in dlls.txt)
	libSiliconPatchPath := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")
	libSiliconPatchExists := utils.PathExists(libSiliconPatchPath)
	libSiliconPatchEnabled := false

	if libSiliconPatchExists && utils.PathExists(dllsTextFile) {
		if dllsContent, err := os.ReadFile(dllsTextFile); err == nil {
			libSiliconPatchEnabled = strings.Contains(string(dllsContent), "libSiliconPatch.dll")
		}
	}
	prefs.EnableLibSiliconPatch = libSiliconPatchExists && libSiliconPatchEnabled

	// Save updated preferences
	if err := utils.SavePrefs(prefs); err != nil {
		return fmt.Errorf("failed to save preferences: %v", err)
	}

	debug.Printf("Loaded graphics settings from Config.wtf: terrain=%v, multisample=%v, shadow=%v, libSiliconPatch=%v",
		prefs.ReduceTerrainDistance, prefs.SetMultisampleTo2x, prefs.SetShadowLOD0, prefs.EnableLibSiliconPatch)

	return nil
}

// CheckGraphicsSettingsPresence checks if libSiliconPatch.dll exists and shadowLOD is applied, updates preferences accordingly
func CheckGraphicsSettingsPresence() {
	if paths.TurtlewowPath == "" {
		return
	}

	libSiliconPatchPath := filepath.Join(paths.TurtlewowPath, "libSiliconPatch.dll")
	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")

	// Check if libSiliconPatch.dll exists
	libSiliconPatchExists := utils.PathExists(libSiliconPatchPath)

	// Check if it's enabled in dlls.txt
	libSiliconPatchEnabled := false
	if utils.PathExists(dllsTextFile) {
		if fileContentBytes, err := os.ReadFile(dllsTextFile); err == nil {
			fileContent := string(fileContentBytes)
			libSiliconPatchEnabled = strings.Contains(fileContent, "libSiliconPatch.dll")
		}
	}

	// Check if shadowLOD is currently applied
	shadowLODApplied := CheckShadowLODSetting()

	// Load current preferences
	prefs, _ := utils.LoadPrefs()

	// Handle libSiliconPatch preference detection
	if libSiliconPatchExists {
		if libSiliconPatchEnabled && !prefs.EnableLibSiliconPatch {
			// DLL is currently enabled but user preference says disabled - likely first run detection
			prefs.EnableLibSiliconPatch = true
			debug.Printf("libSiliconPatch detected as enabled, setting user preference to enabled")
		} else if !libSiliconPatchEnabled && prefs.EnableLibSiliconPatch {
			// DLL exists but not enabled, user preference says enabled - respect user choice
			debug.Printf("libSiliconPatch disabled in dlls.txt but user preference is enabled - keeping user preference")
		}
	}

	// Handle shadowLOD preference detection - enable by default if currently applied
	if shadowLODApplied && !prefs.SetShadowLOD0 {
		// shadowLOD is currently applied but user preference says disabled - likely first run detection
		prefs.SetShadowLOD0 = true
		debug.Printf("shadowLOD detected as applied, setting user preference to enabled")
	} else if !shadowLODApplied && prefs.SetShadowLOD0 {
		// shadowLOD not applied but user preference says enabled - respect user choice
		debug.Printf("shadowLOD not applied but user preference is enabled - keeping user preference")
	}

	// Save any changes
	utils.SavePrefs(prefs)

	debug.Printf("Graphics settings detection: libSiliconPatch exists=%v, enabled_in_dlls=%v, user_setting=%v; shadowLOD applied=%v, user_setting=%v",
		libSiliconPatchExists, libSiliconPatchEnabled, prefs.EnableLibSiliconPatch, shadowLODApplied, prefs.SetShadowLOD0)
}

// enableLibSiliconPatchInDlls adds libSiliconPatch.dll to dlls.txt if not present
func enableLibSiliconPatchInDlls() error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}

	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")
	libSiliconPatchEntry := "libSiliconPatch.dll"

	var fileContentBytes []byte
	var err error
	if utils.PathExists(dllsTextFile) {
		fileContentBytes, err = os.ReadFile(dllsTextFile)
		if err != nil {
			return fmt.Errorf("failed to read dlls.txt: %v", err)
		}
	}

	currentContent := string(fileContentBytes)
	if strings.Contains(currentContent, libSiliconPatchEntry) {
		debug.Printf("libSiliconPatch.dll already present in dlls.txt")
		return nil
	}

	// Add libSiliconPatch.dll to dlls.txt
	if len(currentContent) > 0 && !strings.HasSuffix(currentContent, "\n") {
		currentContent += "\n"
	}
	currentContent += libSiliconPatchEntry + "\n"

	if err := os.WriteFile(dllsTextFile, []byte(currentContent), 0644); err != nil {
		return fmt.Errorf("failed to update dlls.txt: %v", err)
	}

	debug.Printf("Added libSiliconPatch.dll to dlls.txt")
	return nil
}

// disableLibSiliconPatchInDlls removes libSiliconPatch.dll from dlls.txt
func disableLibSiliconPatchInDlls() error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("TurtleWoW path not set")
	}

	dllsTextFile := filepath.Join(paths.TurtlewowPath, "dlls.txt")

	if !utils.PathExists(dllsTextFile) {
		debug.Printf("dlls.txt not found, nothing to remove")
		return nil
	}

	content, err := os.ReadFile(dllsTextFile)
	if err != nil {
		return fmt.Errorf("failed to read dlls.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "libSiliconPatch.dll" {
			filteredLines = append(filteredLines, line)
		}
	}

	updatedContent := strings.Join(filteredLines, "\n")
	if err := os.WriteFile(dllsTextFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to update dlls.txt: %v", err)
	}

	debug.Printf("Removed libSiliconPatch.dll from dlls.txt")
	return nil
}
