package epochsilicon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// RequiredFile represents a file required for EpochSilicon
type RequiredFile struct {
	RelativePath string // Path relative to game directory
	DownloadURL  string
	DisplayName  string
}

// GetRequiredFiles returns the list of required files for EpochSilicon
func GetRequiredFiles() []RequiredFile {
	return []RequiredFile{
		{
			RelativePath: "Project-Epoch.exe",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=Project-Epoch.exe",
			DisplayName:  "Project-Epoch.exe",
		},
		{
			RelativePath: "ClientExtensions.dll",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=ClientExtensions.dll",
			DisplayName:  "ClientExtensions.dll",
		},
		{
			RelativePath: "Data/patch-A.MPQ",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=patch-A.MPQ",
			DisplayName:  "Data/patch-A.MPQ",
		},
		{
			RelativePath: "Data/patch-B.MPQ",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=patch-B.MPQ",
			DisplayName:  "Data/patch-B.MPQ",
		},
		{
			RelativePath: "Data/patch-Y.MPQ",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=patch-Y.MPQ",
			DisplayName:  "Data/patch-Y.MPQ",
		},
		{
			RelativePath: "Data/patch-Z.MPQ",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=patch-Z.MPQ",
			DisplayName:  "Data/patch-Z.MPQ",
		},
		{
			RelativePath: "Data/enUS/realmlist.wtf",
			DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=realmlist",
			DisplayName:  "Data/enUS/realmlist.wtf",
		},
	}
}

// CheckEpochSiliconFiles validates that all required EpochSilicon files exist
func CheckEpochSiliconFiles(gamePath string) ([]RequiredFile, error) {
	if gamePath == "" {
		return nil, fmt.Errorf("game path not set")
	}
	
	// First check if WoW.exe exists
	wowExePath := filepath.Join(gamePath, "WoW.exe")
	if !utils.PathExists(wowExePath) {
		return nil, fmt.Errorf("WoW.exe not found in %s. Please select a valid WoW directory", gamePath)
	}
	
	requiredFiles := GetRequiredFiles()
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

// ShowMissingFilesDialog displays a dialog asking the user if they want to download missing files
func ShowMissingFilesDialog(myWindow fyne.Window, missingFiles []RequiredFile, onDownload func()) {
	if len(missingFiles) == 0 {
		return
	}
	
	// Create the missing files list
	missingFilesList := widget.NewRichText()
	missingFilesText := "**Missing Project Epoch files:**\n\n"
	for _, file := range missingFiles {
		missingFilesText += fmt.Sprintf("• %s\n", file.DisplayName)
	}
	missingFilesText += "\nWould you like EpochSilicon to download these files for you?"
	missingFilesList.ParseMarkdown(missingFilesText)
	
	// Create content container
	content := container.NewVBox(
		missingFilesList,
		widget.NewSeparator(),
	)
	
	// Create custom dialog with Yes/No buttons
	confirmDialog := dialog.NewCustomConfirm(
		"Missing Project Epoch Files",
		"Yes, Download",
		"No, Cancel",
		content,
		func(download bool) {
			if download {
				onDownload()
			}
		},
		myWindow,
	)
	
	confirmDialog.Show()
}

// DownloadMissingFiles downloads the missing files with a progress dialog
func DownloadMissingFiles(myWindow fyne.Window, gamePath string, missingFiles []RequiredFile, onComplete func(bool)) {
	if len(missingFiles) == 0 {
		onComplete(true)
		return
	}
	
	// Create progress dialog
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)
	
	statusLabel := widget.NewLabel("Preparing download...")
	cancelButton := widget.NewButton("Cancel", nil)
	
	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Downloading Project Epoch Files"),
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		widget.NewSeparator(),
		container.NewCenter(cancelButton),
	)
	
	popup := widget.NewModalPopUp(container.NewPadded(content), myWindow.Canvas())
	popup.Resize(fyne.NewSize(500, 250))
	popup.Show()
	
	// Track cancellation
	cancelled := false
	cancelButton.OnTapped = func() {
		cancelled = true
		popup.Hide()
		onComplete(false)
	}
	
	// Download files in goroutine
	go func() {
		defer func() {
			if !cancelled {
				popup.Hide()
			}
		}()
		
		success := true
		for i, file := range missingFiles {
			if cancelled {
				return
			}
			
			// Update status
			fyne.Do(func() {
				statusLabel.SetText(fmt.Sprintf("Downloading %s... (%d/%d)", file.DisplayName, i+1, len(missingFiles)))
				progressBar.SetValue(float64(i) / float64(len(missingFiles)))
			})
			
			// Download file
			if err := downloadFile(gamePath, file); err != nil {
				debug.Printf("Failed to download %s: %v", file.DisplayName, err)
				fyne.Do(func() {
					popup.Hide()
					dialog.ShowError(fmt.Errorf("failed to download %s: %v", file.DisplayName, err), myWindow)
				})
				success = false
				break
			}
		}
		
		if success && !cancelled {
			// Always update realmlist.wtf even if it exists
			realmlistFile := RequiredFile{
				RelativePath: "Data/enUS/realmlist.wtf",
				DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=realmlist",
				DisplayName:  "Data/enUS/realmlist.wtf",
			}
			
			fyne.Do(func() {
				statusLabel.SetText("Updating realmlist.wtf...")
				progressBar.SetValue(0.9)
			})
			
			if err := downloadFile(gamePath, realmlistFile); err != nil {
				debug.Printf("Warning: Failed to update realmlist.wtf: %v", err)
			}
			
			fyne.Do(func() {
				progressBar.SetValue(1.0)
				statusLabel.SetText("Download complete!")
			})
		}
		
		// Complete callback
		fyne.Do(func() {
			onComplete(success && !cancelled)
		})
	}()
}

// downloadFile downloads a single file to the correct location
func downloadFile(gamePath string, file RequiredFile) error {
	fullPath := filepath.Join(gamePath, file.RelativePath)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}
	
	// Download file
	resp, err := http.Get(file.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", file.DownloadURL, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s", resp.StatusCode, file.DownloadURL)
	}
	
	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fullPath, err)
	}
	defer outFile.Close()
	
	// Copy data
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fullPath, err)
	}
	
	debug.Printf("Successfully downloaded: %s", file.DisplayName)
	return nil
}

// UpdateRealmlistForEpochSilicon always updates the realmlist.wtf file for EpochSilicon
func UpdateRealmlistForEpochSilicon(gamePath string) error {
	if gamePath == "" {
		return fmt.Errorf("game path not set")
	}
	
	realmlistFile := RequiredFile{
		RelativePath: "Data/enUS/realmlist.wtf",
		DownloadURL:  "http://updater.project-epoch.net/api/v2/latest?file=realmlist",
		DisplayName:  "Data/enUS/realmlist.wtf",
	}
	
	debug.Printf("Updating realmlist.wtf for EpochSilicon")
	return downloadFile(gamePath, realmlistFile)
}