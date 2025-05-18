package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	log.Printf("Executing AppleScript: %s", scriptString)
	cmd := exec.Command("osascript", "-e", scriptString)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("AppleScript failed: %v\nOutput: %s", err, string(output))
		dialog.ShowError(fmt.Errorf(errMsg), myWindow)
		log.Println(errMsg)
		return false
	}
	log.Printf("osascript output: %s", string(output))
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

