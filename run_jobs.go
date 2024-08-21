package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Submits `nJobsToSubmit` jobs in `simDir`. If `simDir` is a "meta"
// glurmo directory, i.e. does not actually have simulations but has
// subdirectories that do, runs `nJobsToSubmit` in all glurmo
// subdirectories.
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
				return nSubmitted, err
			}
			nSubmitted += submittedJobs
		}

		return nSubmitted, nil
	} else {
		settingsMap, err := GetSettings(simDir)
		if err != nil {
			return nSubmitted, errorString{fmt.Sprintf("failed to submit jobs in directory `%s`: %s", simDir, err)}
		}

		_, submittedMap, err := GetNumberSubmitted(settingsMap.General["id"])
		if err != nil {
			return nSubmitted, err
		}
		_, completedMap, err := GetNumberCompleted(simDir, settingsMap.Script["result_extension"])
		if err != nil {
			return nSubmitted, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
		}

		slurmDir := filepath.Join(simDir, "slurm")
		jobSlice, err := os.ReadDir(slurmDir)
		nJobs := len(jobSlice)

		if err != nil {
			return nSubmitted, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
		}

		curJobNum := 0

		for nSubmitted < nJobsToSubmit && curJobNum < nJobs {
			if !completedMap[curJobNum] && !submittedMap[curJobNum] {
				res, err := CommandString("sbatch", filepath.Join(slurmDir, "slurm_"+fmt.Sprint(curJobNum)))
				if err != nil {
					return 0, errorString{fmt.Sprintf("failed to submit jobs: %s", err)}
				}
				if strings.HasPrefix(res, "Submitted batch job") {
					nSubmitted += 1
				}
			}
			curJobNum += 1
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
