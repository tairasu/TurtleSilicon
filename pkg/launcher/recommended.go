package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"
)

// RecommendedSettings contains the recommended graphics settings for optimal performance
var RecommendedSettings = map[string]string{
	"farclip":              "177",
	"M2UseShaders":         "1",
	"gxColorBits":          "24",
	"gxDepthBits":          "24",
	"gxMultisampleQuality": "0.000000",
	"gxMultisample":        "2",
	"shadowLOD":            "0",
}

// CheckRecommendedSettings reads the Config.wtf file and checks if all recommended settings are applied
// Returns true if all settings are correctly applied, false otherwise
func CheckRecommendedSettings() bool {
	if paths.TurtlewowPath == "" {
		debug.Printf("TurtleWoW path not set, cannot check Config.wtf")
		return false
	}

	configPath := filepath.Join(paths.TurtlewowPath, "WTF", "Config.wtf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		debug.Printf("Config.wtf not found at %s", configPath)
		return false
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		debug.Printf("Failed to read Config.wtf: %v", err)
		return false
	}

	configText := string(content)

	// Check each recommended setting
	for setting, expectedValue := range RecommendedSettings {
		if !isSettingCorrect(configText, setting, expectedValue) {
			debug.Printf("Setting %s not found or incorrect in Config.wtf", setting)
			return false
		}
	}

	debug.Printf("All recommended settings are correctly applied")
	return true
}

// isSettingCorrect checks if a specific setting has the correct value in the config text
func isSettingCorrect(configText, setting, expectedValue string) bool {
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

// ApplyRecommendedSettings applies all recommended graphics settings to Config.wtf
func ApplyRecommendedSettings() error {
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

	// Apply each recommended setting
	for setting, value := range RecommendedSettings {
		configText = updateOrAddSetting(configText, setting, value)
	}

	// Write the updated config back to file
	if err := os.WriteFile(configPath, []byte(configText), 0644); err != nil {
		return fmt.Errorf("failed to write Config.wtf: %v", err)
	}

	debug.Printf("Successfully applied recommended settings to Config.wtf")
	return nil
}

// updateOrAddSetting updates an existing setting or adds a new one if it doesn't exist
func updateOrAddSetting(configText, setting, value string) string {
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
