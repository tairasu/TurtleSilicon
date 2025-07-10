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
	SaveSudoPassword        bool   `json:"save_sudo_password"`
	ShowTerminalNormally    bool   `json:"show_terminal_normally"`
	EnableVanillaTweaks     bool   `json:"enable_vanilla_tweaks"`
	RemapOptionAsAlt        bool   `json:"remap_option_as_alt"`
	AutoDeleteWdb           bool   `json:"auto_delete_wdb"`
	
	// Graphics settings
	ReduceTerrainDistance    bool `json:"reduce_terrain_distance"`
	SetMultisampleTo2x       bool `json:"set_multisample_to_2x"`
	SetShadowLOD0            bool `json:"set_shadow_lod_0"`
	EnableLibSiliconPatch    bool `json:"enable_lib_silicon_patch"`
	
	// Tracking whether user has manually disabled these settings
	UserDisabledShadowLOD        bool `json:"user_disabled_shadow_lod"`
	UserDisabledLibSiliconPatch  bool `json:"user_disabled_lib_silicon_patch"`
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
