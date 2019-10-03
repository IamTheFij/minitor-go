package main

import (
	"bytes"
	"log"
	"os/exec"
	"text/template"
	"time"
)

// Alert is a config driven mechanism for sending a notice
type Alert struct {
	Name                 string
	Command              []string
	CommandShell         string `yaml:"command_shell"`
	commandTemplate      []*template.Template
	commandShellTemplate *template.Template
}

// AlertNotice captures the context for an alert to be sent
type AlertNotice struct {
	MonitorName     string
	AlertCount      int16
	FailureCount    int16
	LastCheckOutput string
	LastSuccess     time.Time
	IsUp            bool
}

// IsValid returns a boolean indicating if the Alert has been correctly
// configured
func (alert Alert) IsValid() bool {
	atLeastOneCommand := (alert.CommandShell != "" || alert.Command != nil)
	atMostOneCommand := (alert.CommandShell == "" || alert.Command == nil)
	return atLeastOneCommand && atMostOneCommand
}

// BuildTemplates compiles command templates for the Alert
func (alert *Alert) BuildTemplates() {
	if alert.commandTemplate == nil && alert.Command != nil {
		// build template
		log.Println("Building template for command...")
		alert.commandTemplate = []*template.Template{}
		for i, cmdPart := range alert.Command {
			alert.commandTemplate = append(alert.commandTemplate, template.Must(
				template.New(alert.Name+string(i)).Parse(cmdPart),
			))
		}
		log.Printf("Template built: %v", alert.commandTemplate)
	} else if alert.commandShellTemplate == nil && alert.CommandShell != "" {
		log.Println("Building template for shell command...")
		alert.commandShellTemplate = template.Must(
			template.New(alert.Name).Parse(alert.CommandShell),
		)
		log.Printf("Template built: %v", alert.commandShellTemplate)
	} else {
		log.Fatalf("No template provided for alert %s", alert.Name)
	}
}

// Send will send an alert notice by executing the command template
func (alert Alert) Send(notice AlertNotice) {
	var cmd *exec.Cmd

	if alert.commandTemplate != nil {
		// build template
		log.Println("Send command thing...")
		command := []string{}
		for _, cmdTmp := range alert.commandTemplate {
			var commandBuffer bytes.Buffer
			err := cmdTmp.Execute(&commandBuffer, notice)
			if err != nil {
				panic(err)
			}
			command = append(command, commandBuffer.String())
		}
		cmd = exec.Command(command[0], command[1:]...)
	} else if alert.commandShellTemplate != nil {
		var commandBuffer bytes.Buffer
		err := alert.commandShellTemplate.Execute(&commandBuffer, notice)
		if err != nil {
			panic(err)
		}
		shellCommand := commandBuffer.String()

		log.Printf("About to run alert command: %s", shellCommand)
		cmd = ShellCommand(shellCommand)
	} else {
		panic("No template compiled?")
	}

	output, err := cmd.CombinedOutput()
	log.Printf("Check %s\n---\n%s\n---", alert.Name, string(output))
	if err != nil {
		panic(err)
	}
}
