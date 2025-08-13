package epochsilicon

import (
	"fmt"

	"turtlesilicon/pkg/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

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

			if file.Size > 0 {
				fileSizes[file.RelativePath] = file.Size
				totalSize += file.Size
			} else if metadata, err := getRemoteFileMetadata(file.DownloadURL); err == nil && metadata.Size > 0 {
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

			// Download file with progress and verification
			if err := downloadFileWithVerification(gamePath, file, func(downloaded, total int64) {
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
