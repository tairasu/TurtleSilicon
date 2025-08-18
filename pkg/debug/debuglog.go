package debug

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DebugInfo contains all the information needed to generate a debug log
type DebugInfo struct {
	CrossoverPath            string
	TurtlewowPath            string
	PatchesAppliedTurtleWoW  bool
	PatchesAppliedCrossOver  bool
	RosettaX87ServiceRunning bool
	ServiceStarting          bool
}

// GameVersionInfo contains version-specific information
type GameVersionInfo struct {
	ID                    string
	DisplayName           string
	WoWVersion            string
	GamePath              string
	ExecutableName        string
	SupportsDLLLoading    bool
	UsesRosettaPatching   bool
	UsesDivxDecoderPatch  bool
	Settings              GameVersionSettings
}

type GameVersionSettings struct {
	RemapOptionAsAlt      bool
	AutoDeleteWdb         bool
	EnableMetalHud        bool
	SaveSudoPassword      bool
	ShowTerminalNormally  bool
	EnvironmentVariables  string
	ReduceTerrainDistance bool
	SetMultisampleTo2x    bool
	SetShadowLOD0         bool
	EnableLibSiliconPatch bool
}

// GenerateDebugLog creates a comprehensive debug log for troubleshooting
func GenerateDebugLog(debugInfo *DebugInfo, currentVersion *GameVersionInfo) string {
	var log strings.Builder

	log.WriteString("=== TurtleSilicon Debug Log ===\n")
	log.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// === System Information ===
	log.WriteString("=== System Information ===\n")
	log.WriteString(fmt.Sprintf("OS: %s\n", runtime.GOOS))
	log.WriteString(fmt.Sprintf("Architecture: %s\n", runtime.GOARCH))

	// Get macOS version
	if cmd := exec.Command("sw_vers"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			log.WriteString(fmt.Sprintf("macOS Version:\n%s\n", string(output)))
		} else {
			log.WriteString(fmt.Sprintf("macOS Version: Unable to detect (%v)\n", err))
		}
	}

	// === TurtleSilicon Version Information ===
	log.WriteString("\n=== TurtleSilicon Configuration ===\n")

	if currentVersion != nil {
		log.WriteString(fmt.Sprintf("Current Game Version: %s (%s)\n", currentVersion.DisplayName, currentVersion.ID))
		log.WriteString(fmt.Sprintf("WoW Version: %s\n", currentVersion.WoWVersion))
		log.WriteString(fmt.Sprintf("Game Path: %s\n", currentVersion.GamePath))
		log.WriteString(fmt.Sprintf("Executable: %s\n", currentVersion.ExecutableName))
		log.WriteString(fmt.Sprintf("Supports DLL Loading: %v\n", currentVersion.SupportsDLLLoading))
		log.WriteString(fmt.Sprintf("Uses Rosetta Patching: %v\n", currentVersion.UsesRosettaPatching))
		log.WriteString(fmt.Sprintf("Uses LibDllLdr Patch: %v\n", currentVersion.UsesDivxDecoderPatch))

		// Version settings
		settings := currentVersion.Settings
		log.WriteString("\nVersion Settings:\n")
		log.WriteString(fmt.Sprintf("  Remap Option as Alt: %v\n", settings.RemapOptionAsAlt))
		log.WriteString(fmt.Sprintf("  Auto Delete WDB: %v\n", settings.AutoDeleteWdb))
		log.WriteString(fmt.Sprintf("  Metal HUD: %v\n", settings.EnableMetalHud))
		log.WriteString(fmt.Sprintf("  Save Sudo Password: %v\n", settings.SaveSudoPassword))
		log.WriteString(fmt.Sprintf("  Show Terminal Normally: %v\n", settings.ShowTerminalNormally))
		log.WriteString(fmt.Sprintf("  Environment Variables: %s\n", settings.EnvironmentVariables))
		log.WriteString(fmt.Sprintf("  Reduce Terrain Distance: %v\n", settings.ReduceTerrainDistance))
		log.WriteString(fmt.Sprintf("  Set Multisample to 2x: %v\n", settings.SetMultisampleTo2x))
		log.WriteString(fmt.Sprintf("  Set Shadow LOD 0: %v\n", settings.SetShadowLOD0))
		log.WriteString(fmt.Sprintf("  Enable LibSilicon Patch: %v\n", settings.EnableLibSiliconPatch))
	} else {
		log.WriteString("Current Version: No version information available\n")
	}

	// === Paths ===
	log.WriteString("\n=== Paths ===\n")
	log.WriteString(fmt.Sprintf("CrossOver Path: %s\n", debugInfo.CrossoverPath))
	log.WriteString(fmt.Sprintf("TurtleWoW Path: %s\n", debugInfo.TurtlewowPath))

	// Check if paths exist
	if debugInfo.CrossoverPath != "" {
		if _, err := os.Stat(debugInfo.CrossoverPath); err == nil {
			log.WriteString("CrossOver Path: ✓ Exists\n")
		} else {
			log.WriteString(fmt.Sprintf("CrossOver Path: ✗ Missing (%v)\n", err))
		}
	}

	if debugInfo.TurtlewowPath != "" {
		if _, err := os.Stat(debugInfo.TurtlewowPath); err == nil {
			log.WriteString("TurtleWoW Path: ✓ Exists\n")
		} else {
			log.WriteString(fmt.Sprintf("TurtleWoW Path: ✗ Missing (%v)\n", err))
		}
	}

	// === CrossOver Information ===
	log.WriteString("\n=== CrossOver Information ===\n")
	crossoverVersion := getCrossoverVersion(debugInfo.CrossoverPath)
	if crossoverVersion != "" {
		log.WriteString(fmt.Sprintf("CrossOver Version: %s\n", crossoverVersion))
		if isCrossoverVersionRecommended(crossoverVersion) {
			log.WriteString("CrossOver Version Status: ✓ Recommended\n")
		} else {
			log.WriteString("CrossOver Version Status: ⚠️ Update recommended\n")
		}
	} else {
		log.WriteString("CrossOver Version: Not found or invalid path\n")
	}

	// === Game Files Check ===
	log.WriteString("\n=== Game Files ===\n")

	// Check for current version's game path and executable
	if currentVersion != nil && currentVersion.GamePath != "" {
		gamePath := currentVersion.GamePath
		exePath := filepath.Join(gamePath, currentVersion.ExecutableName)

		log.WriteString(fmt.Sprintf("Game Directory: %s\n", gamePath))
		if _, err := os.Stat(exePath); err == nil {
			log.WriteString(fmt.Sprintf("Game Executable: ✓ Found (%s)\n", currentVersion.ExecutableName))
		} else {
			log.WriteString(fmt.Sprintf("Game Executable: ✗ Missing (%s)\n", currentVersion.ExecutableName))
		}

		// Check for dlls.txt
		dllsPath := filepath.Join(gamePath, "dlls.txt")
		if content, err := os.ReadFile(dllsPath); err == nil {
			log.WriteString(fmt.Sprintf("dlls.txt: ✓ Found\nContent:\n%s\n", string(content)))
		} else {
			log.WriteString("dlls.txt: Not found\n")
		}

		// Check for vanilla tweaks file
		wowTweakedPath := filepath.Join(gamePath, "WoW_tweaked.exe")
		if _, err := os.Stat(wowTweakedPath); err == nil {
			log.WriteString("WoW_tweaked.exe: ✓ Found\n")
		} else {
			log.WriteString("WoW_tweaked.exe: Not found\n")
		}

		// Check for patched executables (libDllLdr.dll approach)
		wowPatchedPath := filepath.Join(gamePath, "Wow_patched.exe")
		if _, err := os.Stat(wowPatchedPath); err == nil {
			log.WriteString("Wow_patched.exe: ✓ Found\n")
		} else {
			log.WriteString("Wow_patched.exe: Not found\n")
		}

		projectEpochPatchedPath := filepath.Join(gamePath, "Project-Epoch_patched.exe")
		if _, err := os.Stat(projectEpochPatchedPath); err == nil {
			log.WriteString("Project-Epoch_patched.exe: ✓ Found\n")
		} else {
			log.WriteString("Project-Epoch_patched.exe: Not found\n")
		}

		// Check for libDllLdr.dll
		libDllLdrPath := filepath.Join(gamePath, "libDllLdr.dll")
		if _, err := os.Stat(libDllLdrPath); err == nil {
			log.WriteString("libDllLdr.dll: ✓ Found\n")
		} else {
			log.WriteString("libDllLdr.dll: Not found\n")
		}

		// Check for config.wtf
		wdbPath := filepath.Join(gamePath, "WDB")
		configPath := filepath.Join(wdbPath, "enUS", "config.wtf")
		if content, err := os.ReadFile(configPath); err == nil {
			log.WriteString(fmt.Sprintf("config.wtf: ✓ Found\nContent:\n%s\n", string(content)))
		} else {
			// Try alternative path
			configPath = filepath.Join(gamePath, "WTF", "config.wtf")
			if content, err := os.ReadFile(configPath); err == nil {
				log.WriteString(fmt.Sprintf("config.wtf: ✓ Found (in WTF)\nContent:\n%s\n", string(content)))
			} else {
				log.WriteString("config.wtf: Not found in WDB/enUS or WTF\n")
			}
		}
	} else {
		log.WriteString("Game Directory: No game path configured\n")
	}

	// === Wine Information ===
	log.WriteString("\n=== Wine Information ===\n")
	homeDir, _ := os.UserHomeDir()
	userWinePrefix := filepath.Join(homeDir, ".wine")
	turtleWinePrefix := filepath.Join(debugInfo.TurtlewowPath, ".wine")

	// Check wine prefixes
	if _, err := os.Stat(userWinePrefix); err == nil {
		log.WriteString("User Wine Prefix (~/.wine): ✓ Exists\n")
	} else {
		log.WriteString("User Wine Prefix (~/.wine): Not found\n")
	}

	if debugInfo.TurtlewowPath != "" {
		if _, err := os.Stat(turtleWinePrefix); err == nil {
			log.WriteString("TurtleWoW Wine Prefix: ✓ Exists\n")
		} else {
			log.WriteString("TurtleWoW Wine Prefix: Not found\n")
		}
	}

	// === Patch Status ===
	log.WriteString("\n=== Patch Status ===\n")
	log.WriteString(fmt.Sprintf("Patches Applied (TurtleWoW): %v\n", debugInfo.PatchesAppliedTurtleWoW))
	log.WriteString(fmt.Sprintf("Patches Applied (CrossOver): %v\n", debugInfo.PatchesAppliedCrossOver))
	log.WriteString(fmt.Sprintf("Rosetta x87 Service Running: %v\n", debugInfo.RosettaX87ServiceRunning))
	log.WriteString(fmt.Sprintf("Service Starting: %v\n", debugInfo.ServiceStarting))

	// === Launch Command Information ===
	log.WriteString("\n=== Launch Command Information ===\n")
	if currentVersion != nil && currentVersion.GamePath != "" && debugInfo.CrossoverPath != "" {
		wineLoader := filepath.Join(debugInfo.CrossoverPath, "Contents", "SharedSupport", "CrossOver", "bin", "wine")
		exePath := filepath.Join(currentVersion.GamePath, currentVersion.ExecutableName)

		log.WriteString("Expected Launch Components:\n")
		log.WriteString(fmt.Sprintf("  Wine Loader: %s\n", wineLoader))
		log.WriteString(fmt.Sprintf("  Game Executable: %s\n", exePath))
		log.WriteString(fmt.Sprintf("  Environment Variables: %s\n", currentVersion.Settings.EnvironmentVariables))

		// Check if wine loader exists
		if _, err := os.Stat(wineLoader); err == nil {
			log.WriteString("  Wine Loader: ✓ Found\n")
		} else {
			log.WriteString("  Wine Loader: ✗ Missing\n")
		}
	} else {
		log.WriteString("Launch Command: Cannot generate - missing required paths or version info\n")
	}

	log.WriteString("\n=== End Debug Log ===\n")
	return log.String()
}

// Helper functions that need to be moved here or made accessible

func getCrossoverVersion(path string) string {
	if path == "" {
		return ""
	}

	plistPath := filepath.Join(path, "Contents", "Info.plist")

	// Using a simple plist decoder approach
	content, err := os.ReadFile(plistPath)
	if err != nil {
		return ""
	}

	// Simple string search for version (fallback if plist parsing fails)
	contentStr := string(content)
	if strings.Contains(contentStr, "CFBundleShortVersionString") {
		lines := strings.Split(contentStr, "\n")
		for i, line := range lines {
			if strings.Contains(line, "CFBundleShortVersionString") && i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				nextLine = strings.TrimPrefix(nextLine, "<string>")
				nextLine = strings.TrimSuffix(nextLine, "</string>")
				return nextLine
			}
		}
	}

	return ""
}

func isCrossoverVersionRecommended(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	// Convert version parts to integers for proper comparison
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	// Handle patch version (default to 0 if not present)
	patch := 0
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return false
		}
	}

	// Check if version >= 25.0.1
	if major > 25 {
		return true
	}
	if major == 25 {
		if minor > 0 {
			return true
		}
		if minor == 0 && patch >= 1 {
			return true
		}
	}
	return false
}
