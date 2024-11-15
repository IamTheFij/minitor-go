package main_test

import (
	"reflect"
	"testing"
	"time"

	m "git.iamthefij.com/iamthefij/minitor-go"
)

// TestMonitorIsValid tests the Monitor.IsValid()
func TestMonitorIsValid(t *testing.T) {
	cases := []struct {
		monitor  m.Monitor
		expected bool
		name     string
	}{
		{m.Monitor{AlertAfter: 1, Command: []string{"echo", "test"}, AlertDown: []string{"log"}}, true, "Command only"},
		{m.Monitor{AlertAfter: 1, ShellCommand: "echo test", AlertDown: []string{"log"}}, true, "CommandShell only"},
		{m.Monitor{AlertAfter: 1, Command: []string{"echo", "test"}}, false, "No AlertDown"},
		{m.Monitor{AlertAfter: 1, AlertDown: []string{"log"}}, false, "No commands"},
		{m.Monitor{AlertAfter: -1, Command: []string{"echo", "test"}, AlertDown: []string{"log"}}, false, "Invalid alert threshold, -1"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.monitor.IsValid()
			if actual != c.expected {
				t.Errorf("IsValid(%v), expected=%t actual=%t", c.name, c.expected, actual)
			}
		})
	}
}

// TestMonitorShouldCheck tests the Monitor.ShouldCheck()
func TestMonitorShouldCheck(t *testing.T) {
	t.Parallel()

	// Create a monitor that should check every second and then verify it checks with some sleeps
	monitor := m.Monitor{ShellCommand: "true", CheckInterval: time.Second}

	if !monitor.ShouldCheck() {
		t.Errorf("New monitor should be ready to check")
	}

	monitor.Check()

	if monitor.ShouldCheck() {
		t.Errorf("Monitor should not be ready to check after a check")
	}

	time.Sleep(time.Second)

	if !monitor.ShouldCheck() {
		t.Errorf("Monitor should be ready to check after a second")
	}
}

// TestMonitorIsUp tests the Monitor.IsUp()
func TestMonitorIsUp(t *testing.T) {
	t.Parallel()

	// Creating a monitor that should alert after 2 failures. The monitor should be considered up until we reach two failed checks
	monitor := m.Monitor{ShellCommand: "false", AlertAfter: 2}
	if !monitor.IsUp() {
		t.Errorf("New monitor should be considered up")
	}

	monitor.Check()

	if !monitor.IsUp() {
		t.Errorf("Monitor should be considered up with one failure and no alerts")
	}

	monitor.Check()

	if monitor.IsUp() {
		t.Errorf("Monitor should be considered down with one alert")
	}
}

// TestMonitorGetAlertNames tests that proper alert names are returned
func TestMonitorGetAlertNames(t *testing.T) {
	cases := []struct {
		monitor  m.Monitor
		up       bool
		expected []string
		name     string
	}{
		{m.Monitor{}, true, nil, "Empty up"},
		{m.Monitor{}, false, nil, "Empty down"},
		{m.Monitor{AlertUp: []string{"alert"}}, true, []string{"alert"}, "Return up"},
		{m.Monitor{AlertDown: []string{"alert"}}, false, []string{"alert"}, "Return down"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := c.monitor.GetAlertNames(c.up)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("GetAlertNames(%v), expected=%v actual=%v", c.name, c.expected, actual)
			}
		})
	}
}

// TestMonitorFailureAlertAfter tests that alerts will not trigger until
// hitting the threshold provided by AlertAfter
func TestMonitorFailureAlertAfter(t *testing.T) {
	var alertEveryOne int = 1

	cases := []struct {
		monitor      m.Monitor
		numChecks    int
		expectNotice bool
		name         string
	}{
		{m.Monitor{ShellCommand: "false", AlertAfter: 1}, 1, true, "Empty After 1"}, // Defaults to true because and AlertEvery default to 0
		{m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: &alertEveryOne}, 1, true, "Alert after 1: first failure"},
		{m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: &alertEveryOne}, 2, true, "Alert after 1: second failure"},
		{m.Monitor{ShellCommand: "false", AlertAfter: 20, AlertEvery: &alertEveryOne}, 1, false, "Alert after 20: first failure"},
		{m.Monitor{ShellCommand: "false", AlertAfter: 20, AlertEvery: &alertEveryOne}, 20, true, "Alert after 20: 20th failure"},
		{m.Monitor{ShellCommand: "false", AlertAfter: 20, AlertEvery: &alertEveryOne}, 21, true, "Alert after 20: 21st failure"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			hasNotice := false

			for i := 0; i < c.numChecks; i++ {
				_, notice := c.monitor.Check()
				hasNotice = (notice != nil)
			}

			if hasNotice != c.expectNotice {
				t.Errorf("failure(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			}
		})
	}
}

// TestMonitorFailureAlertEvery tests that alerts will trigger
// on the expected intervals
func TestMonitorFailureAlertEvery(t *testing.T) {
	cases := []struct {
		monitor        m.Monitor
		expectedNotice []bool
		name           string
	}{
		{m.Monitor{ShellCommand: "false", AlertAfter: 1}, []bool{true}, "No AlertEvery set"}, // Defaults to true because AlertAfter and AlertEvery default to nil
		// Alert first time only, after 1
		{m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: Ptr(0)}, []bool{true, false, false}, "Alert first time only after 1"},
		// Alert every time, after 1
		{m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: Ptr(1)}, []bool{true, true, true}, "Alert every time after 1"},
		// Alert every other time, after 1
		{m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: Ptr(2)}, []bool{true, false, true, false}, "Alert every other time after 1"},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			for i, expectNotice := range c.expectedNotice {
				_, notice := c.monitor.Check()
				hasNotice := (notice != nil)

				if hasNotice != expectNotice {
					t.Errorf("failed %s check %d: expected=%t actual=%t", c.name, i, expectNotice, hasNotice)
				}
			}
		})
	}
}

// TestMonitorFailureExponential tests that alerts will trigger
// with an exponential backoff after repeated failures
func TestMonitorFailureExponential(t *testing.T) {
	var alertEveryExp int = -1

	cases := []struct {
		expectNotice bool
		name         string
	}{
		{true, "Alert exponential after 1: first failure"},
		{true, "Alert exponential after 1: second failure"},
		{false, "Alert exponential after 1: third failure"},
		{true, "Alert exponential after 1: fourth failure"},
		{false, "Alert exponential after 1: fifth failure"},
		{false, "Alert exponential after 1: sixth failure"},
		{false, "Alert exponential after 1: seventh failure"},
		{true, "Alert exponential after 1: eighth failure"},
	}

	// Unlike previous tests, this one requires a static Monitor with repeated
	// calls to the failure method
	monitor := m.Monitor{ShellCommand: "false", AlertAfter: 1, AlertEvery: &alertEveryExp}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// NOTE: These tests are not parallel because they rely on the state of the Monitor
			_, notice := monitor.Check()
			hasNotice := (notice != nil)

			if hasNotice != c.expectNotice {
				t.Errorf("failure(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			}
		})
	}
}

// TestMonitorCheck tests successful and failed commands and shell commands
func TestMonitorCheck(t *testing.T) {
	type expected struct {
		isSuccess  bool
		hasNotice  bool
		lastOutput string
	}

	cases := []struct {
		monitor m.Monitor
		expect  expected
		name    string
	}{
		{
			m.Monitor{AlertAfter: 1, Command: []string{"echo", "success"}},
			expected{isSuccess: true, hasNotice: false, lastOutput: "success\n"},
			"Test successful command",
		},
		{
			m.Monitor{AlertAfter: 1, ShellCommand: "echo success"},
			expected{isSuccess: true, hasNotice: false, lastOutput: "success\n"},
			"Test successful command shell",
		},
		{
			m.Monitor{AlertAfter: 1, Command: []string{"total", "failure"}},
			expected{isSuccess: false, hasNotice: true, lastOutput: ""},
			"Test failed command",
		},
		{
			m.Monitor{AlertAfter: 1, ShellCommand: "false"},
			expected{isSuccess: false, hasNotice: true, lastOutput: ""},
			"Test failed command shell",
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			isSuccess, notice := c.monitor.Check()
			if isSuccess != c.expect.isSuccess {
				t.Errorf("Check(%v) (success), expected=%t actual=%t", c.name, c.expect.isSuccess, isSuccess)
			}

			hasNotice := (notice != nil)
			if hasNotice != c.expect.hasNotice {
				t.Errorf("Check(%v) (notice), expected=%t actual=%t", c.name, c.expect.hasNotice, hasNotice)
			}

			lastOutput := c.monitor.LastOutput()
			if lastOutput != c.expect.lastOutput {
				t.Errorf("Check(%v) (output), expected=%v actual=%v", c.name, c.expect.lastOutput, lastOutput)
			}
		})
	}
}
