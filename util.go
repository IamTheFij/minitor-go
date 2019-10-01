package main

import (
	"log"
	"os/exec"
	"strings"
)

/// escapeCommandShell accepts a command to be executed by a shell and escapes it
func escapeCommandShell(command string) string {
	// Remove extra spaces and newlines from ends
	command = strings.TrimSpace(command)
	// TODO: Not sure if this part is actually needed. Should verify
	// Escape double quotes since this will be passed in as an argument
	command = strings.Replace(command, `"`, `\"`, -1)
	return command
}

/// ShellCommand takes a string and executes it as a command using `sh`
func ShellCommand(command string) *exec.Cmd {
	shellCommand := []string{"sh", "-c", escapeCommandShell(command)}
	log.Printf("Command: %v", shellCommand)
	return exec.Command(shellCommand[0], shellCommand[1:]...)
}
