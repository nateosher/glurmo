package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Get flags
	setupFlag := flag.Bool("s", false, "sets up the directory before running other commands")
	runFlag := flag.Int("r", 0, "how many simulations to submit in current directory")
	cancelFlag := flag.Int("c", 0, "how many jobs to cancel in the state passed by -cs flag")
	cancelStateFlag := flag.String("cs", "", "state of jobs to cancel")
	// statusFlag := flag.Bool("t", false, "reports status of directory (number running, pending, etc.)")
	flag.Parse()

	// Get simulation directory
	// TODO: if directory is not passed this just becomes name of executable - check for this
	simDir := flag.Arg(0)
	simDir, err := filepath.Abs(simDir)
	if err != nil {
		fmt.Printf("ERROR: could not get absolute path to %s: %s\n", simDir, err)
	}

	// Get settings map
	settings_map, err := GetSettings(simDir)
	if err != nil {
		fmt.Printf("ERROR: could not retrieve settings: %s\n", err)
		return
	}

	// If user requested setup, run setup
	if *setupFlag {
		err = SetupDir(simDir, settings_map, true)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	// If user wants to submit jobs, submit for all sub-directories
	if *runFlag > 0 {
		nSubmitted, err := RunJobs(simDir, *runFlag)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully submitted %d jobs\n", nSubmitted)
	}

	// If user requested job cancellation, cancel jobs
	if *cancelFlag > 0 {
		cancelStatesUpper := strings.ToUpper(*cancelStateFlag)
		cancelStates := strings.Split(cancelStatesUpper, ",")
		cancelStateMap := make(map[string]bool, 2)
		for _, state := range cancelStates {
			cancelStateMap[state] = true
		}
		if !cancelStateMap["PENDING"] && !cancelStateMap["RUNNING"] {
			fmt.Printf("ERROR: cancellation states should be `RUNNING`,`PENDING`, or both (`RUNNING,PENDING`)- got: %s\n",
				*cancelStateFlag)
			os.Exit(1)
		}
		CancelJobs(simDir, *cancelFlag, cancelStateMap)
	}
}
