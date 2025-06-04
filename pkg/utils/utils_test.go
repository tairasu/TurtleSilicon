package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	// Test with a path that should exist (current directory)
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if !PathExists(currentDir) {
		t.Errorf("PathExists(%s) = false, want true", currentDir)
	}

	// Test with a path that should not exist
	nonExistentPath := filepath.Join(currentDir, "this-path-should-not-exist-12345")
	if PathExists(nonExistentPath) {
		t.Errorf("PathExists(%s) = true, want false", nonExistentPath)
	}
}
