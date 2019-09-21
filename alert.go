package main

import (
	"fmt"
	"os/exec"
	"text/template"
	"time"
)

type Alert struct {
	Name                 string
	Command              []string
	CommandShell         string `yaml:"command_shell"`
	commandTemplate      []template.Template
	commandShellTemplate template.Template
}

func (alert Alert) IsValid() bool {
	atLeastOneCommand := (alert.CommandShell != "" || alert.Command != nil)
	atMostOneCommand := (alert.CommandShell == "" || alert.Command == nil)
	return atLeastOneCommand && atMostOneCommand
}

func (alert *Alert) BuildTemplates() {
	if alert.commandTemplate == nil && alert.Command != nil {
		// build template
		fmt.Println("Building template for command...")
	} else if alert.commandShellTemplate == nil && alert.CommandShell != "" {
		alert.commandShellTemplate = template.Must(
			template.New(alert.Name).Parse(alert.CommandShell),
		)
	} else {
		panic("No template?")
	}
}

func (alert Alert) Send(notice AlertNotice) {
	var cmd *exec.Cmd

	if alert.commandTemplate != nil {
		// build template
		fmt.Println("Send command thing...")
	} else if alert.commandShellTemplate != nil {
		var commandBuffer bytes.Buffer
		err := alert.commandShellTemplate.Execute(&commandBuffer, notice)
		// TODO handle error
		cmd = exec.Command(commandBuffer.String())
	} else {
		panic("No template?")
	}
}

type AlertNotice struct {
	MonitorName     string
	AlertCount      int64
	FailureCount    int64
	LastCheckOutput string
	LastSuccess     time.Time
}
