package epochsilicon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

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

// FileMetadata represents metadata for checking file updates
type FileMetadata struct {
	Size         int64
	LastModified time.Time
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

	// Get window size to calculate dialog size (5/6 of window)
	windowSize := myWindow.Canvas().Size()
	dialogWidth := float32(windowSize.Width) * 5.0 / 6.0
	dialogHeight := float32(windowSize.Height) * 5.0 / 6.0

	// Create title
	title := widget.NewRichTextFromMarkdown("# Missing Project Epoch Files")
	title.Resize(fyne.NewSize(dialogWidth-40, 50))

	// Create header text
	header := widget.NewRichTextFromMarkdown("**The following Project Epoch files are missing and need to be downloaded:**")

	// Create file list as simple labels
	var fileLabels []fyne.CanvasObject
	for _, file := range missingFiles {
		label := widget.NewLabel("• " + file.DisplayName)
		label.TextStyle = fyne.TextStyle{Monospace: true}
		fileLabels = append(fileLabels, label)
	}

	// Create container for file list
	fileListContainer := container.NewVBox(fileLabels...)

	// Create question
	question := widget.NewRichTextFromMarkdown("**Would you like EpochSilicon to download these files for you?**")

	// Create buttons
	downloadButton := widget.NewButton("Yes, Download Files", nil)
	downloadButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("No, Cancel", nil)

	buttonContainer := container.NewHBox(
		downloadButton,
		widget.NewSeparator(),
		cancelButton,
	)

	// Create content container
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		header,
		widget.NewCard("", "", fileListContainer),
		widget.NewSeparator(),
		question,
		widget.NewSeparator(),
		container.NewCenter(buttonContainer),
	)

	// Create modal popup with custom size
	popup := widget.NewModalPopUp(
		container.NewPadded(content),
		myWindow.Canvas(),
	)
	popup.Resize(fyne.NewSize(dialogWidth, dialogHeight))

	// Set up button actions
	downloadButton.OnTapped = func() {
		popup.Hide()
		onDownload()
	}

	cancelButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
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
	detailLabel := widget.NewLabel("")
	cancelButton := widget.NewButton("Cancel", nil)

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Downloading Project Epoch Files"),
		widget.NewSeparator(),
		statusLabel,
		detailLabel,
		progressBar,
		widget.NewSeparator(),
		container.NewCenter(cancelButton),
	)

	popup := widget.NewModalPopUp(container.NewPadded(content), myWindow.Canvas())
	popup.Resize(fyne.NewSize(500, 280))
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

		// First, get total sizes for accurate progress
		fyne.Do(func() {
			statusLabel.SetText("Calculating download size...")
		})

		var totalSize int64
		fileSizes := make(map[string]int64)

		for _, file := range missingFiles {
			if cancelled {
				return
			}

			if metadata, err := getRemoteFileMetadata(file.DownloadURL); err == nil && metadata.Size > 0 {
				fileSizes[file.RelativePath] = metadata.Size
				totalSize += metadata.Size
			} else {
				// Fallback to estimated size if we can't get actual size
				fileSizes[file.RelativePath] = 50 * 1024 * 1024 // 50MB estimate
				totalSize += 50 * 1024 * 1024
			}
		}

		var totalDownloaded int64
		success := true

		for i, file := range missingFiles {
			if cancelled {
				return
			}

			fileSize := fileSizes[file.RelativePath]

			// Update status
			fyne.Do(func() {
				statusLabel.SetText(fmt.Sprintf("Downloading %s... (%d/%d)", file.DisplayName, i+1, len(missingFiles)))
				detailLabel.SetText(fmt.Sprintf("%.1f MB / %.1f MB", float64(totalDownloaded)/(1024*1024), float64(totalSize)/(1024*1024)))
			})

			// Download file with progress
			if err := downloadFileWithProgress(gamePath, file, func(downloaded, total int64) {
				currentTotal := totalDownloaded + downloaded
				fyne.Do(func() {
					if totalSize > 0 {
						progressBar.SetValue(float64(currentTotal) / float64(totalSize))
					}
					detailLabel.SetText(fmt.Sprintf("%.1f MB / %.1f MB", float64(currentTotal)/(1024*1024), float64(totalSize)/(1024*1024)))
				})
			}); err != nil {
				debug.Printf("Failed to download %s: %v", file.DisplayName, err)
				fyne.Do(func() {
					popup.Hide()
					dialog.ShowError(fmt.Errorf("failed to download %s: %v", file.DisplayName, err), myWindow)
				})
				success = false
				break
			}

			totalDownloaded += fileSize
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
				progressBar.SetValue(0.95)
			})

			if err := downloadFile(gamePath, realmlistFile); err != nil {
				debug.Printf("Warning: Failed to update realmlist.wtf: %v", err)
			}

			fyne.Do(func() {
				progressBar.SetValue(1.0)
				statusLabel.SetText("Download complete!")
				detailLabel.SetText("All files downloaded successfully")
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

// ProgressWriter wraps an io.Writer and reports progress
type ProgressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	onProgress func(downloaded, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.downloaded += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.downloaded, pw.total)
	}
	return
}

// downloadFileWithProgress downloads a file and reports progress
func downloadFileWithProgress(gamePath string, file RequiredFile, onProgress func(downloaded, total int64)) error {
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

	// Get content length for progress tracking
	contentLength := resp.ContentLength

	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", fullPath, err)
	}
	defer outFile.Close()

	// Create progress writer
	progressWriter := &ProgressWriter{
		writer:     outFile,
		total:      contentLength,
		onProgress: onProgress,
	}

	// Copy data with progress tracking
	_, err = io.Copy(progressWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", fullPath, err)
	}

	debug.Printf("Successfully downloaded: %s", file.DisplayName)
	return nil
}

// getRemoteFileMetadata gets the metadata (size and last-modified) of a remote file
func getRemoteFileMetadata(url string) (*FileMetadata, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d for %s", resp.StatusCode, url)
	}

	metadata := &FileMetadata{}

	// Get file size
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			metadata.Size = size
		}
	}

	// Get last modified date
	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if modTime, err := time.Parse(time.RFC1123, lastModified); err == nil {
			metadata.LastModified = modTime
		}
	}

	return metadata, nil
}

// getLocalFileMetadata gets the metadata of a local file
func getLocalFileMetadata(filePath string) (*FileMetadata, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
	}, nil
}

// fileUpdateResult represents the result of checking a single file for updates
type fileUpdateResult struct {
	file        RequiredFile
	needsUpdate bool
	err         error
}

// CheckForUpdates checks if any files need updating based on remote metadata
func CheckForUpdates(gamePath string) ([]RequiredFile, error) {
	if gamePath == "" {
		return nil, fmt.Errorf("game path not set")
	}

	requiredFiles := GetRequiredFiles()
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

			// Get local file metadata
			localMeta, err := getLocalFileMetadata(fullPath)
			if err != nil {
				result.err = fmt.Errorf("failed to get local metadata: %v", err)
				debug.Printf("Failed to get local metadata for %s: %v", f.RelativePath, err)
				resultChan <- result
				return
			}

			// Get remote file metadata
			remoteMeta, err := getRemoteFileMetadata(f.DownloadURL)
			if err != nil {
				result.err = fmt.Errorf("failed to get remote metadata: %v", err)
				debug.Printf("Failed to get remote metadata for %s: %v", f.RelativePath, err)
				resultChan <- result
				return
			}

			// Compare metadata to determine if update is needed
			needsUpdate := false

			// Check if sizes differ (most reliable indicator)
			if remoteMeta.Size > 0 && localMeta.Size != remoteMeta.Size {
				needsUpdate = true
				debug.Printf("Size mismatch for %s: local=%d, remote=%d", f.RelativePath, localMeta.Size, remoteMeta.Size)
			}

			// Check if remote file is newer (if we have both timestamps)
			if !remoteMeta.LastModified.IsZero() && !localMeta.LastModified.IsZero() {
				if remoteMeta.LastModified.After(localMeta.LastModified) {
					needsUpdate = true
					debug.Printf("Remote file newer for %s: local=%v, remote=%v", f.RelativePath, localMeta.LastModified, remoteMeta.LastModified)
				}
			}

			result.needsUpdate = needsUpdate
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

// CheckForUpdatesWithProgress checks for updates and shows a progress dialog
func CheckForUpdatesWithProgress(myWindow fyne.Window, gamePath string, onComplete func([]RequiredFile, error)) {
	// Create loading dialog
	progressSpinner := widget.NewProgressBarInfinite()
	progressSpinner.Start()

	statusLabel := widget.NewLabel("Checking for Project Epoch file updates...")
	cancelButton := widget.NewButton("Cancel", nil)

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Checking for Updates"),
		widget.NewSeparator(),
		statusLabel,
		progressSpinner,
		widget.NewSeparator(),
		container.NewCenter(cancelButton),
	)

	popup := widget.NewModalPopUp(container.NewPadded(content), myWindow.Canvas())
	popup.Resize(fyne.NewSize(400, 200))
	popup.Show()

	// Track cancellation
	cancelled := false
	cancelButton.OnTapped = func() {
		cancelled = true
		popup.Hide()
		onComplete(nil, fmt.Errorf("update check cancelled"))
	}

	// Check for updates in goroutine
	go func() {
		defer func() {
			progressSpinner.Stop()
			if !cancelled {
				popup.Hide()
			}
		}()

		if cancelled {
			return
		}

		updatesAvailable, err := CheckForUpdates(gamePath)

		if !cancelled {
			fyne.Do(func() {
				onComplete(updatesAvailable, err)
			})
		}
	}()
}

// ShowUpdatePromptDialog displays a dialog asking if the user wants to update files
func ShowUpdatePromptDialog(myWindow fyne.Window, updatesAvailable []RequiredFile, onUpdate func()) {
	if len(updatesAvailable) == 0 {
		return
	}

	// Get window size to calculate dialog size (5/6 of window)
	windowSize := myWindow.Canvas().Size()
	dialogWidth := float32(windowSize.Width) * 5.0 / 6.0
	dialogHeight := float32(windowSize.Height) * 5.0 / 6.0

	// Create title
	title := widget.NewRichTextFromMarkdown("# Project Epoch Updates Available")
	title.Resize(fyne.NewSize(dialogWidth-40, 50))

	// Create header text
	header := widget.NewRichTextFromMarkdown("**The following Project Epoch files have updates available:**")

	// Create file list as simple labels
	var fileLabels []fyne.CanvasObject
	for _, file := range updatesAvailable {
		label := widget.NewLabel("• " + file.DisplayName)
		label.TextStyle = fyne.TextStyle{Monospace: true}
		fileLabels = append(fileLabels, label)
	}

	// Create container for file list
	fileListContainer := container.NewVBox(fileLabels...)

	// Create question
	question := widget.NewRichTextFromMarkdown("**Would you like to download the updates?**")

	// Create buttons
	updateButton := widget.NewButton("Yes, Update Files", nil)
	updateButton.Importance = widget.HighImportance

	skipButton := widget.NewButton("No, Skip Updates", nil)

	buttonContainer := container.NewHBox(
		updateButton,
		widget.NewSeparator(),
		skipButton,
	)

	// Create content container
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		header,
		widget.NewCard("", "", fileListContainer),
		widget.NewSeparator(),
		question,
		widget.NewSeparator(),
		container.NewCenter(buttonContainer),
	)

	// Create modal popup with custom size
	popup := widget.NewModalPopUp(
		container.NewPadded(content),
		myWindow.Canvas(),
	)
	popup.Resize(fyne.NewSize(dialogWidth, dialogHeight))

	// Set up button actions
	updateButton.OnTapped = func() {
		popup.Hide()
		onUpdate()
	}

	skipButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
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
