package main

import (
	"log"
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
		log.Printf("Testing case %s", c.name)
		actual := c.alert.IsValid()
		if actual != c.expected {
			t.Errorf("IsValid(%v), expected=%t actual=%t", c.name, c.expected, actual)
			log.Printf("Case failed: %s", c.name)
		}
		log.Println("-----")
	}
}

func TestAlertSend(t *testing.T) {
	cases := []struct {
		alert          Alert
		notice         AlertNotice
		expectedOutput string
		expectErr      bool
		name           string
	}{
		{
			Alert{Command: CommandOrShell{Command: []string{"echo", "{{.MonitorName}}"}}},
			AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command with template",
		},
		{
			Alert{Command: CommandOrShell{ShellCommand: "echo {{.MonitorName}}"}},
			AlertNotice{MonitorName: "test"},
			"test\n",
			false,
			"Command shell with template",
		},
		{
			Alert{Command: CommandOrShell{Command: []string{"echo", "{{.Bad}}"}}},
			AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command with bad template",
		},
		{
			Alert{Command: CommandOrShell{ShellCommand: "echo {{.Bad}}"}},
			AlertNotice{MonitorName: "test"},
			"",
			true,
			"Command shell with bad template",
		},
		{
			Alert{Command: CommandOrShell{ShellCommand: "echo {alert_message}"}},
			AlertNotice{MonitorName: "test", FailureCount: 1},
			"test check has failed 1 times\n",
			false,
			"Command shell with legacy template",
		},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)
		c.alert.BuildTemplates()
		output, err := c.alert.Send(c.notice)
		hasErr := (err != nil)
		if output != c.expectedOutput {
			t.Errorf("Send(%v output), expected=%v actual=%v", c.name, c.expectedOutput, output)
			log.Printf("Case failed: %s", c.name)
		}
		if hasErr != c.expectErr {
			t.Errorf("Send(%v err), expected=%v actual=%v", c.name, "Err", err)
			log.Printf("Case failed: %s", c.name)
		}
		log.Println("-----")
	}
}

func TestAlertSendNoTemplates(t *testing.T) {
	alert := Alert{}
	notice := AlertNotice{}
	output, err := alert.Send(notice)
	if err == nil {
		t.Errorf("Send(no template), expected=%v actual=%v", "Err", output)
	}
	log.Println("-----")
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
		log.Printf("Testing case %s", c.name)
		err := c.alert.BuildTemplates()
		hasErr := (err != nil)
		if hasErr != c.expectErr {
			t.Errorf("IsValid(%v), expected=%t actual=%t", c.name, c.expectErr, err)
			log.Printf("Case failed: %s", c.name)
		}
		log.Println("-----")
	}
}
