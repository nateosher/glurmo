package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

// Get all subdirectories of directory `target`
func GetSubdirs(target string) ([]string, error) {
	dirs := make([]string, 0, 10)

	allContents, err := os.ReadDir(target)
	if err != nil {
		return dirs, errorString{fmt.Sprintf("could not retrieve subdirectories of %s: %s", target, err)}
	}

	for _, entry := range allContents {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// copy file `src` to `dest`
func CopyFile(src, dest string) error {
	file_contents, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dest, file_contents, 0700)
	if err != nil {
		return err
	}

	return nil
}

// Check if file exists- if so, returns true, otherwise,
// returns false
func FileExists(target string) (bool, error) {
	targetInfo, err := os.Stat(target)
	if err != nil {
		if err.(*fs.PathError).Err.Error() == "no such file or directory" {
			return false, nil
		}

		return false, err
	}

	return !targetInfo.IsDir(), nil
}

// Check if directory exists - if so, returns true,
// otherwise, returns false
func DirExists(target string) (bool, error) {
	targetInfo, err := os.Stat(target)
	if err != nil {
		if err.(*fs.PathError).Err.Error() == "no such file or directory" {
			return false, nil
		}

		return false, err
	}

	return targetInfo.IsDir(), nil
}

// Remove a file or directory if it exists. If the
// file or directory did not exist, does nothing,
// and returns nil
func RemoveIfExists(target string) error {
	targetInfo, err := os.Stat(target)
	if err != nil {
		if err.(*fs.PathError).Err.Error() == "no such file or directory" {
			return nil
		}
		return err
	}

	err = nil
	if targetInfo.IsDir() {
		var err = os.RemoveAll(target)
		if err != nil {
			return err
		}
	}

	return nil
}

// Removes everything in the `paths` slice. If the
// target does not exist, does not return an error
func RemoveAllSlice(paths []string) error {
	for _, path := range paths {
		err := os.RemoveAll(path)
		if err != nil && err.(*fs.PathError).Err.Error() != "no such file or directory" {
			return err
		}
	}
	return nil
}

// Runs a command (`command`) with arguments `args`,
// and returns resulting output as a string
func CommandString(command string, args ...string) (string, error) {
	var rawOutput bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &rawOutput
	err := cmd.Run()
	if err != nil {
		return "", errorString{fmt.Sprintf("could not run command `%s %s`: %s", command, strings.Join(args, " "), err)}
	}

	return rawOutput.String(), nil
}
