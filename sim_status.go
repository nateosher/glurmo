package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func CheckSimStatus(username string) error {
	cur_running, err := GetAllCurrentJobs()
	if err != nil {
		return err
	}

	_, job_states := GetJobNamesAndStates(&cur_running, username)

	state_counts := make(map[string]int)

	state_counts["RUNNING"] = 0
	state_counts["PENDING"] = 0

	for _, state := range job_states {
		state_counts[state] = state_counts[state] + 1
	}

	fmt.Println()
	for state, count := range state_counts {
		plural := "s"
		if count == 1 {
			plural = ""
		}
		fmt.Printf("%d job"+plural+" in state %s\n", count, state)
	}
	fmt.Println()

	return nil
}

func GetAllCurrentJobs() ([][]string, error) {
	cur_running_bytes, err := exec.Command("squeue").Output()
	if err != nil {
		return nil, err
	}

	cur_running_string := strings.TrimSpace(string(cur_running_bytes))

	cur_running := make([][]string, 0, strings.Count(cur_running_string, "\n")+1)

	scanner := bufio.NewScanner(strings.NewReader(cur_running_string))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		// split line
		cur_running = append(cur_running, strings.Fields(scanner.Text()))
	}

	return cur_running, nil
}

func GetJobNamesAndStates(all_jobs *[][]string, username string) (job_names []string, job_states []string) {
	username_column := IndexOf((*all_jobs)[0], "USER")
	job_name_column := IndexOf((*all_jobs)[0], "NAME")
	job_state_column := IndexOf((*all_jobs)[0], "STATE")

	job_names = make([]string, 0, len(*all_jobs))
	job_states = make([]string, 0, len(*all_jobs))

	for _, row := range *all_jobs {
		if row[username_column] == username {
			job_names = append(job_names, row[job_name_column])
			job_states = append(job_states, row[job_state_column])
		}
	}
	return job_names, job_states
}

func IndexOf(list []string, target string) int {
	for i, s := range list {
		if s == target {
			return i
		}
	}

	return -1
}
