package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func RunJobs(simDir string, nJobsToSubmit int) (int, error) {
	nSubmitted := 0
	resultsExists, err := DirExists(filepath.Join(simDir, "results"))
	if err != nil {
		return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
	}

	if !resultsExists {
		// Recursively check for glurmo subdirectories
		allSubdirs, err := GetSubdirs(simDir)
		if err != nil {
			return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
		}

		for _, subdir := range allSubdirs {
			if subdir == ".glurmo" {
				continue
			}
			submittedJobs, err := RunJobs(filepath.Join(simDir, subdir), nJobsToSubmit)
			if err != nil {
				fmt.Printf("Failed to submit jobs in directory %s: %s\n", filepath.Join(simDir, subdir), err)
			}
			nSubmitted += submittedJobs
		}

		return nSubmitted, nil
	} else {
		settingsMap, err := GetSettings(simDir)
		if err != nil {
			return -1, errorString{fmt.Sprintf("failed to submit jobs in directory `%s`: %s", simDir, err)}
		}
		// submit jobs normally
		_, submittedMap, err := GetNumberSubmitted(&settingsMap)
		if err != nil {
			return 0, err
		}
		_, completedMap, err := GetNumberCompleted(simDir, settingsMap.Script["result_extension"])
		if err != nil {
			return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
		}

		slurmDir := filepath.Join(simDir, "slurm")
		jobSlice, err := os.ReadDir(slurmDir)
		nJobs := len(jobSlice)

		if err != nil {
			return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
		}

		curJob := 0

		for nSubmitted < nJobsToSubmit && curJob < nJobs {
			if !completedMap[curJob] && !submittedMap[curJob] {
				res, err := CommandString("sbatch", filepath.Join(slurmDir, "slurm_"+fmt.Sprint(curJob)))
				if err != nil {
					return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
				}
				if strings.HasPrefix(res, "Submitted batch job") {
					nSubmitted += 1
				}
			}
			curJob += 1
		}

		return nSubmitted, nil
	}

}

func GetNumberCompleted(simDir string, resultExtension string) (int, map[int]bool, error) {
	resultsDir := filepath.Join(simDir, "results")
	completedSlice, err := os.ReadDir(resultsDir)

	if err != nil {
		return -1, nil, errorString{fmt.Sprintf("could not get completed simulation count: %s", err)}
	}
	completedMap := make(map[int]bool, len(completedSlice))

	for _, file := range completedSlice {
		if strings.HasSuffix(file.Name(), resultExtension) {
			fileNumber, err := GetFileNumber(file.Name())
			if err != nil {
				return -1, nil, errorString{fmt.Sprintf("could not get completed simulation count for directory %s: %s", simDir, err)}
			}
			completedMap[fileNumber] = true
		}
	}

	return len(completedMap), completedMap, nil
}

func GetNumberSubmitted(settingsMap *SettingsMap) (int, map[int]bool, error) {
	simName := settingsMap.General["id"]
	currentSubmitted, err := GetCurrentSubmitted(settingsMap.General["id"])
	submittedMap := make(map[int]bool, len(currentSubmitted))
	if err != nil {
		return 0, nil, errorString{fmt.Sprintf("could not retrieve current slurm jobs: %s", err.Error())}
	}
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
