package service

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"turtlesilicon/pkg/paths"
	"turtlesilicon/pkg/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	ServiceRunning = false
	serviceCmd     *exec.Cmd
	servicePID     int
)

// CleanupExistingServices kills any existing rosettax87 processes
func CleanupExistingServices() error {
	log.Println("Cleaning up any existing rosettax87 processes...")

	// Find all rosettax87 processes
	cmd := exec.Command("pgrep", "-f", "rosettax87")
	output, err := cmd.Output()
	if err != nil {
		// No processes found, that's fine
		return nil
	}

	pids := strings.Fields(strings.TrimSpace(string(output)))
	if len(pids) == 0 {
		return nil
	}

	// Kill all found processes without sudo (try regular kill first)
	for _, pid := range pids {
		// First try regular kill
		killCmd := exec.Command("kill", "-9", pid)
		err := killCmd.Run()
		if err != nil {
			// If regular kill fails, try with sudo (but this might fail too)
			log.Printf("Regular kill failed for process %s, trying sudo: %v", pid, err)
			sudoKillCmd := exec.Command("sudo", "kill", "-9", pid)
			err2 := sudoKillCmd.Run()
			if err2 != nil {
				log.Printf("Failed to kill process %s with sudo: %v", pid, err2)
			} else {
				log.Printf("Killed existing rosettax87 process with sudo: %s", pid)
			}
		} else {
			log.Printf("Killed existing rosettax87 process: %s", pid)
		}
	}

	// Wait a moment for processes to die
	time.Sleep(1 * time.Second)
	return nil
}

// isRosettaSocketActive checks if the rosetta helper socket is active
func isRosettaSocketActive() bool {
	// Check if the socket file exists and is accessible
	cmd := exec.Command("ls", "-la", "/var/run/rosetta_helper.sock")
	err := cmd.Run()
	return err == nil
}

// StartRosettaX87Service starts the RosettaX87 service with sudo privileges
func StartRosettaX87Service(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Starting RosettaX87 service...")

	if paths.TurtlewowPath == "" {
		dialog.ShowError(fmt.Errorf("TurtleWoW path not set. Please set it first"), myWindow)
		return
	}

	rosettaX87Dir := filepath.Join(paths.TurtlewowPath, "rosettax87")
	rosettaX87Exe := filepath.Join(rosettaX87Dir, "rosettax87")

	if !utils.PathExists(rosettaX87Exe) {
		dialog.ShowError(fmt.Errorf("rosettax87 executable not found at %s. Please apply TurtleWoW patches first", rosettaX87Exe), myWindow)
		return
	}

	if ServiceRunning {
		dialog.ShowInformation("Service Status", "RosettaX87 service is already running.", myWindow)
		return
	}

	// Clean up any existing rosettax87 processes first
	CleanupExistingServices()

	// Load user preferences
	prefs, err := utils.LoadPrefs()
	if err != nil {
		log.Printf("Failed to load preferences: %v", err)
		prefs = &utils.UserPrefs{} // Use default prefs
	}

	// Try to get saved password if the user has enabled saving
	var savedPassword string
	if prefs.SaveSudoPassword {
		savedPassword, _ = utils.GetSudoPassword() // Ignore errors, just use empty string
	}

	// Show password dialog
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter your sudo password")
	passwordEntry.SetText(savedPassword) // Prefill with saved password if available
	passwordEntry.Resize(fyne.NewSize(300, passwordEntry.MinSize().Height))

	// Create checkbox for saving password
	savePasswordCheck := widget.NewCheck("Save password securely in keychain", nil)
	savePasswordCheck.SetChecked(prefs.SaveSudoPassword)

	// Add status label if password is already saved
	var statusLabel *widget.Label
	if utils.HasSavedSudoPassword() {
		statusLabel = widget.NewLabel("✓ Password already saved in keychain")
		statusLabel.Importance = widget.LowImportance
	}

	// Create a container with proper sizing
	passwordForm := widget.NewForm(widget.NewFormItem("Password:", passwordEntry))

	var containerItems []fyne.CanvasObject
	containerItems = append(containerItems,
		widget.NewLabel("Enter your sudo password to start the RosettaX87 service:"),
		passwordForm,
		savePasswordCheck,
	)

	if statusLabel != nil {
		containerItems = append(containerItems, statusLabel)
	}

	passwordContainer := container.NewVBox(containerItems...)
	passwordContainer.Resize(fyne.NewSize(400, 140))

	// Create the dialog variable so we can reference it in the callback
	var passwordDialog dialog.Dialog

	// Define the confirm logic as a function so it can be reused
	confirmFunc := func() {
		password := passwordEntry.Text
		if password == "" {
			dialog.ShowError(fmt.Errorf("password cannot be empty"), myWindow)
			return
		}

		// Handle password saving/deleting based on checkbox state
		shouldSavePassword := savePasswordCheck.Checked
		if shouldSavePassword {
			// Save password to keychain
			if err := utils.SaveSudoPassword(password); err != nil {
				log.Printf("Failed to save password to keychain: %v", err)
				// Don't block the service start, just log the error
			}
		} else {
			// Delete any existing saved password
			utils.DeleteSudoPassword() // Ignore errors
		}

		// Update preferences
		prefs.SaveSudoPassword = shouldSavePassword
		if err := utils.SavePrefs(prefs); err != nil {
			log.Printf("Failed to save preferences: %v", err)
		}

		// Close the dialog
		passwordDialog.Hide()

		// Set starting state
		paths.ServiceStarting = true
		fyne.Do(func() {
			updateAllStatuses()
		})

		// Start the service in a goroutine
		go func() {
			err := startServiceWithPassword(rosettaX87Dir, rosettaX87Exe, password)
			paths.ServiceStarting = false
			if err != nil {
				log.Printf("Failed to start RosettaX87 service: %v", err)
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to start RosettaX87 service: %v", err), myWindow)
				})
				ServiceRunning = false
			} else {
				log.Println("RosettaX87 service started successfully")
				ServiceRunning = true
			}
			fyne.Do(func() {
				updateAllStatuses()
			})
		}()
	}

	// Add Enter key support to password entry
	passwordEntry.OnSubmitted = func(text string) {
		confirmFunc()
	}

	passwordDialog = dialog.NewCustomConfirm("Sudo Password Required", "Start Service", "Cancel",
		passwordContainer,
		func(confirmed bool) {
			if !confirmed {
				log.Println("Service start cancelled by user")
				return
			}
			confirmFunc()
		}, myWindow)

	passwordDialog.Show()

	// Focus the password entry after showing the dialog
	myWindow.Canvas().Focus(passwordEntry)
}

// startServiceWithPassword starts the service using sudo with the provided password
func startServiceWithPassword(workingDir, executable, password string) error {
	// First clear any existing sudo credentials to ensure fresh authentication
	clearCmd := exec.Command("sudo", "-k")
	clearCmd.Run() // Ignore errors

	// Test the password with a simple command that requires sudo
	testCmd := exec.Command("sudo", "-S", "echo", "test")
	testStdin, err := testCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create test stdin pipe: %v", err)
	}

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	testCmd.Stdout = &stdout
	testCmd.Stderr = &stderr

	err = testCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start test command: %v", err)
	}

	// Send the password for testing
	_, err = testStdin.Write([]byte(password + "\n"))
	if err != nil {
		testCmd.Process.Kill()
		return fmt.Errorf("failed to send test password: %v", err)
	}
	testStdin.Close()

	// Wait for the test command to complete
	err = testCmd.Wait()
	stderrOutput := stderr.String()
	stdoutOutput := stdout.String()

	log.Printf("Password test - Exit code: %v, Stderr: %q, Stdout: %q", err, stderrOutput, stdoutOutput)

	// Check for authentication failure indicators
	if strings.Contains(stderrOutput, "Sorry, try again") ||
		strings.Contains(stderrOutput, "incorrect password") ||
		strings.Contains(stderrOutput, "authentication failure") ||
		strings.Contains(stderrOutput, "1 incorrect password attempt") {
		return fmt.Errorf("incorrect password")
	}

	if err != nil {
		return fmt.Errorf("sudo authentication failed: %v, stderr: %s", err, stderrOutput)
	}

	// Additional check: the stdout should contain "test" if the command succeeded
	if !strings.Contains(stdoutOutput, "test") {
		return fmt.Errorf("password authentication failed - no expected output")
	}

	// If we get here, the password is correct, now start the actual service
	log.Println("Password validated successfully, starting rosettax87 service...")

	cmd := exec.Command("sudo", "-S", executable)
	cmd.Dir = workingDir

	// Capture both stdout and stderr for debugging (reuse existing variables)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create a pipe to send the password
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Send the password
	_, err = stdin.Write([]byte(password + "\n"))
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to send password: %v", err)
	}
	stdin.Close()

	// Store the command and PID for later termination
	serviceCmd = cmd
	servicePID = cmd.Process.Pid

	// Wait a moment to see if the process starts successfully
	time.Sleep(3 * time.Second)

	// Check if the process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		stderrOutput := stderr.String()
		stdoutOutput := stdout.String()
		log.Printf("Process exited - Stdout: %q, Stderr: %q", stdoutOutput, stderrOutput)
		return fmt.Errorf("process exited prematurely with code: %d. Stderr: %s", cmd.ProcessState.ExitCode(), stderrOutput)
	}

	// Verify the service is actually listening
	time.Sleep(1 * time.Second)
	if !isRosettaSocketActive() {
		log.Printf("Service started but socket not active - Stdout: %q, Stderr: %q", stdout.String(), stderr.String())
		cmd.Process.Kill()
		return fmt.Errorf("service started but is not listening on socket")
	}

	log.Printf("RosettaX87 service started successfully with PID: %d", servicePID)
	return nil
}

// StopRosettaX87Service stops the running RosettaX87 service
func StopRosettaX87Service(myWindow fyne.Window, updateAllStatuses func()) {
	log.Println("Stopping RosettaX87 service...")

	if !ServiceRunning {
		dialog.ShowInformation("Service Status", "RosettaX87 service is not running.", myWindow)
		return
	}

	if serviceCmd != nil && serviceCmd.Process != nil {
		// Send SIGTERM to gracefully stop the process
		err := serviceCmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("Failed to send SIGTERM to process: %v", err)
			// Try SIGKILL as fallback
			err = serviceCmd.Process.Kill()
			if err != nil {
				log.Printf("Failed to kill process: %v", err)
				dialog.ShowError(fmt.Errorf("failed to stop service: %v", err), myWindow)
				return
			}
		}

		// Wait for the process to exit
		go func() {
			serviceCmd.Wait()
			ServiceRunning = false
			serviceCmd = nil
			servicePID = 0
			log.Println("RosettaX87 service stopped")
			fyne.Do(func() {
				dialog.ShowInformation("Service Stopped", "RosettaX87 service has been stopped.", myWindow)
				updateAllStatuses()
			})
		}()
	} else {
		ServiceRunning = false
		updateAllStatuses()
	}
}

// IsServiceRunning checks if the RosettaX87 service is currently running
func IsServiceRunning() bool {
	// Check for any rosettax87 process running system-wide
	cmd := exec.Command("pgrep", "-f", "rosettax87")
	output, err := cmd.Output()
	if err != nil {
		// No rosettax87 process found
		ServiceRunning = false
		serviceCmd = nil
		servicePID = 0
		return false
	}

	// If there are rosettax87 processes running, the service is considered running
	if len(strings.TrimSpace(string(output))) > 0 {
		ServiceRunning = true
		return true
	}

	ServiceRunning = false
	return false
}

// CleanupService ensures the service is stopped when the application exits
func CleanupService() {
	log.Println("Cleaning up RosettaX87 service on application exit...")
	CleanupExistingServices()
	ServiceRunning = false
	serviceCmd = nil
	servicePID = 0
}

// ClearSavedPassword removes the saved password and shows a confirmation dialog
func ClearSavedPassword(myWindow fyne.Window) {
	if !utils.HasSavedSudoPassword() {
		dialog.ShowInformation("Password Status", "No password is currently saved.", myWindow)
		return
	}

	dialog.ShowConfirm("Clear Saved Password",
		"Are you sure you want to remove the saved password from the keychain?",
		func(confirmed bool) {
			if confirmed {
				err := utils.DeleteSudoPassword()
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to clear saved password: %v", err), myWindow)
				} else {
					dialog.ShowInformation("Password Cleared", "The saved password has been removed from the keychain.", myWindow)
				}
			}
		}, myWindow)
}
