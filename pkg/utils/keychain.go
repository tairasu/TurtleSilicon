package utils

import (
	"fmt"
	"log"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "TurtleSilicon"
	accountName = "sudo_password"
)

// SaveSudoPassword securely stores the sudo password in the system keychain
func SaveSudoPassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	err := keyring.Set(serviceName, accountName, password)
	if err != nil {
		return fmt.Errorf("failed to save password to keychain: %v", err)
	}

	log.Println("Password saved securely to keychain")
	return nil
}

// GetSudoPassword retrieves the saved sudo password from the system keychain
func GetSudoPassword() (string, error) {
	password, err := keyring.Get(serviceName, accountName)
	if err != nil {
		// If the password doesn't exist, return empty string instead of error
		if err == keyring.ErrNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to retrieve password from keychain: %v", err)
	}

	return password, nil
}

// DeleteSudoPassword removes the saved sudo password from the system keychain
func DeleteSudoPassword() error {
	err := keyring.Delete(serviceName, accountName)
	if err != nil {
		// If the password doesn't exist, that's fine
		if err == keyring.ErrNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete password from keychain: %v", err)
	}

	log.Println("Password removed from keychain")
	return nil
}

// HasSavedSudoPassword checks if a sudo password is saved in the keychain
func HasSavedSudoPassword() bool {
	_, err := keyring.Get(serviceName, accountName)
	return err == nil
}

// GetPasswordStatusText returns a user-friendly status text for password saving
func GetPasswordStatusText() string {
	if HasSavedSudoPassword() {
		return "Password saved in keychain"
	}
	return "No password saved"
}
