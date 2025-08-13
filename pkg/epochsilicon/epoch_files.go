package epochsilicon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/utils"
)

// GetRequiredFiles returns the list of required files for EpochSilicon from API
func GetRequiredFiles() ([]RequiredFile, error) {
	return GetRequiredFilesFromAPI()
}

// CheckEpochSiliconFiles validates that all required EpochSilicon files exist using API
func CheckEpochSiliconFiles(gamePath string) ([]RequiredFile, error) {
	if gamePath == "" {
		return nil, fmt.Errorf("game path not set")
	}

	// First check if WoW.exe exists
	wowExePath := filepath.Join(gamePath, "WoW.exe")
	if !utils.PathExists(wowExePath) {
		return nil, fmt.Errorf("WoW.exe not found in %s. Please select a valid WoW directory", gamePath)
	}

	// Get files from API
	requiredFiles, err := GetRequiredFilesFromAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to get required files from Project Epoch API: %v", err)
	}

	var missingFiles []RequiredFile

	for _, file := range requiredFiles {
		fullPath := filepath.Join(gamePath, file.RelativePath)
		if !utils.PathExists(fullPath) {
			missingFiles = append(missingFiles, file)
			debug.Printf("Missing EpochSilicon file: %s", file.RelativePath)
		}
	}

	debug.Printf("EpochSilicon file check complete. Missing: %d files", len(missingFiles))
	return missingFiles, nil
}

// CheckForUpdates checks if any files need updating using hash comparison from API
func CheckForUpdates(gamePath string) ([]RequiredFile, error) {
	if gamePath == "" {
		return nil, fmt.Errorf("game path not set")
	}

	// Get files from API
	requiredFiles, err := GetRequiredFilesFromAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to get required files from Project Epoch API: %v", err)
	}

	var updatesAvailable []RequiredFile

	// Use channels and goroutines for parallel checking
	resultChan := make(chan fileUpdateResult, len(requiredFiles))
	var wg sync.WaitGroup

	// Check each file in parallel
	for _, file := range requiredFiles {
		wg.Add(1)
		go func(f RequiredFile) {
			defer wg.Done()

			fullPath := filepath.Join(gamePath, f.RelativePath)
			result := fileUpdateResult{file: f}

			// If file doesn't exist locally, it needs to be downloaded
			if !utils.PathExists(fullPath) {
				result.needsUpdate = true
				debug.Printf("File missing, needs download: %s", f.RelativePath)
				resultChan <- result
				return
			}

			// Compare hash if available from API
			if f.Hash != "" {
				localHash, err := calculateFileHash(fullPath)
				if err != nil {
					result.err = fmt.Errorf("failed to calculate local hash: %v", err)
					debug.Printf("Failed to calculate hash for %s: %v", f.RelativePath, err)
					resultChan <- result
					return
				}

				if localHash != f.Hash {
					result.needsUpdate = true
					debug.Printf("Hash mismatch for %s: local=%s, remote=%s", f.RelativePath, localHash, f.Hash)
				} else {
					debug.Printf("Hash match for %s: %s", f.RelativePath, localHash)
				}
			} else {
				// Fallback to size comparison if no hash available
				if stat, err := os.Stat(fullPath); err == nil {
					if f.Size > 0 && stat.Size() != f.Size {
						result.needsUpdate = true
						debug.Printf("Size mismatch for %s: local=%d, remote=%d", f.RelativePath, stat.Size(), f.Size)
					}
				}
			}

			resultChan <- result
		}(file)
	}

	// Close the channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		if result.err != nil {
			// Log error but continue checking other files
			debug.Printf("Error checking %s: %v", result.file.RelativePath, result.err)
			continue
		}

		if result.needsUpdate {
			updatesAvailable = append(updatesAvailable, result.file)
		}
	}

	debug.Printf("Update check complete. Files needing updates: %d", len(updatesAvailable))
	return updatesAvailable, nil
}

// UpdateRealmlistForEpochSilicon always updates the realmlist.wtf file for EpochSilicon
func UpdateRealmlistForEpochSilicon(gamePath string) error {
	if gamePath == "" {
		return fmt.Errorf("game path not set")
	}

	// Get realmlist from API
	requiredFiles, err := GetRequiredFilesFromAPI()
	if err != nil {
		return fmt.Errorf("failed to get required files from Project Epoch API: %v", err)
	}

	// Find realmlist file in API response
	for _, file := range requiredFiles {
		if file.RelativePath == "Data/enUS/realmlist.wtf" ||
			strings.Contains(file.RelativePath, "realmlist.wtf") {
			debug.Printf("Updating realmlist.wtf from API for EpochSilicon")
			return downloadFileWithVerification(gamePath, file, nil)
		}
	}

	return fmt.Errorf("realmlist.wtf not found in Project Epoch API manifest")
}
