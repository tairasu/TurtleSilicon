package service

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
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

	// Show password dialog
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter your sudo password")
	passwordEntry.Resize(fyne.NewSize(300, passwordEntry.MinSize().Height))

	// Create a container with proper sizing
	passwordForm := widget.NewForm(widget.NewFormItem("Password:", passwordEntry))
	passwordContainer := container.NewVBox(
		widget.NewLabel("Enter your sudo password to start the RosettaX87 service:"),
		passwordForm,
	)
	passwordContainer.Resize(fyne.NewSize(400, 100))

	// Create the dialog variable so we can reference it in the callback
	var passwordDialog dialog.Dialog

	// Define the confirm logic as a function so it can be reused
	confirmFunc := func() {
		password := passwordEntry.Text
		if password == "" {
			dialog.ShowError(fmt.Errorf("password cannot be empty"), myWindow)
			return
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
	// Use sudo with the password
	cmd := exec.Command("sudo", "-S", executable)
	cmd.Dir = workingDir

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
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return fmt.Errorf("process exited prematurely with code: %d", cmd.ProcessState.ExitCode())
	}

	log.Printf("RosettaX87 service started with PID: %d", servicePID)
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
	if !ServiceRunning {
		return false
	}

	// Double-check by verifying the process is still alive
	if serviceCmd != nil && serviceCmd.Process != nil {
		// Check if process is still running
		err := serviceCmd.Process.Signal(syscall.Signal(0))
		if err != nil {
			// Process is not running
			ServiceRunning = false
			serviceCmd = nil
			servicePID = 0
			return false
		}
	}

	return ServiceRunning
}

// CleanupService ensures the service is stopped when the application exits
func CleanupService() {
	if ServiceRunning && serviceCmd != nil && serviceCmd.Process != nil {
		log.Println("Cleaning up RosettaX87 service on application exit...")
		serviceCmd.Process.Kill()
		ServiceRunning = false
	}
}
