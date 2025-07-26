package paths

import (
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Global version manager instance will be set by main
var CurrentVersionManager interface{}

// Version-aware path selection functions
func SelectVersionGamePath(myWindow fyne.Window, versionID string, gamePathLabel *widget.RichText, updateAllStatuses func(), versionManager interface{}) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			log.Printf("Game path selection cancelled for version %s", versionID)
			updateAllStatuses()
			return
		}
		selectedPath := uri.Path()

		// Type assert the version manager to access its methods
		// This will be handled by the calling code to avoid circular dependencies

		log.Printf("Game path set for version %s: %s", versionID, selectedPath)
		updateAllStatuses()
	}, myWindow)
}

func SelectVersionCrossOverPath(myWindow fyne.Window, versionID string, crossoverPathLabel *widget.RichText, updateAllStatuses func(), versionManager interface{}) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if uri == nil {
			log.Printf("CrossOver path selection cancelled for version %s", versionID)
			updateAllStatuses()
			return
		}
		selectedPath := uri.Path()
		if filepath.Ext(selectedPath) == ".app" {
			// Type assert the version manager to access its methods
			// This will be handled by the calling code to avoid circular dependencies

			log.Printf("CrossOver path set for version %s: %s", versionID, selectedPath)
		} else {
			dialog.ShowError(fmt.Errorf("invalid selection: '%s'. Please select a valid .app bundle", selectedPath), myWindow)
			log.Printf("Invalid CrossOver path selected for version %s: %s", versionID, selectedPath)
		}
		updateAllStatuses()
	}, myWindow)
}

// Version-aware path status functions
func UpdateVersionPathLabels(versionID string, crossoverPathLabel, gamePathLabel *widget.RichText, versionManager interface{}) {
	// This will be implemented by the calling code to avoid circular dependencies
	// The version manager will be passed in to access path information
}

// Version-aware patching status
var VersionPatchingStatus = make(map[string]struct {
	GamePatched      bool
	CrossOverPatched bool
})

func GetVersionPatchingStatus(versionID string) (gamePatched bool, crossoverPatched bool) {
	status, exists := VersionPatchingStatus[versionID]
	if !exists {
		return false, false
	}
	return status.GamePatched, status.CrossOverPatched
}

func SetVersionPatchingStatus(versionID string, gamePatched bool, crossoverPatched bool) {
	VersionPatchingStatus[versionID] = struct {
		GamePatched      bool
		CrossOverPatched bool
	}{
		GamePatched:      gamePatched,
		CrossOverPatched: crossoverPatched,
	}
}
