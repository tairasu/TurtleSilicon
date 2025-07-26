package addons

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"turtlesilicon/pkg/debug"
	"turtlesilicon/pkg/paths"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Addon struct {
	Name         string
	Path         string
	HasGitRepo   bool
	GitRemoteURL string
	LastUpdated  time.Time
	Description  string
	LocalCommit  string
	RemoteCommit string
	NeedsUpdate  bool
}

type AddonManager struct {
	addons           []Addon
	window           fyne.Window
	currentPopup     *widget.PopUp
	showOnlyGit      bool
	addonsList       *container.Scroll
	contentContainer *fyne.Container
}

func NewAddonManager(window fyne.Window) *AddonManager {
	return &AddonManager{
		window: window,
	}
}

func (am *AddonManager) ScanAddons() error {
	return am.ScanAddonsWithProgress(nil)
}

func (am *AddonManager) ScanAddonsWithProgress(updateProgress func(string)) error {
	return am.scanAddonsWithProgress(updateProgress, false)
}

func (am *AddonManager) scanAddonsWithProgress(updateProgress func(string), checkUpdates bool) error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("game path not set")
	}

	addonsPath := filepath.Join(paths.TurtlewowPath, "Interface", "Addons")

	if _, err := os.Stat(addonsPath); os.IsNotExist(err) {
		return fmt.Errorf("addons directory not found: %s", addonsPath)
	}

	if updateProgress != nil {
		updateProgress("Scanning addon directories...")
	}

	entries, err := os.ReadDir(addonsPath)
	if err != nil {
		return fmt.Errorf("failed to read addons directory: %v", err)
	}

	am.addons = []Addon{}

	for i, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if updateProgress != nil {
			updateProgress(fmt.Sprintf("Processing %s (%d/%d)...", entry.Name(), i+1, len(entries)))
		}

		addonPath := filepath.Join(addonsPath, entry.Name())
		addon := Addon{
			Name: entry.Name(),
			Path: addonPath,
		}

		gitPath := filepath.Join(addonPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			addon.HasGitRepo = true
			addon.GitRemoteURL = am.getGitRemoteURL(addonPath)
			addon.LocalCommit = am.getLocalCommit(addonPath)

			if checkUpdates {
				if updateProgress != nil {
					updateProgress(fmt.Sprintf("Checking updates for %s...", entry.Name()))
				}

				addon.RemoteCommit = am.getRemoteCommit(addonPath)
				addon.NeedsUpdate = addon.LocalCommit != addon.RemoteCommit && addon.RemoteCommit != ""
			} else {
				addon.RemoteCommit = ""
				addon.NeedsUpdate = false
			}
		}

		if info, err := entry.Info(); err == nil {
			addon.LastUpdated = info.ModTime()
		}

		addon.Description = am.getAddonDescription(addonPath)
		am.addons = append(am.addons, addon)
	}

	if updateProgress != nil {
		updateProgress("Finalizing...")
	}

	if checkUpdates {
		debug.Printf("Found %d addons (%d with git repos, %d need updates)", len(am.addons), am.countGitAddons(), am.countUpdatableAddons())
	} else {
		debug.Printf("Found %d addons (%d with git repos)", len(am.addons), am.countGitAddons())
	}
	return nil
}

func (am *AddonManager) getGitRemoteURL(addonPath string) string {
	gitConfigPath := filepath.Join(addonPath, ".git", "config")
	if content, err := os.ReadFile(gitConfigPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, "[remote \"origin\"]") && i+1 < len(lines) {
				urlLine := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(urlLine, "url = ") {
					return strings.TrimPrefix(urlLine, "url = ")
				}
			}
		}
	}
	return ""
}

func (am *AddonManager) getAddonDescription(addonPath string) string {
	tocFiles, _ := filepath.Glob(filepath.Join(addonPath, "*.toc"))
	if len(tocFiles) > 0 {
		if content, err := os.ReadFile(tocFiles[0]); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "## Notes:") {
					return strings.TrimSpace(strings.TrimPrefix(line, "## Notes:"))
				}
			}
		}
	}
	return "No description available"
}

func (am *AddonManager) countGitAddons() int {
	count := 0
	for _, addon := range am.addons {
		if addon.HasGitRepo {
			count++
		}
	}
	return count
}

func (am *AddonManager) countUpdatableAddons() int {
	count := 0
	for _, addon := range am.addons {
		if addon.HasGitRepo && addon.NeedsUpdate {
			count++
		}
	}
	return count
}

func (am *AddonManager) getLocalCommit(addonPath string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = addonPath

	output, err := cmd.Output()
	if err != nil {
		debug.Printf("Failed to get local commit for %s: %v", addonPath, err)
		return ""
	}

	return strings.TrimSpace(string(output))
}

func (am *AddonManager) getRemoteCommit(addonPath string) string {
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = addonPath
	fetchCmd.Run()

	// Get the remote commit hash
	cmd := exec.Command("git", "rev-parse", "origin/HEAD")
	cmd.Dir = addonPath

	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "rev-parse", "origin/main")
		cmd.Dir = addonPath
		output, err = cmd.Output()
		if err != nil {
			cmd = exec.Command("git", "rev-parse", "origin/master")
			cmd.Dir = addonPath
			output, err = cmd.Output()
			if err != nil {
				debug.Printf("Failed to get remote commit for %s: %v", addonPath, err)
				return ""
			}
		}
	}

	return strings.TrimSpace(string(output))
}

func (am *AddonManager) UpdateAddon(addon *Addon) error {
	if !addon.HasGitRepo {
		return fmt.Errorf("addon %s does not have a git repository", addon.Name)
	}

	debug.Printf("Updating addon: %s", addon.Name)

	if err := am.runGitPull(addon.Path); err != nil {
		return fmt.Errorf("failed to update addon %s: %v", addon.Name, err)
	}

	// Update the directory's modification time to reflect the update
	now := time.Now()
	if err := os.Chtimes(addon.Path, now, now); err != nil {
		debug.Printf("Warning: failed to update directory modification time for %s: %v", addon.Name, err)
	}

	addon.LastUpdated = now
	debug.Printf("Successfully updated addon: %s", addon.Name)
	return nil
}

func (am *AddonManager) DeleteAddon(addon *Addon) error {
	debug.Printf("Deleting addon: %s", addon.Name)

	if err := os.RemoveAll(addon.Path); err != nil {
		return fmt.Errorf("failed to delete addon %s: %v", addon.Name, err)
	}

	for i, a := range am.addons {
		if a.Name == addon.Name {
			am.addons = append(am.addons[:i], am.addons[i+1:]...)
			break
		}
	}

	debug.Printf("Successfully deleted addon: %s", addon.Name)
	return nil
}

func (am *AddonManager) runGitPull(addonPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = addonPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		debug.Printf("Git pull failed: %s", string(output))
		return fmt.Errorf("git pull failed: %v", err)
	}

	debug.Printf("Git pull output: %s", string(output))
	return nil
}

func (am *AddonManager) ShowAddonManager() {
	if am.currentPopup != nil {
		am.currentPopup.Hide()
		am.currentPopup = nil
	}

	loadingText := widget.NewLabel("Initializing...")
	loadingText.Alignment = fyne.TextAlignCenter

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	loadingContent := container.NewVBox(
		widget.NewLabel("Loading Addons"),
		widget.NewSeparator(),
		loadingText,
		progressBar,
	)

	loadingPopup := widget.NewModalPopUp(container.NewPadded(loadingContent), am.window.Canvas())
	loadingPopup.Resize(fyne.NewSize(300, 150))
	loadingPopup.Show()

	go func() {
		// Create progress update function that updates the loading text
		updateProgress := func(message string) {
			fyne.Do(func() {
				loadingText.SetText(message)
			})
		}

		if err := am.ScanAddonsWithProgress(updateProgress); err != nil {
			fyne.Do(func() {
				progressBar.Stop()
				loadingPopup.Hide()
				dialog.ShowError(fmt.Errorf("failed to scan addons: %v", err), am.window)
			})
			return
		}

		fyne.Do(func() {
			progressBar.Stop()
			loadingPopup.Hide()
			am.createAddonManagerPopup()
		})
	}()
}

func (am *AddonManager) createAddonManagerPopup() {
	titleText := widget.NewLabel("Addon Manager")
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	summaryText := widget.NewLabel(fmt.Sprintf("Found %d addons (%d git repos, %d need updates)", len(am.addons), am.countGitAddons(), am.countUpdatableAddons()))
	summaryText.TextStyle = fyne.TextStyle{Italic: true}

	refreshButton := widget.NewButton("Refresh", func() {})
	refreshButton.Importance = widget.MediumImportance

	updateAllButton := widget.NewButton("Update All", func() {
		am.updateAllAddons()
	})
	updateAllButton.Importance = widget.HighImportance
	if am.countUpdatableAddons() == 0 {
		updateAllButton.Disable()
	}

	addButton := widget.NewButton("Add", func() {
		am.showAddAddonPopup()
	})
	addButton.Importance = widget.HighImportance

	onlyGitCheckbox := widget.NewCheck("Only GIT", func(checked bool) {
		am.showOnlyGit = checked
		am.refreshAddonList()
	})
	onlyGitCheckbox.SetChecked(am.showOnlyGit)

	leftSide := container.NewHBox(summaryText, onlyGitCheckbox)
	rightSide := container.NewHBox(addButton, refreshButton, updateAllButton)

	headerContainer := container.NewBorder(
		nil,
		widget.NewSeparator(),
		leftSide,
		rightSide,
		nil,
	)

	am.addonsList = am.createAddonsList()

	am.contentContainer = container.NewBorder(
		headerContainer,
		nil,
		nil,
		nil,
		am.addonsList,
	)

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge
	closeButton.Resize(fyne.NewSize(30, 30))

	// Create top bar with close button on left and title in center
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		container.NewCenter(titleText),
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		container.NewPadded(am.contentContainer),
	)

	popup := widget.NewModalPopUp(mainContainer, am.window.Canvas())

	canvasSize := am.window.Canvas().Size()
	popup.Resize(canvasSize)

	// Store reference to current popup
	am.currentPopup = popup

	// Add keyboard shortcut for Escape key
	canvas := am.window.Canvas()
	originalOnTypedKey := canvas.OnTypedKey()

	closeAction := func() {
		// Restore original key handler before closing
		canvas.SetOnTypedKey(originalOnTypedKey)
		popup.Hide()
		am.currentPopup = nil
	}

	closeButton.OnTapped = closeAction

	refreshButton.OnTapped = func() {
		am.refreshWithUpdateCheck()
	}

	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			closeAction()
			return
		}
		if originalOnTypedKey != nil {
			originalOnTypedKey(key)
		}
	})

	popup.Show()
}

func (am *AddonManager) createAddonsList() *container.Scroll {
	list := container.NewVBox()

	filteredAddons := am.getFilteredAddons()

	for i := range filteredAddons {
		addonCopy := filteredAddons[i]
		addonCard := am.createAddonCard(&addonCopy)
		list.Add(addonCard)
		if i < len(filteredAddons)-1 {
			list.Add(widget.NewSeparator())
		}
	}

	if len(filteredAddons) == 0 {
		var emptyMessage string
		if am.showOnlyGit {
			emptyMessage = "No GIT addons found"
		} else {
			emptyMessage = "No addons found in Interface/Addons directory"
		}
		emptyLabel := widget.NewLabel(emptyMessage)
		emptyLabel.Alignment = fyne.TextAlignCenter
		list.Add(emptyLabel)
	}

	return container.NewScroll(list)
}

func (am *AddonManager) getFilteredAddons() []Addon {
	if !am.showOnlyGit {
		return am.addons
	}

	var filteredAddons []Addon
	for _, addon := range am.addons {
		if addon.HasGitRepo {
			filteredAddons = append(filteredAddons, addon)
		}
	}
	return filteredAddons
}

func (am *AddonManager) refreshAddonList() {
	if am.currentPopup == nil || am.addonsList == nil {
		return
	}

	// Create new filtered list content
	newAddonsList := am.createAddonsList()

	// Replace the content of the existing scroll container
	am.addonsList.Content = newAddonsList.Content
	am.addonsList.Refresh()
}

func (am *AddonManager) refreshWithUpdateCheck() {
	if am.currentPopup != nil {
		am.currentPopup.Hide()
		am.currentPopup = nil
	}

	loadingText := widget.NewLabel("Initializing...")
	loadingText.Alignment = fyne.TextAlignCenter

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	loadingContent := container.NewVBox(
		widget.NewLabel("Refreshing Addons"),
		widget.NewSeparator(),
		loadingText,
		progressBar,
	)

	loadingPopup := widget.NewModalPopUp(container.NewPadded(loadingContent), am.window.Canvas())
	loadingPopup.Resize(fyne.NewSize(300, 150))
	loadingPopup.Show()

	go func() {
		updateProgress := func(message string) {
			fyne.Do(func() {
				loadingText.SetText(message)
			})
		}

		if err := am.scanAddonsWithProgress(updateProgress, true); err != nil {
			fyne.Do(func() {
				progressBar.Stop()
				loadingPopup.Hide()
				dialog.ShowError(fmt.Errorf("failed to refresh addons: %v", err), am.window)
			})
			return
		}

		fyne.Do(func() {
			progressBar.Stop()
			loadingPopup.Hide()
			am.createAddonManagerPopup()
		})
	}()
}

func (am *AddonManager) createAddonCard(addon *Addon) *fyne.Container {
	nameText := fmt.Sprintf("**%s**", addon.Name)

	nameLabel := widget.NewRichTextFromMarkdown(nameText)
	nameLabel.Wrapping = fyne.TextWrapOff

	lastUpdatedLabel := widget.NewLabel(fmt.Sprintf("Last modified: %s", addon.LastUpdated.Format("2006-01-02 15:04")))
	lastUpdatedLabel.TextStyle = fyne.TextStyle{Italic: true}

	nameContainer := container.NewHBox(nameLabel)

	textGroup := container.NewWithoutLayout(nameContainer, lastUpdatedLabel)
	nameContainer.Move(fyne.NewPos(0, 0))
	nameContainer.Resize(fyne.NewSize(400, 18))
	lastUpdatedLabel.Move(fyne.NewPos(0, 16))
	lastUpdatedLabel.Resize(fyne.NewSize(400, 14))
	textGroup.Resize(fyne.NewSize(400, 30))

	infoContainer := container.NewWithoutLayout(textGroup)
	textGroup.Move(fyne.NewPos(0, -5))
	infoContainer.Resize(fyne.NewSize(400, 50))

	var gitButton *widget.Button
	if addon.HasGitRepo {
		var gitText string
		if addon.NeedsUpdate {
			gitText = "GIT*"
		} else {
			gitText = "GIT"
		}

		gitButton = widget.NewButton(gitText, func() {})
		gitButton.Importance = widget.LowImportance
		gitButton.Disable()
	}

	var infoButton *widget.Button
	if addon.Description != "" && addon.Description != "No description available" {
		infoButton = widget.NewButton("Info", func() {
			am.showDescriptionPopup(addon.Name, addon.Description)
		})
		infoButton.Importance = widget.LowImportance
	}

	updateButton := widget.NewButton("Update", func() {
		am.updateSingleAddon(addon)
	})
	updateButton.Importance = widget.MediumImportance
	if !addon.HasGitRepo || !addon.NeedsUpdate {
		updateButton.Disable()
	}

	deleteButton := widget.NewButton("Delete", func() {
		am.confirmDeleteAddon(addon)
	})
	deleteButton.Importance = widget.MediumImportance

	var buttonsContainer *fyne.Container
	var buttons []fyne.CanvasObject

	if gitButton != nil {
		buttons = append(buttons, gitButton)
	}
	if infoButton != nil {
		buttons = append(buttons, infoButton)
	}
	buttons = append(buttons, updateButton, deleteButton)

	buttonsContainer = container.NewHBox(buttons...)

	buttonsWithMargin := container.NewPadded(buttonsContainer)

	cardContainer := container.NewBorder(
		nil,
		nil,
		infoContainer,
		buttonsWithMargin,
		nil,
	)

	cardContainer.Resize(fyne.NewSize(600, 70))

	return container.NewPadded(cardContainer)
}

func (am *AddonManager) updateSingleAddon(addon *Addon) {
	progressDialog := dialog.NewProgressInfinite("Updating addon", fmt.Sprintf("Updating %s...", addon.Name), am.window)
	progressDialog.Show()

	go func() {
		defer progressDialog.Hide()

		if err := am.UpdateAddon(addon); err != nil {
			dialog.ShowError(err, am.window)
		} else {
			dialog.ShowInformation("Success", fmt.Sprintf("Successfully updated %s", addon.Name), am.window)
			am.ShowAddonManager()
		}
	}()
}

func (am *AddonManager) updateAllAddons() {
	updatableAddons := []Addon{}
	for _, addon := range am.addons {
		if addon.HasGitRepo && addon.NeedsUpdate {
			updatableAddons = append(updatableAddons, addon)
		}
	}

	if len(updatableAddons) == 0 {
		dialog.ShowInformation("Info", "No addons need updates - all are up to date!", am.window)
		return
	}

	progressDialog := dialog.NewProgressInfinite("Updating addons", fmt.Sprintf("Updating %d addons...", len(updatableAddons)), am.window)
	progressDialog.Show()

	go func() {
		defer progressDialog.Hide()

		updated := 0
		failed := 0

		for i := range updatableAddons {
			if err := am.UpdateAddon(&updatableAddons[i]); err != nil {
				debug.Printf("Failed to update %s: %v", updatableAddons[i].Name, err)
				failed++
			} else {
				updated++
			}
		}

		message := fmt.Sprintf("Update complete!\nUpdated: %d\nFailed: %d", updated, failed)
		if failed > 0 {
			dialog.ShowInformation("Update Results", message, am.window)
		} else {
			dialog.ShowInformation("Success", message, am.window)
		}

		am.refreshAddonManager()
	}()
}

func (am *AddonManager) refreshAddonManager() {
	if am.currentPopup == nil {
		// No popup is currently open, so just show a new one
		am.ShowAddonManager()
		return
	}

	// Update the existing popup content
	go func() {
		// Re-scan addons
		if err := am.ScanAddons(); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to refresh addons: %v", err), am.window)
			})
			return
		}

		// Update the popup content on main thread
		fyne.Do(func() {
			// Get current popup size and position
			popupSize := am.currentPopup.Size()

			// Hide current popup
			am.currentPopup.Hide()

			// Create new popup with updated content
			am.createAddonManagerPopup()

			// Maintain the same size
			if am.currentPopup != nil {
				am.currentPopup.Resize(popupSize)
			}
		})
	}()
}

func (am *AddonManager) showDescriptionPopup(addonName, description string) {
	titleText := widget.NewRichTextFromMarkdown(fmt.Sprintf("# %s", addonName))
	titleText.Wrapping = fyne.TextWrapOff

	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	contentContainer := container.NewVBox(
		container.NewCenter(titleText),
		widget.NewSeparator(),
		descriptionLabel,
	)

	windowSize := am.window.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 2

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge

	// Create top bar with close button
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		nil,
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		contentContainer,
	)

	popup := widget.NewModalPopUp(mainContainer, am.window.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}

func (am *AddonManager) showAddAddonPopup() {
	titleText := widget.NewLabel("Add Addon from Repository")
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	instructionText := widget.NewLabel("Enter the GitHub or GitLab repository URL:")
	instructionText.TextStyle = fyne.TextStyle{Italic: true}

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://github.com/username/addon-name")

	// Create help button for finding repository URLs
	findAddonsButton := widget.NewButton("Where do I get repository URLs?", func() {
		parsedURL, err := url.Parse("https://turtle-wow.fandom.com/wiki/Addons#Full_Addons_List")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open addon wiki: %v", err), am.window)
			return
		}
		fyne.CurrentApp().OpenURL(parsedURL)
	})
	findAddonsButton.Importance = widget.LowImportance

	installButton := widget.NewButton("Install", func() {
		url := strings.TrimSpace(urlEntry.Text)
		if url == "" {
			dialog.ShowError(fmt.Errorf("please enter a repository URL"), am.window)
			return
		}
		am.installAddonFromRepo(url)
	})
	installButton.Importance = widget.HighImportance

	addMultipleButton := widget.NewButton("Add Multiple", func() {
		// This will be set when the popup is created
	})
	addMultipleButton.Importance = widget.MediumImportance

	contentContainer := container.NewVBox(
		titleText,
		widget.NewSeparator(),
		instructionText,
		urlEntry,
		widget.NewSeparator(),
		container.NewCenter(findAddonsButton),
		widget.NewSeparator(),
		container.NewHBox(installButton, addMultipleButton),
	)

	windowSize := am.window.Content().Size()
	popupWidth := windowSize.Width * 2 / 3
	popupHeight := windowSize.Height / 3

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge

	// Create top bar with close button
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		nil,
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		contentContainer,
	)

	popup := widget.NewModalPopUp(mainContainer, am.window.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	addMultipleButton.OnTapped = func() {
		popup.Hide()
		am.showAddMultipleAddonsPopup()
	}

	popup.Show()
}

func (am *AddonManager) showAddMultipleAddonsPopup() {
	titleText := widget.NewLabel("Add Multiple Addons")
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	instructionText := widget.NewLabel("Enter one repository URL per line:")
	instructionText.TextStyle = fyne.TextStyle{Italic: true}

	urlsEntry := widget.NewMultiLineEntry()
	urlsEntry.SetPlaceHolder("https://github.com/username/addon1\nhttps://github.com/username/addon2\nhttps://gitlab.com/username/addon3")
	urlsEntry.Resize(fyne.NewSize(500, 200))

	installAllButton := widget.NewButton("Install All", func() {
		urls := strings.Split(urlsEntry.Text, "\n")
		var validUrls []string

		for _, url := range urls {
			url = strings.TrimSpace(url)
			if url != "" {
				validUrls = append(validUrls, url)
			}
		}

		if len(validUrls) == 0 {
			dialog.ShowError(fmt.Errorf("please enter at least one repository URL"), am.window)
			return
		}

		am.installMultipleAddons(validUrls)
	})
	installAllButton.Importance = widget.HighImportance

	backButton := widget.NewButton("Back", func() {
		// This will be set when the popup is created
	})
	backButton.Importance = widget.MediumImportance

	contentContainer := container.NewVBox(
		titleText,
		widget.NewSeparator(),
		instructionText,
		urlsEntry,
		widget.NewSeparator(),
		container.NewHBox(backButton, installAllButton),
	)

	windowSize := am.window.Content().Size()
	popupWidth := windowSize.Width * 4 / 5
	popupHeight := windowSize.Height * 2 / 3

	// Create square close button without text padding
	closeButton := widget.NewButton("✕", func() {
		// This will be set when the popup is created
	})
	closeButton.Importance = widget.LowImportance

	// Force square dimensions by setting both min size and resize
	closeButton.Resize(fyne.NewSize(24, 24))
	closeButton.Move(fyne.NewPos(8, 8)) // Add small margin from edge

	// Create top bar with close button
	topBar := container.NewBorder(
		nil,
		nil,
		closeButton,
		nil,
		nil,
	)

	mainContainer := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		contentContainer,
	)

	popup := widget.NewModalPopUp(mainContainer, am.window.Canvas())
	popup.Resize(fyne.NewSize(popupWidth, popupHeight))

	closeButton.OnTapped = func() {
		popup.Hide()
	}

	backButton.OnTapped = func() {
		popup.Hide()
		am.showAddAddonPopup()
	}

	popup.Show()
}

func (am *AddonManager) installMultipleAddons(urls []string) {
	// Create loading dialog with updatable text
	loadingText := widget.NewLabel(fmt.Sprintf("Installing %d addons...", len(urls)))
	loadingText.Alignment = fyne.TextAlignCenter

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	loadingContent := container.NewVBox(
		widget.NewLabel("Installing Addons"),
		widget.NewSeparator(),
		loadingText,
		progressBar,
	)

	loadingPopup := widget.NewModalPopUp(container.NewPadded(loadingContent), am.window.Canvas())
	loadingPopup.Resize(fyne.NewSize(300, 150))
	loadingPopup.Show()

	go func() {
		defer func() {
			progressBar.Stop()
			loadingPopup.Hide()
		}()

		installed := 0
		failed := 0

		for i, url := range urls {
			// Update progress
			fyne.Do(func() {
				loadingText.SetText(fmt.Sprintf("Installing addon %d/%d...", i+1, len(urls)))
			})

			if err := am.cloneRepository(url); err != nil {
				debug.Printf("Failed to install %s: %v", url, err)
				failed++
			} else {
				installed++
			}
		}

		// Show results
		message := fmt.Sprintf("Installation complete!\nInstalled: %d\nFailed: %d", installed, failed)
		if failed > 0 {
			dialog.ShowInformation("Installation Results", message, am.window)
		} else {
			dialog.ShowInformation("Success", message, am.window)
		}

		// Refresh the addon manager to show new addons
		fyne.Do(func() {
			am.refreshAddonManager()
		})
	}()
}

func (am *AddonManager) installAddonFromRepo(repoURL string) {
	progressDialog := dialog.NewProgressInfinite("Installing addon", "Cloning repository...", am.window)
	progressDialog.Show()

	go func() {
		defer func() {
			fyne.Do(func() {
				progressDialog.Hide()
			})
		}()

		if err := am.cloneRepository(repoURL); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to install addon: %v", err), am.window)
			})
		} else {
			fyne.Do(func() {
				dialog.ShowInformation("Success", "Addon installed successfully!", am.window)
				am.refreshAddonManager()
			})
		}
	}()
}

func (am *AddonManager) cloneRepository(repoURL string) error {
	if paths.TurtlewowPath == "" {
		return fmt.Errorf("game path not set")
	}

	addonsPath := filepath.Join(paths.TurtlewowPath, "Interface", "Addons")

	// Extract addon name from URL
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repository URL")
	}
	addonName := parts[len(parts)-1]

	addonPath := filepath.Join(addonsPath, addonName)

	// Check if addon already exists
	if _, err := os.Stat(addonPath); err == nil {
		return fmt.Errorf("addon '%s' already exists", addonName)
	}

	debug.Printf("Cloning repository %s to %s", repoURL, addonPath)

	cmd := exec.Command("git", "clone", repoURL, addonPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		debug.Printf("Git clone failed: %s", string(output))
		return fmt.Errorf("git clone failed: %v", err)
	}

	debug.Printf("Successfully cloned repository: %s", string(output))
	return nil
}

func (am *AddonManager) confirmDeleteAddon(addon *Addon) {
	message := fmt.Sprintf("Are you sure you want to delete the addon '%s'?\n\nThis action cannot be undone.", addon.Name)

	confirmDialog := dialog.NewConfirm("Confirm Deletion", message, func(confirmed bool) {
		if confirmed {
			if err := am.DeleteAddon(addon); err != nil {
				dialog.ShowError(err, am.window)
			} else {
				dialog.ShowInformation("Success", fmt.Sprintf("Successfully deleted %s", addon.Name), am.window)
				am.refreshAddonManager()
			}
		}
	}, am.window)

	confirmDialog.Show()
}
