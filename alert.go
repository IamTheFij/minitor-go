package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"text/template"
	"time"
)

// Alert is a config driven mechanism for sending a notice
type Alert struct {
	Name                 string
	Command              CommandOrShell
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
	return !alert.Command.Empty()
}

// BuildTemplates compiles command templates for the Alert
func (alert *Alert) BuildTemplates() error {
	if LogDebug {
		log.Printf("DEBUG: Building template for alert %s", alert.Name)
	}
	if alert.commandTemplate == nil && alert.Command.Command != nil {
		alert.commandTemplate = []*template.Template{}
		for i, cmdPart := range alert.Command.Command {
			alert.commandTemplate = append(alert.commandTemplate, template.Must(
				template.New(alert.Name+string(i)).Parse(cmdPart),
			))
		}
	} else if alert.commandShellTemplate == nil && alert.Command.ShellCommand != "" {
		alert.commandShellTemplate = template.Must(
			template.New(alert.Name).Parse(alert.Command.ShellCommand),
		)
	} else {
		return fmt.Errorf("No template provided for alert %s", alert.Name)
	}

	return nil
}

// Send will send an alert notice by executing the command template
func (alert Alert) Send(notice AlertNotice) (output_str string, err error) {
	log.Printf("INFO: Sending alert %s for %s", alert.Name, notice.MonitorName)
	var cmd *exec.Cmd
	if alert.commandTemplate != nil {
		command := []string{}
		for _, cmdTmp := range alert.commandTemplate {
			var commandBuffer bytes.Buffer
			err = cmdTmp.Execute(&commandBuffer, notice)
			if err != nil {
				return
			}
			command = append(command, commandBuffer.String())
		}
		cmd = exec.Command(command[0], command[1:]...)
	} else if alert.commandShellTemplate != nil {
		var commandBuffer bytes.Buffer
		err = alert.commandShellTemplate.Execute(&commandBuffer, notice)
		if err != nil {
			return
		}
		shellCommand := commandBuffer.String()

		cmd = ShellCommand(shellCommand)
	} else {
		err = fmt.Errorf("No templates compiled for alert %v", alert.Name)
		return
	}

	// Exit if we're not ready to run the command
	if cmd == nil || err != nil {
		return
	}

	var output []byte
	output, err = cmd.CombinedOutput()
	output_str = string(output)
	if LogDebug {
		log.Printf("DEBUG: Alert output for: %s\n---\n%s\n---", alert.Name, output_str)
	}

	return output_str, err
}
