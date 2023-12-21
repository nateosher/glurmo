package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

	// Get directory to manage
	sim_dir := args[1]
	sim_dir, err := filepath.Abs(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not get absolute path to %s: %s\n", sim_dir, err)
	}

	// Get username
	uname_bytes, err := exec.Command("whoami").Output()
	if err != nil {
		// TODO: have them input it manually
		fmt.Println("ERROR: could not get username")
		return
	}

	username := strings.TrimSpace(string(uname_bytes))

	script_dict, slurm_dict, err := GetSettingsMap(sim_dir)
	if err != nil {
		fmt.Printf("ERROR: could not load settings: %s\n", err)
		return
	}

	input := bufio.NewReader(os.Stdin)
	fmt.Printf("MANAGING: %s\n", sim_dir)
	fmt.Printf("AS:       %s\n\n", username)

main_loop:
	for true {
		fmt.Println(
			"What would you like to do?\n" +
				"s -> run setup\n" +
				"t -> simulation status\n" +
				"r -> reload settings.toml\n" +
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
				err = CleanupOnErr(sim_dir)
			}
			if err != nil {
				fmt.Println("ERROR: could not remove `scripts` " +
					"and `slurm` directories or contents; please " +
					"remove manually")
			}
		case "t":
			// TODO: get status of simulations currently running
			err := CheckSimStatus(username)
			if err != nil {
				fmt.Printf("ERROR: could not check simulation status: %s\n", err)
			}
		case "r":
			script_dict, slurm_dict, err = GetSettingsMap(sim_dir)
			if err != nil {
				fmt.Printf("ERROR: could not reload settings: %s\n", err)
				return
			}
		default:
			fmt.Printf("Unrecognized command: %s\n", next_action[0])
		}
	}

}
