package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage:")
		fmt.Println("slurminator path/to/simulation/directory")

		return
	}

	sim_dir := args[1]
	sim_dir, err := filepath.Abs(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not get absolute path to %s: %s\n", sim_dir, err)
	}

	script_dict, slurm_dict, err := GetSettingsMap(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not load settings: %s\n", err)
		return
	}

	input := bufio.NewReader(os.Stdin)
	fmt.Printf("MANAGING: %s\n\n", sim_dir)

main_loop:
	for true {
		fmt.Println(
			"\nWhat would you like to do?\n" +
				"s -> run setup\n" +
				"q -> quit",
		)
		fmt.Println()

		next_action_string, err := input.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: could not read line: %s", err)
			return
		}

		next_action := strings.Split(strings.TrimSpace(next_action_string), " ")

		switch next_action[0] {
		case "q":
			break main_loop
		case "s":
			err := RunSetup(next_action, sim_dir, script_dict, slurm_dict)
			if err != nil {
				fmt.Printf("ERROR: %s", err)
				// TODO: cleanup dirs + files if it fails
				err = CleanupOnErr()
			}
			if err != nil {
				fmt.Println("ERROR: could not remove `script` " +
					"and `slurm` directories or contents; please " +
					"remove manually")
			}
		default:
			fmt.Printf("Unrecognized command: %s\n", next_action[0])
		}
	}

}
