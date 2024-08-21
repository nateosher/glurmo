package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func CancelJobs(simDir string, nJobsToCancel int, stateMap map[string]bool) (int, error) {
	nCanceled := 0
	resultsExists, err := DirExists(filepath.Join(simDir, "results"))
	if err != nil {
		return 0, errorString{fmt.Sprintf("failed to cancel jobs: %s", err)}
	}

	if !resultsExists {
		// Recursively check for glurmo subdirectories
		allSubdirs, err := GetSubdirs(simDir)
		if err != nil {
			return 0, errorString{fmt.Sprintf("failed to cancel jobs: %s", err)}
		}

		for _, subdir := range allSubdirs {
			if subdir == ".glurmo" {
				continue
			}
			canceledJobs, err := CancelJobs(filepath.Join(simDir, subdir), nJobsToCancel, stateMap)
			if err != nil {
				return nCanceled, err
			}
			nCanceled += canceledJobs
		}

		return nCanceled, nil
	} else {
		settingsMap, err := GetSettings(simDir)
		if err != nil {
			return -1, errorString{fmt.Sprintf("failed to cancel jobs in directory `%s`: %s", simDir, err)}
		}

		simID := settingsMap.General["id"]

		submittedJobs, err := GetCurrentSubmitted(simID)
		if err != nil {
			return 0, errorString{fmt.Sprintf("failed to cancel jobs: %s", err)}
		}

		nJobs := len(submittedJobs)
		curJobNum := 0

		for nCanceled < nJobsToCancel && curJobNum < nJobs {
			curJob := submittedJobs[curJobNum]
			if strings.HasPrefix(curJob.JobName, simID) && stateMap[curJob.State] {
				_, err := CommandString("scancel", curJob.ID)
				if err != nil {
					return 0, errorString{fmt.Sprintf("failed to cancel jobs: %s", err)}
				}
				nCanceled += 1
			}
			curJobNum += 1
		}

		return nCanceled, nil
	}

}
