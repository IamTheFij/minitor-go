package main

import (
	"testing"
)

func TestAlertIsValid(t *testing.T) {
	cases := []struct {
		alert    Alert
		expected bool
		name     string
	}{
		{Alert{Command: CommandOrShell{Command: []string{"echo", "test"}}}, true, "Command only"},
		{Alert{Command: CommandOrShell{ShellCommand: "echo test"}}, true, "CommandShell only"},
		{Alert{}, false, "No commands"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.alert.IsValid()
			if actual != c.expected {
				t.Errorf("expected=%t actual=%t", c.expected, actual)
			}
		})
	}
}

func TestAlertSend(t *testing.T) {
	cases := []struct {
		alert          Alert
		notice         AlertNotice
		expectedOutput string
		expectErr      bool
		name           string
		pyCompat       bool
	}{
		{
			Alert{Command: CommandOrShell{Command: []string{"echo", "{{.MonitorName}}"}}},
			AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command with template",
			false,
		},
		{
			Alert{Command: CommandOrShell{ShellCommand: "echo {{.MonitorName}}"}},
			AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command shell with template",
			false,
		},
		{
			Alert{Command: CommandOrShell{Command: []string{"echo", "{{.Bad}}"}}},
			AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command with bad template",
			false,
		},
		{
			Alert{Command: CommandOrShell{ShellCommand: "echo {{.Bad}}"}},
			AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command shell with bad template",
			false,
		},
		// Test default log alert down
		{
			*NewLogAlert(),
			AlertNotice{MonitorName: "Test", FailureCount: 1, IsUp: false},
			"Test check has failed 1 times\n",
			false,
			"Default log alert down",
			false,
		},
		// Test default log alert up
		{
			*NewLogAlert(),
			AlertNotice{MonitorName: "Test", IsUp: true},
			"Test has recovered\n",
			false,
			"Default log alert up",
			false,
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
	alert := Alert{}
	notice := AlertNotice{}

	output, err := alert.Send(notice)
	if err == nil {
		t.Errorf("Send(no template), expected=%v actual=%v", "Err", output)
	}
}

func TestAlertBuildTemplate(t *testing.T) {
	cases := []struct {
		alert     Alert
		expectErr bool
		name      string
	}{
		{Alert{Command: CommandOrShell{Command: []string{"echo", "test"}}}, false, "Command only"},
		{Alert{Command: CommandOrShell{ShellCommand: "echo test"}}, false, "CommandShell only"},
		{Alert{}, true, "No commands"},
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
