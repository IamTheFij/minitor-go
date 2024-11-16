package main_test

import (
	"errors"
	"testing"

	m "git.iamthefij.com/iamthefij/minitor-go"
)

func TestAlertValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		alert    m.Alert
		expected error
		name     string
	}{
		{m.Alert{Command: []string{"echo", "test"}}, nil, "Command only"},
		{m.Alert{ShellCommand: "echo test"}, nil, "CommandShell only"},
		{m.Alert{Command: []string{"echo", "test"}, ShellCommand: "echo test"}, m.ErrInvalidAlert, "Both commands"},
		{m.Alert{}, m.ErrInvalidAlert, "No commands"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.alert.Validate()
			hasErr := (actual != nil)
			expectErr := (c.expected != nil)

			if hasErr != expectErr || !errors.Is(actual, c.expected) {
				t.Errorf("expected=%t actual=%t", c.expected, actual)
			}
		})
	}
}

func TestAlertSend(t *testing.T) {
	cases := []struct {
		alert          m.Alert
		notice         m.AlertNotice
		expectedOutput string
		expectErr      bool
		name           string
	}{
		{
			m.Alert{Command: []string{"echo", "{{.MonitorName}}"}},
			m.AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command with template",
		},
		{
			m.Alert{ShellCommand: "echo {{.MonitorName}}"},
			m.AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command shell with template",
		},
		{
			m.Alert{Command: []string{"echo", "{{.Bad}}"}},
			m.AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command with bad template",
		},
		{
			m.Alert{ShellCommand: "echo {{.Bad}}"},
			m.AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command shell with bad template",
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.alert.BuildTemplates()
			if err != nil {
				t.Errorf("Send(%v output), error building templates: %v", c.name, err)
			}

			output, err := c.alert.Send(c.notice)
			hasErr := (err != nil)

			if output != c.expectedOutput {
				t.Errorf("Send(%v output), expected=%v actual=%v", c.name, c.expectedOutput, output)
			}

			if hasErr != c.expectErr {
				t.Errorf("Send(%v err), expected=%v actual=%v", c.name, "Err", err)
			}
		})
	}
}

func TestAlertSendNoTemplates(t *testing.T) {
	alert := m.Alert{}
	notice := m.AlertNotice{}

	output, err := alert.Send(notice)
	if err == nil {
		t.Errorf("Send(no template), expected=%v actual=%v", "Err", output)
	}
}

func TestAlertBuildTemplate(t *testing.T) {
	cases := []struct {
		alert     m.Alert
		expectErr bool
		name      string
	}{
		{m.Alert{Command: []string{"echo", "test"}}, false, "Command only"},
		{m.Alert{ShellCommand: "echo test"}, false, "CommandShell only"},
		{m.Alert{}, true, "No commands"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.alert.BuildTemplates()
			hasErr := (err != nil)

			if hasErr != c.expectErr {
				t.Errorf("IsValid(%v), expected=%t actual=%t", c.name, c.expectErr, err)
			}

		})
	}
}
