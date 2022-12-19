package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"text/template"
	"time"

	"git.iamthefij.com/iamthefij/slog"
)

var (
	errNoTemplate = errors.New("no template")

	// ErrAlertFailed indicates that an alert failed to send
	ErrAlertFailed = errors.New("alert failed")
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
	AlertCount      int16
	FailureCount    int16
	IsUp            bool
	LastSuccess     time.Time
	MonitorName     string
	LastCheckOutput string
}

// IsValid returns a boolean indicating if the Alert has been correctly
// configured
func (alert Alert) IsValid() bool {
	return !alert.Command.Empty()
}

// BuildTemplates compiles command templates for the Alert
func (alert *Alert) BuildTemplates() error {
	slog.Debugf("Building template for alert %s", alert.Name)

	switch {
	case alert.commandTemplate == nil && alert.Command.Command != nil:
		alert.commandTemplate = []*template.Template{}
		for i, cmdPart := range alert.Command.Command {
			alert.commandTemplate = append(alert.commandTemplate, template.Must(
				template.New(alert.Name+fmt.Sprint(i)).Parse(cmdPart),
			))
		}
	case alert.commandShellTemplate == nil && alert.Command.ShellCommand != "":
		shellCmd := alert.Command.ShellCommand

		alert.commandShellTemplate = template.Must(
			template.New(alert.Name).Parse(shellCmd),
		)
	default:
		return fmt.Errorf("No template provided for alert %s: %w", alert.Name, errNoTemplate)
	}

	return nil
}

// Send will send an alert notice by executing the command template
func (alert Alert) Send(notice AlertNotice) (outputStr string, err error) {
	slog.Infof("Sending alert %s for %s", alert.Name, notice.MonitorName)

	var cmd *exec.Cmd

	switch {
	case alert.commandTemplate != nil:
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
	case alert.commandShellTemplate != nil:
		var commandBuffer bytes.Buffer

		err = alert.commandShellTemplate.Execute(&commandBuffer, notice)
		if err != nil {
			return
		}

		shellCommand := commandBuffer.String()

		cmd = ShellCommand(shellCommand)
	default:
		err = fmt.Errorf("No templates compiled for alert %s: %w", alert.Name, errNoTemplate)

		return
	}

	// Exit if we're not ready to run the command
	if cmd == nil || err != nil {
		return
	}

	var output []byte
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	slog.Debugf("Alert output for: %s\n---\n%s\n---", alert.Name, outputStr)

	if err != nil {
		err = fmt.Errorf(
			"Alert '%s' failed to send. Returned %v: %w",
			alert.Name,
			err,
			ErrAlertFailed,
		)
	}

	return outputStr, err
}

// NewLogAlert creates an alert that does basic logging using echo
func NewLogAlert() *Alert {
	return &Alert{
		Name: "log",
		Command: CommandOrShell{
			Command: []string{
				"echo",
				"{{.MonitorName}} {{if .IsUp}}has recovered{{else}}check has failed {{.FailureCount}} times{{end}}",
			},
		},
	}
}
