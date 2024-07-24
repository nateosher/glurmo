package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type SettingsMap struct {
	General map[string]string      `json:"general"`
	Script  map[string]string      `json:"script"`
	Slurm   map[string]interface{} `json:"slurm"`
}

func GetSettings(simDir string) (SettingsMap, error) {
	var settingsMap SettingsMap
	settingsDir, err := GetSettingsDir(simDir)
	if err != nil {
		return settingsMap, err
	}

	settingsFile, err := GetSettingsFile(settingsDir)
	if err != nil {
		return settingsMap, err
	}

	settingsMap, err = GetSettingsMap(settingsFile)
	if err != nil {
		return settingsMap, err
	}

	return settingsMap, nil
}

func GetSettingsDir(simDir string) (string, error) {
	simDirFiles, err := os.ReadDir(simDir)
	if err != nil {
		return "", err
	}

	hasSettingsDir := false

	for _, f := range simDirFiles {
		hasSettingsDir = hasSettingsDir || (f.Name() == ".glurmo" && f.IsDir())
		if hasSettingsDir {
			break
		}
	}

	if !hasSettingsDir {
		return "", errorString{s: "could not find settings directory (.glurmo) in directory " + simDir}
	}

	settingsDir := filepath.Join(simDir, ".glurmo")

	return settingsDir, nil
}

func GetSettingsFile(settingsDir string) (string, error) {
	settingsDirFiles, err := os.ReadDir(settingsDir)
	if err != nil {
		return "", err
	}

	hasSettingsFile := false

	for _, f := range settingsDirFiles {
		hasSettingsFile = hasSettingsFile || (f.Name() == "settings.json" && !f.IsDir())
		if hasSettingsFile {
			break
		}
	}

	if !hasSettingsFile {
		return "", errorString{s: "could not find settings file (settings.json) in directory " + settingsDir}
	}

	settingsFile := filepath.Join(settingsDir, "settings.json")

	return settingsFile, nil
}

func GetSettingsMap(settingsFile string) (SettingsMap, error) {
	var settingsMap SettingsMap
	rawBytes, err := os.ReadFile(settingsFile)
	if err != nil {
		return settingsMap, err
	}
	err = json.Unmarshal(rawBytes, &settingsMap)
	if err != nil {
		return settingsMap, err
	}

	return settingsMap, nil
}

func CheckSettingsDicts(script_dict map[string]string, slurm_dict map[string]string) error {
	// Script must have extension
	if _, extension_set := script_dict["extension"]; !extension_set {
		return errorString{s: "script settings must specify " +
			"script file extension with `extension` setting"}
	}
	// slurm must have simulation_id
	if _, id_set := slurm_dict["simulation_id"]; !id_set {
		return errorString{s: "slurm settings must specify " +
			"simulation set id with `simulation_id` setting"}
	}

	return nil
}
