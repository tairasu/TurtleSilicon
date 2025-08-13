package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"turtlesilicon/pkg/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PathExists checks if a path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// CopyDir copies a directory recursively from src to dst.
func CopyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	dir, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range dir {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RunOsascript runs an AppleScript command using osascript.
func RunOsascript(scriptString string, myWindow fyne.Window) bool {
	debug.Printf("Executing AppleScript: %s", scriptString)
	cmd := exec.Command("osascript", "-e", scriptString)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("AppleScript failed: %v\nOutput: %s", err, string(output))
		dialog.ShowError(errors.New(errMsg), myWindow)
		debug.Println(errMsg)
		return false
	}
	debug.Printf("osascript output: %s", string(output))
	return true
}

// EscapeStringForAppleScript escapes a string for AppleScript.
func EscapeStringForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// QuotePathForShell quotes paths for shell commands.
func QuotePathForShell(path string) string {
	return fmt.Sprintf(`"%s"`, path)
}

func CheckForUpdate(currentVersion string) (latestVersion, releaseNotes string, updateAvailable bool, err error) {
	resp, err := http.Get("https://api.github.com/repos/tairasu/TurtleSilicon/releases/latest")
	if err != nil {
		return "", "", false, err
	}
	defer resp.Body.Close()

	var data struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", false, err
	}

	latest := strings.TrimPrefix(data.TagName, "v")
	return latest, data.Body, latest != currentVersion, nil
}

// UpdateInfo contains information about the latest release
type UpdateInfo struct {
	TagName string  `json:"tag_name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckForUpdateWithAssets returns update information including download assets
func CheckForUpdateWithAssets(currentVersion string) (*UpdateInfo, bool, error) {
	resp, err := http.Get("https://api.github.com/repos/tairasu/TurtleSilicon/releases/latest")
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read the response body first to check content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response looks like HTML (rate limiting or other errors)
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<!DOCTYPE html>") || strings.Contains(bodyStr, "<html") {
		return nil, false, fmt.Errorf("GitHub API returned HTML instead of JSON (possible rate limiting): %s", bodyStr[:min(200, len(bodyStr))])
	}

	var updateInfo UpdateInfo
	if err := json.Unmarshal(body, &updateInfo); err != nil {
		return nil, false, fmt.Errorf("failed to parse JSON response: %v. Response: %s", err, bodyStr[:min(200, len(bodyStr))])
	}

	latest := strings.TrimPrefix(updateInfo.TagName, "v")
	updateAvailable := latest != currentVersion

	return &updateInfo, updateAvailable, nil
}

// DownloadUpdate downloads the latest release and returns the path to the downloaded file
func DownloadUpdate(downloadURL string, progressCallback func(downloaded, total int64)) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "TurtleSilicon-update-*.dmg")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Download the file
	resp, err := http.Get(downloadURL)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	totalSize := resp.ContentLength
	var downloaded int64

	// Copy with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := tempFile.Write(buffer[:n])
			if writeErr != nil {
				os.Remove(tempFile.Name())
				return "", fmt.Errorf("failed to write update file: %v", writeErr)
			}
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempFile.Name())
			return "", fmt.Errorf("failed to read update data: %v", err)
		}
	}

	return tempFile.Name(), nil
}

// InstallUpdate installs the downloaded update by mounting the DMG and replacing the current app
func InstallUpdate(dmgPath string) error {
	// Get current app path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Navigate up to find the .app bundle
	currentAppPath := execPath
	for !strings.HasSuffix(currentAppPath, ".app") && currentAppPath != "/" {
		currentAppPath = filepath.Dir(currentAppPath)
	}

	if !strings.HasSuffix(currentAppPath, ".app") {
		return fmt.Errorf("could not find app bundle path")
	}

	// Mount the DMG and parse the mount point from plist output
	debug.Printf("Mounting DMG: %s", dmgPath)
	mountCmd := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse", "-plist")
	mountOutput, err := mountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount DMG: %v, output: %s", err, string(mountOutput))
	}

	debug.Printf("Mount output: %s", string(mountOutput))

	mountPoint := ""

	// Parse the plist XML output to find mount points
	outputStr := string(mountOutput)

	// Look for mount-point entries in the XML
	// The plist contains mount-point entries that show where volumes are mounted
	lines := strings.Split(outputStr, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "<key>mount-point</key>") && i+1 < len(lines) {
			// The next line should contain the mount point path
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.HasPrefix(nextLine, "<string>/Volumes/") {
				// Extract the path from <string>/Volumes/...</string>
				start := strings.Index(nextLine, "<string>") + 8
				end := strings.Index(nextLine, "</string>")
				if start >= 8 && end > start {
					mountPoint = nextLine[start:end]
					debug.Printf("Found mount point in plist: %s", mountPoint)
					break
				}
			}
		}
	}

	// Fallback: try without -plist flag for simpler output
	if mountPoint == "" {
		debug.Printf("Plist parsing failed, trying simple mount")
		// Unmount first if something was mounted
		exec.Command("hdiutil", "detach", dmgPath, "-force").Run()

		// Try mounting without plist
		simpleMountCmd := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse")
		simpleOutput, simpleErr := simpleMountCmd.CombinedOutput()
		if simpleErr != nil {
			return fmt.Errorf("failed to mount DMG (simple): %v, output: %s", simpleErr, string(simpleOutput))
		}

		// Parse simple output
		simpleLines := strings.Split(string(simpleOutput), "\n")
		for _, line := range simpleLines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "/Volumes/") {
				parts := strings.Fields(line)
				for i := len(parts) - 1; i >= 0; i-- {
					if strings.HasPrefix(parts[i], "/Volumes/") {
						mountPoint = parts[i]
						debug.Printf("Found mount point in simple output: %s", mountPoint)
						break
					}
				}
				if mountPoint != "" {
					break
				}
			}
		}
	}

	if mountPoint == "" {
		return fmt.Errorf("could not find mount point. Mount output: %s", string(mountOutput))
	}

	debug.Printf("Using mount point: %s", mountPoint)

	defer func() {
		// Unmount the DMG
		debug.Printf("Unmounting DMG from: %s", mountPoint)
		unmountCmd := exec.Command("hdiutil", "detach", mountPoint, "-force")
		unmountCmd.Run()
	}()

	// Find the app in the mounted DMG - search for any .app bundle
	var newAppPath string

	// First, try the exact name
	exactPath := filepath.Join(mountPoint, "TurtleSilicon.app")
	if PathExists(exactPath) {
		newAppPath = exactPath
	} else {
		// Search for any .app bundle in the mount point
		debug.Printf("TurtleSilicon.app not found at exact path, searching for .app bundles")
		entries, err := os.ReadDir(mountPoint)
		if err != nil {
			return fmt.Errorf("failed to read DMG contents: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
				candidatePath := filepath.Join(mountPoint, entry.Name())
				debug.Printf("Found .app bundle: %s", candidatePath)
				newAppPath = candidatePath
				break
			}
		}
	}

	if newAppPath == "" {
		return fmt.Errorf("no .app bundle found in DMG at %s", mountPoint)
	}

	debug.Printf("Found app to install: %s", newAppPath)

	// Create backup of current app
	backupPath := currentAppPath + ".backup"
	debug.Printf("Creating backup: %s -> %s", currentAppPath, backupPath)

	// Remove old backup if it exists
	if PathExists(backupPath) {
		os.RemoveAll(backupPath)
	}

	if err := CopyDir(currentAppPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Remove current app
	debug.Printf("Removing current app: %s", currentAppPath)
	if err := os.RemoveAll(currentAppPath); err != nil {
		return fmt.Errorf("failed to remove current app: %v", err)
	}

	// Copy new app
	debug.Printf("Installing new app: %s -> %s", newAppPath, currentAppPath)
	if err := CopyDir(newAppPath, currentAppPath); err != nil {
		// Try to restore backup on failure
		debug.Printf("Installation failed, restoring backup")
		os.RemoveAll(currentAppPath)
		CopyDir(backupPath, currentAppPath)
		return fmt.Errorf("failed to install new app: %v", err)
	}

	// Fix executable permissions for the main binary
	executablePath := filepath.Join(currentAppPath, "Contents", "MacOS", "turtlesilicon")
	if PathExists(executablePath) {
		debug.Printf("Setting executable permissions for: %s", executablePath)
		if err := os.Chmod(executablePath, 0755); err != nil {
			debug.Printf("Warning: failed to set executable permissions: %v", err)
			// Don't fail the entire update for this, but log it
		}
	} else {
		debug.Printf("Warning: executable not found at expected path: %s", executablePath)
	}

	// Remove backup on success
	os.RemoveAll(backupPath)

	debug.Printf("Update installed successfully")
	return nil
}

// TestDMGMount tests DMG mounting and returns mount point and app path for debugging
func TestDMGMount(dmgPath string) (string, string, error) {
	debug.Printf("Testing DMG mount: %s", dmgPath)

	// Mount the DMG with verbose output to better parse mount point
	mountCmd := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse", "-plist")
	mountOutput, err := mountCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to mount DMG: %v, output: %s", err, string(mountOutput))
	}

	debug.Printf("Mount output: %s", string(mountOutput))

	// Parse mount output to get mount point
	mountPoint := ""

	// First try: look for /Volumes/ in the output lines
	lines := strings.Split(string(mountOutput), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "/Volumes/") {
			parts := strings.Fields(line)
			for i := len(parts) - 1; i >= 0; i-- {
				if strings.HasPrefix(parts[i], "/Volumes/") {
					mountPoint = parts[i]
					break
				}
			}
			if mountPoint != "" {
				break
			}
		}
	}

	// Second try: use hdiutil info to get mount points if first method failed
	if mountPoint == "" {
		debug.Printf("First mount point detection failed, trying hdiutil info")
		infoCmd := exec.Command("hdiutil", "info", "-plist")
		infoOutput, infoErr := infoCmd.CombinedOutput()
		if infoErr == nil {
			infoLines := strings.Split(string(infoOutput), "\n")
			for _, line := range infoLines {
				if strings.Contains(line, "/Volumes/") && strings.Contains(line, "TurtleSilicon") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "/Volumes/") {
						mountPoint = line
						break
					}
				}
			}
		}
	}

	if mountPoint == "" {
		return "", "", fmt.Errorf("could not find mount point. Mount output: %s", string(mountOutput))
	}

	debug.Printf("Using mount point: %s", mountPoint)

	// Find the app in the mounted DMG
	var newAppPath string

	// First, try the exact name
	exactPath := filepath.Join(mountPoint, "TurtleSilicon.app")
	if PathExists(exactPath) {
		newAppPath = exactPath
	} else {
		// Search for any .app bundle in the mount point
		debug.Printf("TurtleSilicon.app not found at exact path, searching for .app bundles")
		entries, err := os.ReadDir(mountPoint)
		if err != nil {
			// Unmount before returning error
			exec.Command("hdiutil", "detach", mountPoint, "-force").Run()
			return "", "", fmt.Errorf("failed to read DMG contents: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
				candidatePath := filepath.Join(mountPoint, entry.Name())
				debug.Printf("Found .app bundle: %s", candidatePath)
				newAppPath = candidatePath
				break
			}
		}
	}

	// Unmount after testing
	debug.Printf("Unmounting test DMG from: %s", mountPoint)
	exec.Command("hdiutil", "detach", mountPoint, "-force").Run()

	if newAppPath == "" {
		return mountPoint, "", fmt.Errorf("no .app bundle found in DMG at %s", mountPoint)
	}

	return mountPoint, newAppPath, nil
}

// calculateResourceHash calculates the MD5 hash of a bundled resource
func calculateResourceHash(resource *fyne.StaticResource) string {
	hash := md5.New()
	hash.Write(resource.Content())
	return hex.EncodeToString(hash.Sum(nil))
}

// calculateFileHash calculates the MD5 hash of a local file
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CompareFileWithBundledResource compares both size and hash of a file on disk with a bundled resource
func CompareFileWithBundledResource(filePath, resourceName string) bool {
	// Check if the file exists first
	if !PathExists(filePath) {
		return false
	}

	// Get file info for the existing file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		debug.Printf("Failed to stat file %s: %v", filePath, err)
		return false
	}

	// Load the bundled resource
	resource, err := fyne.LoadResourceFromPath(resourceName)
	if err != nil {
		debug.Printf("Failed to load bundled resource %s: %v", resourceName, err)
		return false
	}

	// Compare sizes first (quick check)
	fileSize := fileInfo.Size()
	resourceSize := int64(len(resource.Content()))

	debug.Printf("Comparing file sizes - %s: %d bytes vs bundled %s: %d bytes",
		filePath, fileSize, resourceName, resourceSize)

	if fileSize != resourceSize {
		debug.Printf("Size mismatch for %s: file=%d, resource=%d", filePath, fileSize, resourceSize)
		return false
	}

	// Calculate and compare hashes for integrity verification
	fileHash, err := calculateFileHash(filePath)
	if err != nil {
		debug.Printf("Failed to calculate hash for %s: %v", filePath, err)
		return false
	}

	// Type assert to StaticResource for hash calculation
	staticResource, ok := resource.(*fyne.StaticResource)
	if !ok {
		debug.Printf("Failed to convert resource to StaticResource for %s", resourceName)
		return false
	}

	resourceHash := calculateResourceHash(staticResource)

	debug.Printf("Comparing file hashes - %s: %s vs bundled %s: %s",
		filePath, fileHash, resourceName, resourceHash)

	if fileHash != resourceHash {
		debug.Printf("Hash mismatch for %s: file=%s, resource=%s", filePath, fileHash, resourceHash)
		return false
	}

	debug.Printf("File verification successful for %s: size=%d, hash=%s", filePath, fileSize, fileHash)
	return true
}

// RestartApp restarts the application
func RestartApp() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Find the .app bundle
	appPath := execPath
	for !strings.HasSuffix(appPath, ".app") && appPath != "/" {
		appPath = filepath.Dir(appPath)
	}

	if strings.HasSuffix(appPath, ".app") {
		// Launch the app bundle
		cmd := exec.Command("open", appPath)
		return cmd.Start()
	} else {
		// Launch the executable directly
		cmd := exec.Command(execPath)
		return cmd.Start()
	}
}
