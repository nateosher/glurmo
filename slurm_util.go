package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Simple struct representing a submitted slurm job.
// `ID“ is the job id, `JobName` is the job name, and
// `State` is the current state of the job
type SlurmJob struct {
	ID      string
	JobName string
	State   string
}

// For simulation `simName`, returns a list of `SlurmJobs`, representing
// all jobs from this simulation that have currently been submitted
func GetCurrentSubmitted(simName string) ([]SlurmJob, error) {
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
		if strings.HasPrefix(curJob.JobName, simName) {
			slurmJobs = append(slurmJobs, curJob)
		}
	}

	return slurmJobs, nil
}

// Given a job name in the format [simulation name]___[job number],
// retrieves the number of the job as an integer
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

// Gets number for files with name of format [prefix]___[#][.extension]
func GetFileNumber(fname string) (int, error) {
	suffix := strings.Split(fname, "___")[1]
	numericRunes := map[byte]bool{'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true}

	takeUntil := 0
	for takeUntil <= len(suffix) && numericRunes[suffix[takeUntil]] {
		takeUntil++
	}

	fileNumber, err := strconv.Atoi(suffix[0:takeUntil])
	if err != nil {
		return -1, errorString{fmt.Sprintf("could not parse number of file %s: %s", fname, err)}
	}

	return fileNumber, nil
}

// Given the name of a simulation, retrieves the number submitted
// (returned as an int) and a map[int]bool that indicates
// which numbers have been submitted and which have not
func GetNumberSubmitted(simName string) (int, map[int]bool, error) {
	currentSubmitted, err := GetCurrentSubmitted(simName)
	submittedMap := make(map[int]bool, len(currentSubmitted))
	if err != nil {
		return 0, nil, errorString{fmt.Sprintf("could not retrieve current slurm jobs: %s", err.Error())}
	}
	fmt.Println("submitted: ", currentSubmitted)
	for _, job := range currentSubmitted {
		if strings.HasPrefix(job.JobName, simName) {
			curJobNum, err := GetJobNumber(job.JobName)
			if err != nil {
				return -1, nil, errorString{fmt.Sprintf("could not retrieve current slurm jobs: %s", err.Error())}
			}
			submittedMap[curJobNum] = true
		}
	}
	return len(submittedMap), submittedMap, nil
}
