package main

import (
	"fmt"
	"strconv"
	"strings"
)

type SlurmJob struct {
	ID      string
	JobName string
	State   string
}

// Resulting [][]string has format ID, NAME, STATE
func GetCurrentSubmitted(jobID string) ([]SlurmJob, error) {
	raw, err := CommandString("squeue", "--format=\"%.18i %.50j %.8T\"", "--me")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(raw, "\n")
	slurmJobs := make([]SlurmJob, 0, len(lines))

	for i, jobString := range lines {
		if i == 0 || len(jobString) < 1 {
			continue
		}

		jobString = jobString[1 : len(jobString)-1]

		splitJob := strings.Fields(jobString)
		curJob := SlurmJob{splitJob[0], splitJob[1], splitJob[2]}
		if strings.HasPrefix(curJob.JobName, jobID) {
			slurmJobs = append(slurmJobs, curJob)
		}
	}

	return slurmJobs, nil
}

func GetJobNumber(jobName string) (int, error) {
	nameAndNumber := strings.Split(jobName, "___")
	if len(nameAndNumber) != 2 {
		return -1, errorString{fmt.Sprintf("malformed job name: %s", jobName)}
	}
	jobNum, err := strconv.Atoi(nameAndNumber[1])
	if err != nil {
		return -1, errorString{fmt.Sprintf("could not parse job number: %s", jobName)}
	}

	return jobNum, nil
}
