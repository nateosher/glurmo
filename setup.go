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
func SetupDir(simDir string, settingsMap SettingsMap, checkEmpty bool) error {
	if checkEmpty {
		isEmpty, contents, err := CheckIfEmpty(simDir)
		if err != nil {
			return errorString{fmt.Sprintf("could not complete setup: %s", err)}
		}

		if !isEmpty {
			fmt.Printf("The current directory (\"%s\") has contents other than the .glurmo directory.\n", simDir)
			fmt.Println("Specifically, it contains the following files or directories:")
			for _, entry := range contents {
				fmt.Println(entry)
			}
			fmt.Printf("Proceeding with setup may overwrite some or all of these contents. Would you like to proceed anyway? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			nextActionString, err := reader.ReadString('\n')
			if err != nil {
				return errorString{fmt.Sprintf("could not read user selection: %s\n", err)}
			}

			if nextActionString != "y\n" {
				return errorString{"setup was cancelled by user"}
			}
		}
	}

	// TODO: get "list" variables
	listVariables := GetListVars(settingsMap.Templates)
	if firstVariable, nonEmpty := FirstKey(listVariables); nonEmpty {
		variableValues, err := UnpackList(listVariables[firstVariable])
		if err != nil {
			return errorString{fmt.Sprintf("could not parse script settings: %s", err)}
		}
		if len(variableValues) == 0 {
			return errorString{fmt.Sprintf("script variable '%s' contains an empty list", firstVariable)}
		}

		dirsToMake := variableValues

		// TODO: cleanup `dirsToMake` on error
		for i := range dirsToMake {
			newSettings := DeepCopySettings(settingsMap)
			newSettings.Templates[firstVariable] = variableValues[i]
			newSettings.General["id"] += "_" + variableValues[i]

			dirsToMake[i] = filepath.Join(simDir, fmt.Sprintf("%s_%s", firstVariable, variableValues[i]))
			RemoveIfExists(dirsToMake[i])
			os.Mkdir(dirsToMake[i], 0700)
			os.Mkdir(filepath.Join(dirsToMake[i], ".glurmo"), 0700)

			CopyFile(filepath.Join(simDir, ".glurmo", "script_template"),
				filepath.Join(dirsToMake[i], ".glurmo", "script_template"))
			CopyFile(filepath.Join(simDir, ".glurmo", "slurm_template"),
				filepath.Join(dirsToMake[i], ".glurmo", "slurm_template"))

			newSettingsJSON, err := json.MarshalIndent(newSettings, "", "\t")
			if err != nil {
				return err
			}
			os.WriteFile(filepath.Join(dirsToMake[i], ".glurmo", "settings.json"),
				newSettingsJSON, 0700)

			fmt.Println("Creating ", dirsToMake[i], "...")
			err = SetupDir(dirsToMake[i], newSettings, false)
			if err != nil {
				// TODO: cleanup directories
				cleanupErr := RemoveAllSlice(dirsToMake)
				if cleanupErr != nil {
					fmt.Printf("WARNING: could not clean up directory %s: %s", simDir, cleanupErr)
				}
				return err
			}
		}
	} else {
		// TODO: cleanup dirs on error
		// No list variables, just set up as single directory
		err := ScriptSetup(simDir, settingsMap.Templates, settingsMap.General)
		if err != nil {
			return err
		}
		err = SlurmSetup(simDir, settingsMap.Templates, settingsMap.General)
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
func CheckIfEmpty(simDir string) (bool, []string, error) {
	simDirFiles, err := os.ReadDir(simDir)
	if err != nil {
		return false, nil, err
	}

	simDirFileStrings := make([]string, 0, len(simDirFiles))
	for _, dirEntry := range simDirFiles {
		if dirEntry.Name() != ".glurmo" {
			simDirFileStrings = append(simDirFileStrings, dirEntry.Name())
		}
	}

	if len(simDirFiles) != 0 {
		return false, simDirFileStrings, nil
	}
	return true, nil, nil
}

// Sets up the script subdirectory of `simDir` directory
func ScriptSetup(simDir string, scriptDict map[string]string, generalSettings map[string]string) error {
	scriptTemplate, err := GetScriptTemplate(simDir)
	if err != nil {
		return errorString{fmt.Sprintf("could not get script template: %s\n", err)}

	}
	scriptTemplate.Option("missingkey=error")

	os.Mkdir(filepath.Join(simDir, "scripts"), 0700)
	os.Mkdir(filepath.Join(simDir, "results"), 0700)

	nSimsString, hasKey := generalSettings["n_sims"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"n_sims\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", simDir)}
	}
	nSims, err := strconv.Atoi(nSimsString)
	if err != nil {
		return errorString{fmt.Sprintf("could not set up script files: %s", err)}
	}

	for i := 0; i < nSims; i++ {
		scriptDict["index"] = fmt.Sprint(i)
		scriptDict["results_path"] = filepath.Join(simDir, "results", "results___"+scriptDict["index"])
		var finalScriptRaw bytes.Buffer

		err = scriptTemplate.Execute(&finalScriptRaw, scriptDict)
		if err != nil {
			return errorString{fmt.Sprintf("could not populate script template: %s\n", err)}
		}

		currentScriptString := finalScriptRaw.String()

		f, err := os.Create(filepath.Join(simDir, "scripts", "script_"+
			scriptDict["index"]+scriptDict["script_extension"]))
		if err != nil {
			return err
		}

		_, err = f.WriteString(currentScriptString)
		if err != nil {
			return err
		}

	}

	return nil
}

// Sets up slurm subdirectory of `simDir`
func SlurmSetup(simDir string, slurmDict map[string]string, generalSettings map[string]string) error {
	simID, hasKey := generalSettings["id"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"id\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", simDir)}
	}

	slurmDict["id"] = simID

	slurmTemplate, err := GetSlurmTemplate(simDir)
	if err != nil {
		return errorString{fmt.Sprintf("could not get slurm template: %s\n", err)}
	}

	slurmTemplate.Option("missingkey=error")

	os.Mkdir(filepath.Join(simDir, "slurm"), 0700)
	os.Mkdir(filepath.Join(simDir, "slurm_out"), 0700)
	os.Mkdir(filepath.Join(simDir, "slurm_errors"), 0700)

	nSimsString, hasKey := generalSettings["n_sims"]
	if !hasKey {
		return errorString{fmt.Sprintf("\"n_sims\" must be specified in \"general\" section of \".glurmo/settings.json\" (%s)", simDir)}
	}
	nSims, err := strconv.Atoi(nSimsString)
	if err != nil {
		return errorString{fmt.Sprintf("could not set up slurm files: %s", err)}
	}

	for i := 0; i < nSims; i++ {
		slurmDict["index"] = fmt.Sprint(i)
		slurmDict["path_to_script"] = filepath.Join(simDir, "slurm", "slurm_"+slurmDict["index"])
		slurmDict["job_id"] = slurmDict["id"] + "___" + slurmDict["index"]
		slurmDict["output_path"] = filepath.Join(simDir, "slurm_out", "output___"+slurmDict["index"])
		slurmDict["error_path"] = filepath.Join(simDir, "slurm_errors", "error___"+slurmDict["index"])

		var slurmRaw bytes.Buffer

		err = slurmTemplate.Execute(&slurmRaw, slurmDict)
		if err != nil {
			return errorString{fmt.Sprintf("could not populate slurm template: %s\n", err)}
		}

		slurmString := slurmRaw.String()

		f, err := os.Create(slurmDict["path_to_script"])
		if err != nil {
			return err
		}

		_, err = f.WriteString(slurmString)
		if err != nil {
			return err
		}

	}

	return nil
}

// Cleans up glurmo directory in case of an error
// TODO: clean up other directories as well
func CleanupOnErr(simDir string) error {
	err := os.RemoveAll(filepath.Join(simDir, "scripts"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(simDir, "slurm"))
	if err != nil {
		return err
	}
	return nil
}

// Determines which variables in the simulation settings are
// list variables, i.e. will create their own glurmo subdirectories
// recursively.
func GetListVars(settings map[string]string) map[string]string {
	listVars := make(map[string]string)
	for k, v := range settings {
		if strings.HasPrefix(v, "@") {
			listVars[k] = v
		}
	}
	return listVars
}

// Given a variable list of the form `@[v_1, ..., v_n]`,
// returns a slice [v_1, v_n]
func UnpackList(s string) ([]string, error) {
	if s[0:2] != "@[" || s[len(s)-1] != ']' {
		return nil, errorString{"malformed list - lists must be enclosed by @[ ... ]"}
	}
	re := regexp.MustCompile(", *")
	splitList := re.Split(s[2:len(s)-1], -1)
	return splitList, nil
}
