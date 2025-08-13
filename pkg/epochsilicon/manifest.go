package epochsilicon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"turtlesilicon/pkg/debug"
)

// fetchManifest retrieves the latest manifest from the Project Epoch API
func fetchManifest() (*EpochManifest, error) {
	debug.Printf("Fetching manifest from: %s", ManifestAPIURL)

	resp, err := http.Get(ManifestAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest API returned status %d", resp.StatusCode)
	}

	var manifest EpochManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %v", err)
	}

	debug.Printf("Fetched manifest: Version=%s, Files=%d", manifest.Version, len(manifest.Files))
	return &manifest, nil
}

// convertManifestToRequiredFiles converts manifest files to RequiredFile format
func convertManifestToRequiredFiles(manifest *EpochManifest) []RequiredFile {
	var files []RequiredFile

	cdnPriority := strings.Split(CDNPriority, ",")

	for _, file := range manifest.Files {
		// Determine best download URL based on CDN priority
		downloadURL := ""
		for _, cdn := range cdnPriority {
			if url, exists := file.URLs[cdn]; exists {
				downloadURL = url
				break
			}
		}

		// Fallback to any available URL if none match priority
		if downloadURL == "" {
			for _, url := range file.URLs {
				downloadURL = url
				break
			}
		}

		if downloadURL != "" {
			// Convert Windows path separators to forward slashes for cross-platform compatibility
			normalizedPath := strings.ReplaceAll(file.Path, "\\", "/")

			requiredFile := RequiredFile{
				RelativePath: normalizedPath,
				DownloadURL:  downloadURL,
				DisplayName:  normalizedPath,
				Hash:         file.Hash,
				Size:         file.Size,
				URLs:         file.URLs,
			}
			files = append(files, requiredFile)
		}
	}

	debug.Printf("Converted %d manifest files to RequiredFiles", len(files))
	return files
}

// GetRequiredFilesFromAPI returns the list of required files from the API
func GetRequiredFilesFromAPI() ([]RequiredFile, error) {
	manifest, err := fetchManifest()
	if err != nil {
		return nil, err
	}

	return convertManifestToRequiredFiles(manifest), nil
}
