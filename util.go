package main

import (
	"os/exec"
	"strings"
)

// ShellCommand takes a string and executes it as a command using `sh`
func ShellCommand(command string) *exec.Cmd {
	shellCommand := []string{"sh", "-c", strings.TrimSpace(command)}

	return exec.Command(shellCommand[0], shellCommand[1:]...)
}

// EqualSliceString checks if two string slices are equivalent
func EqualSliceString(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, val := range a {
		if val != b[i] {
			return false
		}
	}
	return true
}
