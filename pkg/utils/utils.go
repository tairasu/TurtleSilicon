package utils

import (
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

// GetBundledResourceSize returns the size of a bundled resource
func GetBundledResourceSize(resourcePath string) (int64, error) {
	resource, err := fyne.LoadResourceFromPath(resourcePath)
	if err != nil {
		return 0, fmt.Errorf("failed to load bundled resource %s: %v", resourcePath, err)
	}
	return int64(len(resource.Content())), nil
}

// CompareFileWithBundledResource compares the size of a file with a bundled resource
func CompareFileWithBundledResource(filePath, resourcePath string) bool {
	if !PathExists(filePath) {
		return false
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		debug.Printf("Failed to get file info for %s: %v", filePath, err)
		return false
	}
	fileSize := fileInfo.Size()

	// Get bundled resource size
	resourceSize, err := GetBundledResourceSize(resourcePath)
	if err != nil {
		debug.Printf("Failed to get bundled resource size for %s: %v", resourcePath, err)
		return false
	}

	debug.Printf("Comparing file sizes: %s (%d bytes) vs %s (%d bytes)", filePath, fileSize, resourcePath, resourceSize)
	return fileSize == resourceSize
}
