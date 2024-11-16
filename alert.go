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
	Name                 string   `hcl:"name,label"`
	Command              []string `hcl:"command,optional"`
	ShellCommand         string   `hcl:"shell_command,optional"`
	commandTemplate      []*template.Template
	commandShellTemplate *template.Template
}

// AlertNotice captures the context for an alert to be sent
type AlertNotice struct {
	AlertCount      int
	FailureCount    int
	IsUp            bool
	LastSuccess     time.Time
	MonitorName     string
	LastCheckOutput string
}

// Validate checks that the Alert is properly configured and returns errors if not
func (alert Alert) Validate() error {
	hasCommand := len(alert.Command) > 0
	hasShellCommand := alert.ShellCommand != ""

	var err error

	hasAtLeastOneCommand := hasCommand || hasShellCommand
	if !hasAtLeastOneCommand {
		err = errors.Join(err, fmt.Errorf(
			"%w: alert %s has no command or shell_command configured",
			ErrInvalidAlert,
			alert.Name,
		))
	}

	hasAtMostOneCommand := !(hasCommand && hasShellCommand)
	if !hasAtMostOneCommand {
		err = errors.Join(err, fmt.Errorf(
			"%w: alert %s has both command and shell_command configured",
			ErrInvalidAlert,
			alert.Name,
		))
	}

	return err
}

// BuildTemplates compiles command templates for the Alert
func (alert *Alert) BuildTemplates() error {
	slog.Debugf("Building template for alert %s", alert.Name)

	// Time format func factory
	tff := func(formatString string) func(time.Time) string {
		return func(t time.Time) string {
			return t.Format(formatString)
		}
	}

	// Create some functions for formatting datetimes in popular formats
	timeFormatFuncs := template.FuncMap{
		"ANSIC":       tff(time.ANSIC),
		"UnixDate":    tff(time.UnixDate),
		"RubyDate":    tff(time.RubyDate),
		"RFC822Z":     tff(time.RFC822Z),
		"RFC850":      tff(time.RFC850),
		"RFC1123":     tff(time.RFC1123),
		"RFC1123Z":    tff(time.RFC1123Z),
		"RFC3339":     tff(time.RFC3339),
		"RFC3339Nano": tff(time.RFC3339Nano),
		"FormatTime": func(t time.Time, timeFormat string) string {
			return t.Format(timeFormat)
		},
		"InTZ": func(t time.Time, tzName string) (time.Time, error) {
			tz, err := time.LoadLocation(tzName)
			if err != nil {
				return t, fmt.Errorf("failed to convert time to specified tz: %w", err)
			}

			return t.In(tz), nil
		},
	}

	switch {
	case alert.commandTemplate == nil && alert.Command != nil:
		alert.commandTemplate = []*template.Template{}
		for i, cmdPart := range alert.Command {
			alert.commandTemplate = append(alert.commandTemplate, template.Must(
				template.New(alert.Name+fmt.Sprint(i)).Funcs(timeFormatFuncs).Parse(cmdPart),
			))
		}
	case alert.commandShellTemplate == nil && alert.ShellCommand != "":
		shellCmd := alert.ShellCommand

		alert.commandShellTemplate = template.Must(
			template.New(alert.Name).Funcs(timeFormatFuncs).Parse(shellCmd),
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
			"Alert %s failed to send. Returned %w: %w",
			alert.Name,
			err,
			ErrAlertFailed,
		)
	}

	return outputStr, err
}
