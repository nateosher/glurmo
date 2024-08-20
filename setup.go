package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Given the path to a glurmo directory, a SettingsMap, and an indication of whether
// or not to check if the directory is empty, sets up the infrastructure of the
// glurmo directory
func SetupDir(sim_dir string, settings_map SettingsMap, checkEmpty bool) error {
	if checkEmpty {
		isEmpty, err := CheckIfEmpty(sim_dir)
		if err != nil {
			return errorString{fmt.Sprintf("could not complete setup: %s", err)}
		}

		if !isEmpty {
			fmt.Printf(`The current directory ("%s") has contents other than the .glurmo directory.
Proceeding with setup may overwrite some or all of these contents. Would you like to proceed anyway? (y/n): `, sim_dir)
			reader := bufio.NewReader(os.Stdin)
			next_action_string, err := reader.ReadString('\n')
			if err != nil {
				return errorString{fmt.Sprintf("could not read user selection: %s\n", err)}
			}

			if next_action_string != "y\n" {
				return errorString{fmt.Sprintf("setup was cancelled by user")}
			}
		}
	}

	// TODO: get "list" variables
	list_variables := GetListVars(settings_map.Script)
	if first_variable, non_empty := FirstKey(list_variables); non_empty {
		variableValues, err := UnpackList(list_variables[first_variable])
		if err != nil {
			return errorString{fmt.Sprintf("could not parse script settings: %s", err)}
		}
		if len(variableValues) == 0 {
			return errorString{fmt.Sprintf("script variable '%s' contains an empty list", first_variable)}
		}

		dirsToMake := variableValues

		// TODO: cleanup `dirsToMake` on error
		for i := range dirsToMake {
			new_settings := DeepCopySettings(settings_map)
			new_settings.Script[first_variable] = variableValues[i]
			new_settings.General["id"] += "_" + variableValues[i]

			dirsToMake[i] = filepath.Join(sim_dir, fmt.Sprintf("%s_%s", first_variable, variableValues[i]))
			RemoveIfExists(dirsToMake[i])
			os.Mkdir(dirsToMake[i], 0700)
			os.Mkdir(filepath.Join(dirsToMake[i], ".glurmo"), 0700)

			CopyFile(filepath.Join(sim_dir, ".glurmo", "script_template"),
				filepath.Join(dirsToMake[i], ".glurmo", "script_template"))
			CopyFile(filepath.Join(sim_dir, ".glurmo", "slurm_template"),
				filepath.Join(dirsToMake[i], ".glurmo", "slurm_template"))

			new_settings_json, err := json.MarshalIndent(new_settings, "", "\t")
			if err != nil {
				return err
			}
			os.WriteFile(filepath.Join(dirsToMake[i], ".glurmo", "settings.json"),
				new_settings_json, 0700)

			fmt.Println(dirsToMake[i])
			err = SetupDir(dirsToMake[i], new_settings, false)
			if err != nil {
				// TODO: cleanup directories
				cleanupErr := RemoveAllSlice(dirsToMake)
				if cleanupErr != nil {
					fmt.Printf("WARNING: could not clean up directory %s: %s", sim_dir, cleanupErr)
				}
				return err
			}
		}
	} else {
		// TODO: cleanup dirs on error
		// No list variables, just set up as single directory
		err := ScriptSetup(sim_dir, settings_map.Script, settings_map.General)
		if err != nil {
			return err
		}
		err = SlurmSetup(sim_dir, settings_map.Slurm, settings_map.General)
		if err != nil {
			return err
		}
	}

	// TODO: get "dict" variables
	// TODO: make sub-directories accordingly

	return nil
}

// Checks if the given directory is empty (aside from a
// .glurmo subdirectory)
func CheckIfEmpty(sim_dir string) (bool, error) {
	sim_dir_files, err := os.ReadDir(sim_dir)
	if err != nil {
		return false, err
	}
	fmt.Println(sim_dir_files)

	if len(sim_dir_files) != 1 || sim_dir_files[0].Name() != ".glurmo" {
		return false, nil
	}
	return true, nil
}

// Sets up the script subdirectory of `sim_dir` directory
func ScriptSetup(sim_dir string, script_dict map[string]string, generalSettings map[string]string) error {
	script_template, err := GetScriptTemplate(sim_dir)
	if err != nil {
		return errorString{fmt.Sprintf("could not get script template: %s\n", err)}

	}
	script_template.Option("missingkey=error")

	os.Mkdir(filepath.Join(sim_dir, "scripts"), 0700)
	os.Mkdir(filepath.Join(sim_dir, "results"), 0700)

	nSimsString, hasKey := generalSettings["n_sims"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"n_sims\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", sim_dir)}
	}
	n_sims, err := strconv.Atoi(nSimsString)
	if err != nil {
		return errorString{fmt.Sprintf("could not set up script files: %s", err)}
	}

	for i := 0; i < n_sims; i++ {
		script_dict["index"] = fmt.Sprint(i)
		script_dict["results_path"] = filepath.Join(sim_dir, "results", "results___"+script_dict["index"])
		var final_script_raw bytes.Buffer

		err = script_template.Execute(&final_script_raw, script_dict)
		if err != nil {
			return errorString{fmt.Sprintf("could not populate script template: %s\n", err)}
		}

		current_script_string := final_script_raw.String()

		f, err := os.Create(filepath.Join(sim_dir, "scripts", "script_"+
			script_dict["index"]+script_dict["script_extension"]))
		if err != nil {
			return err
		}

		_, err = f.WriteString(current_script_string)
		if err != nil {
			return err
		}

	}

	return nil
}

// Sets up slurm subdirectory of `sim_dir`
func SlurmSetup(sim_dir string, slurm_dict map[string]interface{}, generalSettings map[string]string) error {
	slurmStringDict, err := InterfaceToStringMap(slurm_dict)
	if err != nil {
		return errorString{fmt.Sprintf("could not set up \"slurm\" directory: %s", err)}
	}
	simID, hasKey := generalSettings["id"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"id\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", sim_dir)}
	}

	slurmStringDict["id"] = simID

	slurm_template, err := GetSlurmTemplate(sim_dir)
	if err != nil {
		return errorString{fmt.Sprintf("could not get slurm template: %s\n", err)}
	}

	slurm_template.Option("missingkey=error")

	os.Mkdir(filepath.Join(sim_dir, "slurm"), 0700)
	os.Mkdir(filepath.Join(sim_dir, "slurm_out"), 0700)
	os.Mkdir(filepath.Join(sim_dir, "slurm_errors"), 0700)

	nSimsString, hasKey := generalSettings["n_sims"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"n_sims\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", sim_dir)}
	}
	n_sims, err := strconv.Atoi(nSimsString)
	if err != nil {
		return errorString{fmt.Sprintf("could not set up slurm files: %s", err)}
	}

	for i := 0; i < n_sims; i++ {
		slurmStringDict["index"] = fmt.Sprint(i)
		slurmStringDict["path_to_script"] = filepath.Join(sim_dir, "slurm", "slurm_"+slurmStringDict["index"])
		slurmStringDict["job_id"] = slurmStringDict["id"] + "___" + slurmStringDict["index"]
		slurmStringDict["output_path"] = filepath.Join(sim_dir, "slurm_out", "output___"+slurmStringDict["index"])
		slurmStringDict["error_path"] = filepath.Join(sim_dir, "slurm_errors", "error___"+slurmStringDict["index"])

		var slurm_raw bytes.Buffer

		err = slurm_template.Execute(&slurm_raw, slurmStringDict)
		if err != nil {
			return errorString{fmt.Sprintf("could not populate slurm template: %s\n", err)}
		}

		slurm_string := slurm_raw.String()

		f, err := os.Create(slurmStringDict["path_to_script"])
		if err != nil {
			return err
		}

		_, err = f.WriteString(slurm_string)
		if err != nil {
			return err
		}

	}

	return nil
}

// Cleans up glurmo directory in case of an error
// TODO: clean up other directories as well
func CleanupOnErr(sim_dir string) error {
	err := os.RemoveAll(filepath.Join(sim_dir, "scripts"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(sim_dir, "slurm"))
	if err != nil {
		return err
	}
	return nil
}

// Determines which variables in the simulation settings are
// list variables, i.e. will create their own glurmo subdirectories
// recursively.
func GetListVars(settings map[string]string) map[string]string {
	list_vars := make(map[string]string)
	for k, v := range settings {
		if strings.HasPrefix(v, "@") {
			list_vars[k] = v
		}
	}
	return list_vars
}

// Given a variable list of the form `@[v_1, ..., v_n]`,
// returns a slice [v_1, v_n]
func UnpackList(s string) ([]string, error) {
	if s[0:2] != "@[" || s[len(s)-1] != ']' {
		return nil, errorString{"malformed list - lists must be enclosed by @[ ... ]"}
	}
	re := regexp.MustCompile(", *")
	split_list := re.Split(s[2:len(s)-1], -1)
	return split_list, nil
}
