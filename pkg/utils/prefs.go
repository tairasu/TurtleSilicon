package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UserPrefs struct {
	SuppressedUpdateVersion string `json:"suppressed_update_version"`
	TurtleWoWPath           string `json:"turtlewow_path"`
	CrossOverPath           string `json:"crossover_path"`
	EnvironmentVariables    string `json:"environment_variables"`
}

func getPrefsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "TurtleSilicon", "prefs.json"), nil
}

func LoadPrefs() (*UserPrefs, error) {
	path, err := getPrefsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &UserPrefs{}, nil // default prefs if not found
	}
	var prefs UserPrefs
	if err := json.Unmarshal(data, &prefs); err != nil {
		return &UserPrefs{}, nil
	}
	return &prefs, nil
}

func SavePrefs(prefs *UserPrefs) error {
	path, err := getPrefsPath()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
