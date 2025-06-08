package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	wineRegistrySection = "[Software\\Wine\\Mac Driver]"
	leftOptionKey       = "\"LeftOptionIsAlt\"=\"Y\""
	rightOptionKey      = "\"RightOptionIsAlt\"=\"Y\""
	wineLoaderPath      = "/Applications/CrossOver.app/Contents/SharedSupport/CrossOver/CrossOver-Hosted Application/wineloader2"
	registryKeyPath     = "HKEY_CURRENT_USER\\Software\\Wine\\Mac Driver"
)

// GetWineUserRegPath returns the path to the Wine user.reg file
func GetWineUserRegPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return filepath.Join(homeDir, ".wine", "user.reg"), nil
}

// queryRegistryValue queries a specific registry value using Wine's reg command
func queryRegistryValue(winePrefix, valueName string) bool {
	cmd := exec.Command(wineLoaderPath, "reg", "query", registryKeyPath, "/v", valueName)
	cmd.Env = append(os.Environ(), fmt.Sprintf("WINEPREFIX=%s", winePrefix))

	output, err := cmd.Output()
	if err != nil {
		// Value doesn't exist or other error
		return false
	}

	// Check if the output contains the value set to "Y"
	outputStr := string(output)
	return strings.Contains(outputStr, valueName) && strings.Contains(outputStr, "Y")
}

func setRegistryValuesOptimized(winePrefix string, enabled bool) error {
	if enabled {
		return addBothRegistryValues(winePrefix)
	} else {
		return deleteBothRegistryValues(winePrefix)
	}
}

func addBothRegistryValues(winePrefix string) error {
	batchContent := fmt.Sprintf(`@echo off
reg add "%s" /v "LeftOptionIsAlt" /t REG_SZ /d "Y" /f
reg add "%s" /v "RightOptionIsAlt" /t REG_SZ /d "Y" /f
`, registryKeyPath, registryKeyPath)

	// Create temporary batch file
	tempDir := os.TempDir()
	batchFile := filepath.Join(tempDir, "wine_registry_add.bat")

	if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err != nil {
		return fmt.Errorf("failed to create batch file: %v", err)
	}
	defer os.Remove(batchFile)

	// Run the batch file with Wine
	cmd := exec.Command(wineLoaderPath, "cmd", "/c", batchFile)
	cmd.Env = append(os.Environ(), fmt.Sprintf("WINEPREFIX=%s", winePrefix))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("batch registry add failed: %v, output: %s", err, string(output))
	}

	log.Printf("Successfully enabled Option-as-Alt mapping in Wine registry (optimized)")
	return nil
}

func deleteBothRegistryValues(winePrefix string) error {
	batchContent := fmt.Sprintf(`@echo off
reg delete "%s" /v "LeftOptionIsAlt" /f 2>nul
reg delete "%s" /v "RightOptionIsAlt" /f 2>nul
`, registryKeyPath, registryKeyPath)

	// Create temporary batch file
	tempDir := os.TempDir()
	batchFile := filepath.Join(tempDir, "wine_registry_delete.bat")

	if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err != nil {
		return fmt.Errorf("failed to create batch file: %v", err)
	}
	defer os.Remove(batchFile) // Clean up

	// Run the batch file with Wine
	cmd := exec.Command(wineLoaderPath, "cmd", "/c", batchFile)
	cmd.Env = append(os.Environ(), fmt.Sprintf("WINEPREFIX=%s", winePrefix))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("batch registry delete failed: %v, output: %s", err, string(output))
	}

	log.Printf("Successfully disabled Option-as-Alt mapping in Wine registry (optimized)")
	return nil
}

// CheckOptionAsAltEnabled checks if Option keys are remapped as Alt keys in Wine registry
func CheckOptionAsAltEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get user home directory: %v", err)
		return false
	}

	winePrefix := filepath.Join(homeDir, ".wine")

	// Check if CrossOver wine loader exists
	if !PathExists(wineLoaderPath) {
		log.Printf("CrossOver wine loader not found at: %s", wineLoaderPath)
		return false
	}

	// Query both registry values
	leftEnabled := queryRegistryValue(winePrefix, "LeftOptionIsAlt")
	rightEnabled := queryRegistryValue(winePrefix, "RightOptionIsAlt")

	return leftEnabled && rightEnabled
}

// CheckOptionAsAltEnabledFast checks status by reading user.reg file directly
func CheckOptionAsAltEnabledFast() bool {
	regPath, err := GetWineUserRegPath()
	if err != nil {
		log.Printf("Failed to get Wine registry path: %v", err)
		return false
	}

	if !PathExists(regPath) {
		log.Printf("Wine user.reg file not found at: %s", regPath)
		return false
	}

	content, err := os.ReadFile(regPath)
	if err != nil {
		log.Printf("Failed to read Wine registry file: %v", err)
		return false
	}

	contentStr := string(content)

	// Check for the Mac Driver section in different possible formats
	macDriverSectionFound := strings.Contains(contentStr, wineRegistrySection) ||
		strings.Contains(contentStr, "[SoftwareWineMac Driver]")

	if macDriverSectionFound {
		// Look for the registry values in the proper format or any format
		leftOptionFound := strings.Contains(contentStr, leftOptionKey) ||
			strings.Contains(contentStr, "\"LeftOptionIsAlt\"=\"Y\"")
		rightOptionFound := strings.Contains(contentStr, rightOptionKey) ||
			strings.Contains(contentStr, "\"RightOptionIsAlt\"=\"Y\"")

		return leftOptionFound && rightOptionFound
	}

	return false
}

// SetOptionAsAltEnabled enables or disables Option key remapping in Wine registry
func SetOptionAsAltEnabled(enabled bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	winePrefix := filepath.Join(homeDir, ".wine")

	// Check if CrossOver wine loader exists
	if !PathExists(wineLoaderPath) {
		return fmt.Errorf("CrossOver wine loader not found at: %s", wineLoaderPath)
	}

	// Ensure the .wine directory exists
	if err := os.MkdirAll(winePrefix, 0755); err != nil {
		return fmt.Errorf("failed to create .wine directory: %v", err)
	}

	if enabled {
		return setRegistryValuesOptimized(winePrefix, true)
	} else {
		err := setRegistryValuesOptimized(winePrefix, false)
		if err != nil {
			log.Printf("Wine registry disable failed: %v", err)
		}

		err2 := setRegistryValuesFast(false)
		if err2 != nil {
			log.Printf("File-based cleanup failed: %v", err2)
		}

		if err != nil {
			return err
		}
		return err2
	}
}

// setRegistryValuesFast directly modifies the user.reg file (much faster)
func setRegistryValuesFast(enabled bool) error {
	regPath, err := GetWineUserRegPath()
	if err != nil {
		return fmt.Errorf("failed to get Wine registry path: %v", err)
	}

	// Ensure the .wine directory exists
	wineDir := filepath.Dir(regPath)
	if err := os.MkdirAll(wineDir, 0755); err != nil {
		return fmt.Errorf("failed to create .wine directory: %v", err)
	}

	var content string
	var lines []string

	// Read existing content if file exists
	if PathExists(regPath) {
		contentBytes, err := os.ReadFile(regPath)
		if err != nil {
			return fmt.Errorf("failed to read existing registry file: %v", err)
		}
		content = string(contentBytes)
		lines = strings.Split(content, "\n")
	} else {
		// Create basic registry structure if file doesn't exist
		content = "WINE REGISTRY Version 2\n;; All keys relative to \\User\n\n"
		lines = strings.Split(content, "\n")
	}

	if enabled {
		return addOptionAsAltSettingsFast(regPath, lines)
	} else {
		return removeOptionAsAltSettingsFast(regPath, lines)
	}
}

// addOptionAsAltSettingsFast adds the Option-as-Alt registry settings directly to the file
func addOptionAsAltSettingsFast(regPath string, lines []string) error {
	var newLines []string
	sectionFound := false
	sectionIndex := -1
	leftOptionFound := false
	rightOptionFound := false

	// Find the Mac Driver section
	for i, line := range lines {
		if strings.TrimSpace(line) == wineRegistrySection {
			sectionFound = true
			sectionIndex = i
			break
		}
	}

	if !sectionFound {
		// Add the section at the end
		newLines = append(lines, "")
		newLines = append(newLines, wineRegistrySection)
		newLines = append(newLines, "#time=1dbd859c084de18")
		newLines = append(newLines, leftOptionKey)
		newLines = append(newLines, rightOptionKey)
	} else {
		// Section exists, check if keys are already present
		newLines = make([]string, len(lines))
		copy(newLines, lines)

		// Look for existing keys in the section
		for i := sectionIndex + 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "[") && line != wineRegistrySection {
				// Found start of another section, stop looking
				break
			}
			if strings.Contains(line, "LeftOptionIsAlt") {
				leftOptionFound = true
				if !strings.Contains(line, "\"Y\"") {
					newLines[i] = leftOptionKey
				}
			}
			if strings.Contains(line, "RightOptionIsAlt") {
				rightOptionFound = true
				if !strings.Contains(line, "\"Y\"") {
					newLines[i] = rightOptionKey
				}
			}
		}

		// Add missing keys
		if !leftOptionFound || !rightOptionFound {
			insertIndex := sectionIndex + 1

			// Add timestamp if it doesn't exist
			timestampExists := false
			for i := sectionIndex + 1; i < len(newLines); i++ {
				if strings.HasPrefix(strings.TrimSpace(newLines[i]), "#time=") {
					timestampExists = true
					break
				}
				if strings.HasPrefix(strings.TrimSpace(newLines[i]), "[") && newLines[i] != wineRegistrySection {
					break
				}
			}

			if !timestampExists {
				timestampLine := "#time=1dbd859c084de18"
				newLines = insertLine(newLines, insertIndex, timestampLine)
				insertIndex++
			}

			if !leftOptionFound {
				newLines = insertLine(newLines, insertIndex, leftOptionKey)
				insertIndex++
			}
			if !rightOptionFound {
				newLines = insertLine(newLines, insertIndex, rightOptionKey)
			}
		}
	}

	// Write the updated content
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(regPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %v", err)
	}

	log.Printf("Successfully enabled Option-as-Alt mapping in Wine registry (fast method)")
	return nil
}

// removeOptionAsAltSettingsFast removes the Option-as-Alt registry settings directly from the file
func removeOptionAsAltSettingsFast(regPath string, lines []string) error {
	if !PathExists(regPath) {
		// File doesn't exist, nothing to remove
		log.Printf("Successfully disabled Option-as-Alt mapping in Wine registry (no file to modify)")
		return nil
	}

	var newLines []string

	// Remove lines that contain our option key settings from any section
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip lines that contain our option key settings
		if strings.Contains(trimmedLine, "LeftOptionIsAlt") || strings.Contains(trimmedLine, "RightOptionIsAlt") {
			continue
		}

		newLines = append(newLines, line)
	}

	// Check if any Mac Driver sections are now empty and remove them
	newLines = removeEmptyMacDriverSections(newLines)

	// Write the updated content
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(regPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %v", err)
	}

	log.Printf("Successfully disabled Option-as-Alt mapping in Wine registry (fast method)")
	return nil
}

// removeEmptyMacDriverSections removes empty Mac Driver sections from the registry
func removeEmptyMacDriverSections(lines []string) []string {
	var finalLines []string
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Check if this is a Mac Driver section
		if line == wineRegistrySection || line == "[SoftwareWineMac Driver]" {
			// Check if the section is empty (only contains timestamp or nothing)
			sectionStart := i
			sectionEnd := i + 1
			sectionEmpty := true

			// Find the end of this section
			for sectionEnd < len(lines) {
				nextLine := strings.TrimSpace(lines[sectionEnd])
				if nextLine == "" {
					sectionEnd++
					continue
				}
				if strings.HasPrefix(nextLine, "[") {
					// Start of new section
					break
				}
				if !strings.HasPrefix(nextLine, "#time=") {
					// Found non-timestamp content in section
					sectionEmpty = false
				}
				sectionEnd++
			}

			if sectionEmpty {
				// Skip the entire empty section
				i = sectionEnd
				continue
			} else {
				// Keep the section
				for j := sectionStart; j < sectionEnd; j++ {
					finalLines = append(finalLines, lines[j])
				}
				i = sectionEnd
				continue
			}
		}

		finalLines = append(finalLines, lines[i])
		i++
	}

	return finalLines
}

// insertLine inserts a line at the specified index
func insertLine(lines []string, index int, newLine string) []string {
	if index >= len(lines) {
		return append(lines, newLine)
	}

	lines = append(lines, "")
	copy(lines[index+1:], lines[index:])
	lines[index] = newLine
	return lines
}
