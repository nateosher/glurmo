package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Get username
	uname_bytes, err := exec.Command("whoami").Output()
	if err != nil {
		// TODO: have them input it manually
		fmt.Println("ERROR: could not get username")
		return
	}
	username := strings.TrimSpace(string(uname_bytes))

	// Get flags
	setupFlag := flag.Bool("s", false, "sets up the directory before running other commands")
	runFlag := flag.Int("r", 0, "how many simulations to submit in current directory")
	cancelFlag := flag.Int("c", 0, "how many jobs to cancel in the state passed by -cs flag")
	cancelStateFlag := flag.String("cs", "", "state of jobs to cancel")
	statusFlag := flag.Bool("t", false, "reports status of directory (number running, pending, etc.)")
	flag.Parse()

	// Get simulation directory
	// TODO: if directory is not passed this just becomes name of executable - check for this
	sim_dir := flag.Arg(0)
	sim_dir, err = filepath.Abs(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not get absolute path to %s: %s\n", sim_dir, err)
	}

	// Get settings map
	settings_map, err := GetSettings(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not retrieve settings: %s\n", err)
		return
	}
	// Copying experiments
	settings_copy := DeepCopySettings(settings_map)
	settings_copy.Script["n_chains"] = "@[100, 200,300, 400, 500]"
	unpacked_list, err := UnpackList(settings_copy.Script["n_chains"])
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	// fmt.Printf("DIRECTORY: %s\n", sim_dir)
	fmt.Println("SIM DIRECTORY: ", sim_dir)
	fmt.Println("SETTINGS: ", settings_map)
	fmt.Println("SETTINGS COPY: ", settings_copy)
	fmt.Println("SPLIT LIST: ", unpacked_list)
	fmt.Printf("USERNAME: %s\n", username)
	fmt.Println("#########################################")
	fmt.Printf("VALUE OF setupFlag:  %t\n", *setupFlag)
	fmt.Printf("VALUE OF statusFlag:  %t\n", *statusFlag)
	fmt.Printf("VALUE OF runFlag:  %d\n", *runFlag)
	fmt.Printf("VALUE OF cancelFlag: %d\n", *cancelFlag)
	fmt.Printf("VALUE OF cancelStateFlag: %s\n", *cancelStateFlag)
	fmt.Println("#########################################")
	parsedInt, _ := GetFileNumber("abcd___123.pkl")
	fmt.Println("Parse int: ", parsedInt)

	nonexistantDir, err := os.Stat("simulation_study_2")
	if err != nil {
		fmt.Println("Op: ", err.(*fs.PathError).Op)
		fmt.Println("Path: ", err.(*fs.PathError).Path)
		fmt.Println("Err: ", err.(*fs.PathError).Err)
	}

	if exists, err := DirExists("simulation_study_3"); err == nil && !exists {
		fmt.Println("simulation_study_3 does not exist")
	}

	if exists, err := DirExists("simulation_study"); err == nil && exists {
		fmt.Println("simulation_study exists")
	}

	if exists, err := FileExists("simulation_study"); err == nil && !exists {
		fmt.Println("simulation_study (file) does not exist")
	}

	fmt.Println("nonexistantDir: ", nonexistantDir)
	if *setupFlag {
		err = SetupDir(sim_dir, settings_map, true)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	simSubdirs, err := GetSubdirs(sim_dir)
	if err != nil {
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("subdirs: ", simSubdirs)

	if *runFlag > 0 {
		nSubmitted, err := RunJobs(sim_dir, *runFlag)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully submitted %d jobs\n", nSubmitted)
	}
}
