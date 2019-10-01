package main

import (
	"log"
	"os/exec"
	"time"
)

type Monitor struct {
	// Config values
	Name          string
	Command       []string
	CommandShell  string   `yaml:"command_shell"`
	AlertDown     []string `yaml:"alert_down"`
	AlertUp       []string `yaml:"alert_up"`
	CheckInterval float64  `yaml:"check_interval"`
	AlertAfter    int16    `yaml:"alert_after"`
	AlertEvey     int16    `yaml:"alert_every"`
	// Other values
	LastCheck  time.Time
	LastOutput string
}

func (monitor Monitor) IsValid() bool {
	atLeastOneCommand := (monitor.CommandShell != "" || monitor.Command != nil)
	atMostOneCommand := (monitor.CommandShell == "" || monitor.Command == nil)
	return atLeastOneCommand && atMostOneCommand
}

func (monitor Monitor) ShouldCheck() bool {
	if monitor.LastCheck.IsZero() {
		return true
	}

	sinceLastCheck := time.Now().Sub(monitor.LastCheck).Seconds()
	return sinceLastCheck >= monitor.CheckInterval
}

func (monitor *Monitor) Check() bool {
	// TODO: This should probably return a list of alerts since the `raise`
	// pattern doesn't carry over from Python
	var cmd *exec.Cmd

	if monitor.Command != nil {
		cmd = exec.Command(monitor.Command[0], monitor.Command[1:]...)
	} else {
		// TODO: Handle a command shell as well. This is untested

		//cmd = exec.Command("sh", "-c", "echo \"This is a test of the command system\"")
		cmd = ShellCommand(monitor.CommandShell)
	}

	output, err := cmd.CombinedOutput()
	log.Printf("Check %s\n---\n%s\n---", monitor.Name, string(output))

	is_success := (err == nil)
	if err != nil {
		log.Printf("error: %v", err)
	}

	monitor.LastCheck = time.Now()
	monitor.LastOutput = string(output)

	if is_success {
		monitor.success()
	} else {
		monitor.failure()
	}

	return is_success
}

func (monitor Monitor) success() {
	log.Printf("Great success!")
}

func (monitor *Monitor) failure() {
	log.Printf("Devastating failure. :(")
}
