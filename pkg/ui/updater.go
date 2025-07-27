package ui

import (
	"fmt"
	"strings"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowUpdateDialog displays an enhanced update dialog with download and install options
func ShowUpdateDialog(updateInfo *utils.UpdateInfo, currentVersion string, myWindow fyne.Window) {
	latestVersion := strings.TrimPrefix(updateInfo.TagName, "v")

	// Find the DMG asset
	var dmgAsset *utils.Asset
	for _, asset := range updateInfo.Assets {
		if strings.HasSuffix(asset.Name, ".dmg") {
			dmgAsset = &asset
			break
		}
	}

	if dmgAsset == nil {
		dialog.ShowError(fmt.Errorf("no DMG file found in the latest release"), myWindow)
		return
	}

	// Create content for the update dialog
	titleLabel := widget.NewLabel(fmt.Sprintf("Update Available: v%s", latestVersion))
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	currentVersionLabel := widget.NewLabel(fmt.Sprintf("Current version: v%s", currentVersion))

	// Format file size
	fileSize := formatFileSize(dmgAsset.Size)
	fileSizeLabel := widget.NewLabel(fmt.Sprintf("Download size: %s", fileSize))

	// Release notes
	notesLabel := widget.NewLabel("Release notes:")
	notesLabel.TextStyle = fyne.TextStyle{Bold: true}
	notesText := widget.NewRichTextFromMarkdown(updateInfo.Body)
	notesScroll := container.NewScroll(notesText)
	notesScroll.SetMinSize(fyne.NewSize(480, 120))

	// Progress bar (initially hidden)
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Hide()

	detailLabel := widget.NewLabel("")
	detailLabel.Hide()

	// Checkbox for suppressing this version
	suppressCheck := widget.NewCheck("Don't show this update again", nil)

	content := container.NewVBox(
		titleLabel,
		currentVersionLabel,
		fileSizeLabel,
		widget.NewSeparator(),
		notesLabel,
		notesScroll,
		widget.NewSeparator(),
		statusLabel,
		detailLabel,
		progressBar,
		suppressCheck,
	)

	windowSize := myWindow.Canvas().Size()
	dialogWidth := float32(windowSize.Width) * 4.0 / 5.0
	dialogHeight := float32(windowSize.Height) * 4.0 / 5.0

	d := dialog.NewCustom("New Update Available", "", content, myWindow)
	d.Resize(fyne.NewSize(dialogWidth, dialogHeight))

	// Download and install function
	downloadAndInstall := func() {
		// Show progress elements
		progressBar.Show()
		statusLabel.Show()
		detailLabel.Show()
		statusLabel.SetText("Preparing download...")
		detailLabel.SetText("")

		// Disable dialog closing during download
		d.SetButtons([]fyne.CanvasObject{})

		go func() {
			// Update status
			fyne.Do(func() {
				statusLabel.SetText("Downloading TurtleSilicon update...")
				detailLabel.SetText("0.0 MB / " + formatFileSize(dmgAsset.Size))
			})

			// Download with progress
			downloadPath, err := utils.DownloadUpdate(dmgAsset.BrowserDownloadURL, func(downloaded, total int64) {
				// Update progress on UI thread
				fyne.Do(func() {
					if total > 0 {
						progress := float64(downloaded) / float64(total)
						progressBar.SetValue(progress)
						detailLabel.SetText(fmt.Sprintf("%.1f MB / %.1f MB",
							float64(downloaded)/(1024*1024), float64(total)/(1024*1024)))
					}
				})
			})

			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Download failed!")
					detailLabel.SetText(fmt.Sprintf("Error: %v", err))
				})
				debug.Printf("Download failed: %v", err)

				// Re-enable close button
				d.SetButtons([]fyne.CanvasObject{
					widget.NewButton("Close", func() { d.Hide() }),
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("Installing update...")
				detailLabel.SetText("Update downloaded successfully")
				progressBar.SetValue(1.0)
			})

			// Install update
			err = utils.InstallUpdate(downloadPath)
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Installation failed!")
					detailLabel.SetText(fmt.Sprintf("Error: %v", err))
				})
				debug.Printf("Installation failed: %v", err)

				// Re-enable close button
				d.SetButtons([]fyne.CanvasObject{
					widget.NewButton("Close", func() { d.Hide() }),
				})
				return
			}

			// Success - show restart dialog
			fyne.Do(func() {
				statusLabel.SetText("Update installed successfully!")
				detailLabel.SetText("All files updated successfully")
			})

			restartDialog := dialog.NewConfirm(
				"Update Complete",
				"The update has been installed successfully and will require a restart. Would you like to close the application now?",
				func(restart bool) {
					d.Hide()
					if restart {
						utils.RestartApp()
						fyne.CurrentApp().Quit()
					}
				},
				myWindow,
			)
			restartDialog.Show()
		}()
	}

	// Set dialog buttons
	d.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Download & Install", downloadAndInstall),
		widget.NewButton("Later", func() {
			if suppressCheck.Checked {
				// Save suppressed version
				prefs, _ := utils.LoadPrefs()
				prefs.SuppressedUpdateVersion = latestVersion
				utils.SavePrefs(prefs)
			}
			d.Hide()
		}),
	})

	d.Show()
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
