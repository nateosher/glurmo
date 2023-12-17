package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

func RunSetup(next_action []string, sim_dir string,
	script_dict map[string]string,
	slurm_dict map[string]string) error {
	if len(next_action) != 2 {
		fmt.Println("Usage: s [# simulations]")
		return nil
	}

	err := CheckIfEmpty(sim_dir)
	if err != nil {
		return err
	}

	n_sims, err := strconv.Atoi(next_action[1])
	if err != nil {
		fmt.Println("ERROR: please enter a valid number of simulations")
		fmt.Println("Usage: s [# simulations]")
		return nil
	}

	err = ScriptSetup(sim_dir, n_sims, script_dict)
	if err != nil {
		return err
	}

	err = SlurmSetup(sim_dir, n_sims, slurm_dict)
	if err != nil {
		return err
	}

	return nil
}

func CheckIfEmpty(sim_dir string) error {
	sim_dir_files, err := os.ReadDir(sim_dir)
	if err != nil {
		return err
	}

	if len(sim_dir_files) != 1 || sim_dir_files[0].Name() != ".slurminator" {
		return errorString{"directory has contents aside from `.slurminator` " +
			"directory; please remove all other contents before running setup\n"}
	}

	return nil
}

func ScriptSetup(sim_dir string, n_sims int, script_dict map[string]string) error {
	script_template, err := GetScriptTemplate(sim_dir)
	if err != nil {
		return errorString{fmt.Sprintf("ERROR: could not get script template: %s\n", err)}

	}
	script_template.Option("missingkey=error")

	os.Mkdir(sim_dir+"/scripts", 0700)

	for i := 0; i < n_sims; i++ {
		script_dict["index"] = fmt.Sprint(i)
		var final_script_raw bytes.Buffer

		err = script_template.Execute(&final_script_raw, script_dict)
		if err != nil {
			return errorString{fmt.Sprintf("ERROR: could not populate script template: %s\n", err)}
		}

		current_script_string := final_script_raw.String()

		f, err := os.Create(sim_dir + "/scripts/script_" +
			script_dict["index"] + script_dict["script_extension"])
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

func SlurmSetup(sim_dir string, n_sims int, slurm_dict map[string]string) error {
	slurm_template, err := GetSlurmTemplate(sim_dir)
	if err != nil {
		return errorString{fmt.Sprintf("ERROR: could not get slurm template: %s\n", err)}
	}

	slurm_template.Option("missingkey=error")

	os.Mkdir(sim_dir+"/slurm", 0700)

	for i := 0; i < n_sims; i++ {
		slurm_dict["index"] = fmt.Sprint(i)
		slurm_dict["pathtoscript"] = sim_dir + "/slurm/slurm_" + slurm_dict["index"]
		slurm_dict["job_id"] = slurm_dict["simulation_id"] + "|||" +
			slurm_dict["index"]

		var slurm_raw bytes.Buffer

		err = slurm_template.Execute(&slurm_raw, slurm_dict)
		if err != nil {
			return errorString{fmt.Sprintf("ERROR: could not populate slurm template: %s\n", err)}
		}

		slurm_string := slurm_raw.String()

		f, err := os.Create(sim_dir + "/slurm/slurm_" + slurm_dict["index"])
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

func CleanupOnErr() error {
	return nil
}
