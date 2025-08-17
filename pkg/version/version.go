package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type GameVersion struct {
	ID                    string          `json:"id"`
	DisplayName           string          `json:"display_name"`
	WoWVersion            string          `json:"wow_version"`
	GamePath              string          `json:"game_path"`
	CrossOverPath         string          `json:"crossover_path"`
	ExecutableName        string          `json:"executable_name"`
	SupportsVanillaTweaks bool            `json:"supports_vanilla_tweaks"`
	SupportsDLLLoading    bool            `json:"supports_dll_loading"`
	UsesRosettaPatching   bool            `json:"uses_rosetta_patching"`
	UsesDivxDecoderPatch  bool            `json:"uses_divx_decoder_patch"`
	Settings              VersionSettings `json:"settings"`
}

type VersionSettings struct {
	EnableVanillaTweaks  bool   `json:"enable_vanilla_tweaks"`
	RemapOptionAsAlt     bool   `json:"remap_option_as_alt"`
	AutoDeleteWdb        bool   `json:"auto_delete_wdb"`
	EnableMetalHud       bool   `json:"enable_metal_hud"`
	SaveSudoPassword     bool   `json:"save_sudo_password"`
	ShowTerminalNormally bool   `json:"show_terminal_normally"`
	EnvironmentVariables string `json:"environment_variables"`

	// Graphics settings
	ReduceTerrainDistance bool `json:"reduce_terrain_distance"`
	SetMultisampleTo2x    bool `json:"set_multisample_to_2x"`
	SetShadowLOD0         bool `json:"set_shadow_lod_0"`
	EnableLibSiliconPatch bool `json:"enable_lib_silicon_patch"`

	// Tracking whether user has manually disabled these settings
	UserDisabledShadowLOD       bool `json:"user_disabled_shadow_lod"`
	UserDisabledLibSiliconPatch bool `json:"user_disabled_lib_silicon_patch"`
}

type VersionManager struct {
	CurrentVersionID string                  `json:"current_version_id"`
	Versions         map[string]*GameVersion `json:"versions"`
}

var DefaultVersions = map[string]*GameVersion{
	"turtlesilicon": {
		ID:                    "turtlesilicon",
		DisplayName:           "TurtleSilicon",
		WoWVersion:            "1.12.1",
		ExecutableName:        "WoW.exe",
		SupportsVanillaTweaks: true,
		SupportsDLLLoading:    true,
		UsesRosettaPatching:   true,
		UsesDivxDecoderPatch:  false,
		Settings:              VersionSettings{},
	},
	"epochsilicon": {
		ID:                    "epochsilicon",
		DisplayName:           "EpochSilicon (3.3.5a)",
		WoWVersion:            "3.3.5a",
		ExecutableName:        "Project-Epoch.exe",
		SupportsVanillaTweaks: false,
		SupportsDLLLoading:    true,
		UsesRosettaPatching:   false,
		UsesDivxDecoderPatch:  true,
		Settings:              VersionSettings{},
	},
	"vanillasilicon": {
		ID:                    "vanillasilicon",
		DisplayName:           "VanillaSilicon (1.12.1)",
		WoWVersion:            "1.12.1",
		ExecutableName:        "WoW.exe",
		SupportsVanillaTweaks: false,
		SupportsDLLLoading:    false,
		UsesRosettaPatching:   false,
		UsesDivxDecoderPatch:  true,
		Settings:              VersionSettings{},
	},
	"burningsilicon": {
		ID:                    "burningsilicon",
		DisplayName:           "BurningSilicon (2.4.3)",
		WoWVersion:            "2.4.3",
		ExecutableName:        "WoW.exe",
		SupportsVanillaTweaks: false,
		SupportsDLLLoading:    false,
		UsesRosettaPatching:   false,
		UsesDivxDecoderPatch:  true,
		Settings:              VersionSettings{},
	},
	"wrathsilicon": {
		ID:                    "wrathsilicon",
		DisplayName:           "WrathSilicon (3.3.5a)",
		WoWVersion:            "3.3.5a",
		ExecutableName:        "WoW.exe",
		SupportsVanillaTweaks: false,
		SupportsDLLLoading:    true,
		UsesRosettaPatching:   false,
		UsesDivxDecoderPatch:  true,
		Settings:              VersionSettings{},
	},
}

func getVersionManagerPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "TurtleSilicon", "versions.json"), nil
}

func LoadVersionManager() (*VersionManager, error) {
	path, err := getVersionManagerPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Create default version manager if file doesn't exist
		vm := &VersionManager{
			CurrentVersionID: "turtlesilicon",
			Versions:         make(map[string]*GameVersion),
		}

		// Copy default versions
		for id, version := range DefaultVersions {
			vm.Versions[id] = &GameVersion{}
			*vm.Versions[id] = *version
		}

		return vm, nil
	}

	var vm VersionManager
	if err := json.Unmarshal(data, &vm); err != nil {
		return nil, err
	}

	// Ensure all default versions exist (for updates)
	for id, defaultVersion := range DefaultVersions {
		if _, exists := vm.Versions[id]; !exists {
			vm.Versions[id] = &GameVersion{}
			*vm.Versions[id] = *defaultVersion
		}
	}

	return &vm, nil
}

func (vm *VersionManager) SaveVersionManager() error {
	path, err := getVersionManagerPath()
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(vm, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (vm *VersionManager) GetCurrentVersion() (*GameVersion, error) {
	version, exists := vm.Versions[vm.CurrentVersionID]
	if !exists {
		return nil, fmt.Errorf("current version %s not found", vm.CurrentVersionID)
	}
	return version, nil
}

func (vm *VersionManager) SetCurrentVersion(versionID string) error {
	if _, exists := vm.Versions[versionID]; !exists {
		return fmt.Errorf("version %s not found", versionID)
	}
	vm.CurrentVersionID = versionID
	return vm.SaveVersionManager()
}

func (vm *VersionManager) GetVersionList() []string {
	versions := make([]string, 0, len(vm.Versions))
	for id := range vm.Versions {
		versions = append(versions, id)
	}
	return versions
}

func (vm *VersionManager) GetVersion(versionID string) (*GameVersion, error) {
	version, exists := vm.Versions[versionID]
	if !exists {
		return nil, fmt.Errorf("version %s not found", versionID)
	}
	return version, nil
}

func (vm *VersionManager) UpdateVersion(version *GameVersion) error {
	vm.Versions[version.ID] = version
	return vm.SaveVersionManager()
}
