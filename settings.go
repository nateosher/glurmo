package main

import (
	"fmt"
	"os"
	"strings"
)

func GetSettingsMap(sim_dir string) (map[string]string, map[string]string, error) {
	settings_dir, err := GetSettingsDir(sim_dir)
	if err != nil {
		return nil, nil, err
	}

	settings_file, err := GetSettingsFile(settings_dir)
	if err != nil {
		return nil, nil, err
	}

	script_dict, slurm_dict, err := ParseSettingsDict(settings_file)
	if err != nil {
		return nil, nil, err
	}

	err = CheckSettingsDicts(script_dict, slurm_dict)
	if err != nil {
		return nil, nil, err
	}

	return script_dict, slurm_dict, nil

}

func GetSettingsDir(sim_dir string) (string, error) {
	sim_dir_files, err := os.ReadDir(sim_dir)
	if err != nil {
		return "", err
	}

	has_settings_dir := false

	for _, f := range sim_dir_files {
		has_settings_dir = has_settings_dir || (f.Name() == ".slurminator" && f.IsDir())
		if has_settings_dir {
			break
		}
	}

	if !has_settings_dir {
		return "", errorString{s: "could not find settings directory (.slurminator) in directory " + sim_dir}
	}

	settings_dir := sim_dir
	if settings_dir[len(settings_dir)-1] != os.PathSeparator {
		settings_dir += string(os.PathSeparator)
	}

	settings_dir += ".slurminator"

	return settings_dir, nil
}

func GetSettingsFile(settings_dir string) (string, error) {
	settings_dir_files, err := os.ReadDir(settings_dir)
	if err != nil {
		return "", err
	}

	has_settings_file := false

	for _, f := range settings_dir_files {
		has_settings_file = has_settings_file || (f.Name() == "settings.toml" && !f.IsDir())
		if has_settings_file {
			break
		}
	}

	if !has_settings_file {
		return "", errorString{s: "could not find settings file (.slurminator) in directory " + settings_dir}
	}

	settings_file := settings_dir
	if settings_file[len(settings_file)-1] != '/' {
		settings_file += "/"
	}

	settings_file += "settings.toml"

	return settings_file, nil
}

func ParseSettingsDict(settings_file string) (map[string]string, map[string]string, error) {
	raw_bytes, err := os.ReadFile(settings_file)
	if err != nil {
		return nil, nil, err
	}

	settings_str := string(raw_bytes)
	settings_str = strings.ReplaceAll(settings_str, "\"", "")
	lines := strings.Split(strings.ReplaceAll(settings_str, "\r\n", "\n"), "\n")

	script_dict := make(map[string]string)
	slurm_dict := make(map[string]string)
	var cur_dict *map[string]string

	for i, l := range lines {
		if comment_index := strings.IndexByte(l, '#'); comment_index > -1 {
			l = l[:comment_index]
		}
		if l == "" {
			continue
		}

		split_line := strings.Split(l, "=")
		if len(split_line) == 1 && split_line[0] == "[script]" {
			cur_dict = &script_dict
			continue
		} else if len(split_line) == 1 && split_line[0] == "[slurm]" {
			cur_dict = &slurm_dict
			continue
		} else if len(split_line) == 1 {
			return nil, nil,
				errorString{s: "malformed settings file at line at line " +
					fmt.Sprint(i+1) + "\n",
				}
		}

		key := strings.TrimSpace(split_line[0])
		val := strings.TrimSpace(split_line[1])

		if _, has_key := (*cur_dict)[key]; has_key {
			return nil, nil,
				errorString{s: "settings contain repeated key: " + key}
		}

		(*cur_dict)[key] = val

	}

	return script_dict, slurm_dict, nil
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
