package main

import (
	"os"
	"path/filepath"
	"text/template"
)

func GetScriptTemplate(sim_dir string) (template.Template, error) {
	script_bytes, err := os.ReadFile(filepath.Join(sim_dir, ".slurminator", "script_template"))
	if err != nil {
		return template.Template{}, err
	}

	script_string := string(script_bytes)

	// TODO: check if script contains {{.results_path}} and throw
	// warning if not
	script_template := template.New("Script Template")
	script_template, err = script_template.Parse(script_string)
	if err != nil {
		return template.Template{}, err
	}

	return *script_template, nil
}

func GetSlurmTemplate(sim_dir string) (template.Template, error) {
	slurm_bytes, err := os.ReadFile(filepath.Join(sim_dir, ".slurminator", "slurm_template"))
	if err != nil {
		return template.Template{}, err
	}

	slurm_string := string(slurm_bytes)

	slurm_template := template.New("Slurm Template")
	slurm_template, err = slurm_template.Parse(slurm_string)
	if err != nil {
		return template.Template{}, err
	}

	return *slurm_template, nil
}
