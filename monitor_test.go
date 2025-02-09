package main

import (
	"log"
	"testing"
	"time"
)

// TestMonitorIsValid tests the Monitor.IsValid()
func TestMonitorIsValid(t *testing.T) {
	cases := []struct {
		monitor  Monitor
		expected bool
		name     string
	}{
		{Monitor{Command: CommandOrShell{Command: []string{"echo", "test"}}, AlertDown: []string{"log"}}, true, "Command only"},
		{Monitor{Command: CommandOrShell{ShellCommand: "echo test"}, AlertDown: []string{"log"}}, true, "CommandShell only"},
		{Monitor{Command: CommandOrShell{Command: []string{"echo", "test"}}}, false, "No AlertDown"},
		{Monitor{AlertDown: []string{"log"}}, false, "No commands"},
		{Monitor{Command: CommandOrShell{Command: []string{"echo", "test"}}, AlertDown: []string{"log"}, AlertAfter: -1}, false, "Invalid alert threshold, -1"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		actual := c.monitor.IsValid()
		if actual != c.expected {
			t.Errorf("IsValid(%v), expected=%t actual=%t", c.name, c.expected, actual)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorShouldCheck tests the Monitor.ShouldCheck()
func TestMonitorShouldCheck(t *testing.T) {
	timeNow := time.Now()
	timeTenSecAgo := time.Now().Add(time.Second * -10)
	timeTwentySecAgo := time.Now().Add(time.Second * -20)

	cases := []struct {
		monitor  Monitor
		expected bool
		name     string
	}{
		{Monitor{}, true, "Empty"},
		{Monitor{lastCheck: timeNow, CheckInterval: SecondsOrDuration{time.Second * 15}}, false, "Just checked"},
		{Monitor{lastCheck: timeTenSecAgo, CheckInterval: SecondsOrDuration{time.Second * 15}}, false, "-10s"},
		{Monitor{lastCheck: timeTwentySecAgo, CheckInterval: SecondsOrDuration{time.Second * 15}}, true, "-20s"},
	}

	for _, c := range cases {
		actual := c.monitor.ShouldCheck()
		if actual != c.expected {
			t.Errorf("ShouldCheck(%v), expected=%t actual=%t", c.name, c.expected, actual)
		}
	}
}

// TestMonitorIsUp tests the Monitor.IsUp()
func TestMonitorIsUp(t *testing.T) {
	cases := []struct {
		monitor  Monitor
		expected bool
		name     string
	}{
		{Monitor{}, true, "Empty"},
		{Monitor{alertCount: 1}, false, "Has alert"},
		{Monitor{alertCount: -1}, false, "Negative alerts"},
		{Monitor{alertCount: 0}, true, "No alerts"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		actual := c.monitor.IsUp()
		if actual != c.expected {
			t.Errorf("IsUp(%v), expected=%t actual=%t", c.name, c.expected, actual)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorGetAlertNames tests that proper alert names are returned
func TestMonitorGetAlertNames(t *testing.T) {
	cases := []struct {
		monitor  Monitor
		up       bool
		expected []string
		name     string
	}{
		{Monitor{}, true, nil, "Empty up"},
		{Monitor{}, false, nil, "Empty down"},
		{Monitor{AlertUp: []string{"alert"}}, true, []string{"alert"}, "Return up"},
		{Monitor{AlertDown: []string{"alert"}}, false, []string{"alert"}, "Return down"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		actual := c.monitor.GetAlertNames(c.up)
		if !EqualSliceString(actual, c.expected) {
			t.Errorf("GetAlertNames(%v), expected=%v actual=%v", c.name, c.expected, actual)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorSuccess tests the Monitor.success()
func TestMonitorSuccess(t *testing.T) {
	cases := []struct {
		monitor      Monitor
		expectNotice bool
		name         string
	}{
		{Monitor{}, false, "Empty"},
		{Monitor{alertCount: 0}, false, "No alerts"},
		{Monitor{alertCount: 1}, true, "Has alert"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		notice := c.monitor.success()
		hasNotice := (notice != nil)

		if hasNotice != c.expectNotice {
			t.Errorf("success(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

func TestMonitorAlertCount(t *testing.T) {
	var alertEvery int16 = 1

	cases := []struct {
		checkSuccess bool
		alertCount   int16
		name         string
	}{
		{false, 1, "First failure and first alert"},
		{false, 2, "Second failure and first alert"},
		{true, 2, "Success should preserve past alert count"},
		{false, 1, "First failure and first alert after success"},
	}

	// Unlike previous tests, this one requires a static Monitor with repeated
	// calls to the failure method
	monitor := Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEvery}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		var notice *AlertNotice
		if c.checkSuccess {
			notice = monitor.success()
		} else {
			notice = monitor.failure()
		}

		if notice == nil {
			t.Errorf("failure(%v) expected notice, got nil", c.name)
		}

		if notice.AlertCount != c.alertCount {
			t.Errorf("failure(%v), expected=%v actual=%v", c.name, c.alertCount, notice.AlertCount)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorFailureAlertAfter tests that alerts will not trigger until
// hitting the threshold provided by AlertAfter
func TestMonitorFailureAlertAfter(t *testing.T) {
	var alertEvery int16 = 1

	cases := []struct {
		monitor      Monitor
		expectNotice bool
		name         string
	}{
		{Monitor{AlertAfter: 1}, true, "Empty"}, // Defaults to true because and AlertEvery default to 0
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEvery}, true, "Alert after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: &alertEvery}, true, "Alert after 1: second failure"},
		{Monitor{failureCount: 0, AlertAfter: 20, AlertEvery: &alertEvery}, false, "Alert after 20: first failure"},
		{Monitor{failureCount: 19, AlertAfter: 20, AlertEvery: &alertEvery}, true, "Alert after 20: 20th failure"},
		{Monitor{failureCount: 20, AlertAfter: 20, AlertEvery: &alertEvery}, true, "Alert after 20: 21st failure"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		notice := c.monitor.failure()
		hasNotice := (notice != nil)

		if hasNotice != c.expectNotice {
			t.Errorf("failure(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorFailureAlertEvery tests that alerts will trigger
// on the expected intervals
func TestMonitorFailureAlertEvery(t *testing.T) {
	var alertEvery0, alertEvery1, alertEvery2 int16
	alertEvery0 = 0
	alertEvery1 = 1
	alertEvery2 = 2

	cases := []struct {
		monitor      Monitor
		expectNotice bool
		name         string
	}{
		/*
			TODO: Actually found a bug in original implementation. There is an inconsistency in the way AlertAfter is treated.
			For "First alert only" (ie. AlertEvery=0), it is the number of failures to ignore before alerting, so AlertAfter=1
			will ignore the first failure and alert on the second failure
			For other intervals (ie. AlertEvery=1), it is essentially indexed on one. Essentially making AlertAfter=1 trigger
			on the first failure.

			For usabilty, this should be consistent. Consistent with what though? minitor-py? Or itself? Dun dun duuuunnnnn!
		*/
		{Monitor{AlertAfter: 1}, true, "Empty"}, // Defaults to true because AlertAfter and AlertEvery default to nil
		// Alert first time only, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEvery0}, true, "Alert first time only after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: &alertEvery0}, false, "Alert first time only after 1: second failure"},
		{Monitor{failureCount: 2, AlertAfter: 1, AlertEvery: &alertEvery0}, false, "Alert first time only after 1: third failure"},
		// Alert every time, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEvery1}, true, "Alert every time after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: &alertEvery1}, true, "Alert every time after 1: second failure"},
		{Monitor{failureCount: 2, AlertAfter: 1, AlertEvery: &alertEvery1}, true, "Alert every time after 1: third failure"},
		// Alert every other time, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEvery2}, true, "Alert every other time after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: &alertEvery2}, false, "Alert every other time after 1: second failure"},
		{Monitor{failureCount: 2, AlertAfter: 1, AlertEvery: &alertEvery2}, true, "Alert every other time after 1: third failure"},
		{Monitor{failureCount: 3, AlertAfter: 1, AlertEvery: &alertEvery2}, false, "Alert every other time after 1: fourth failure"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		notice := c.monitor.failure()
		hasNotice := (notice != nil)

		if hasNotice != c.expectNotice {
			t.Errorf("failure(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}

// TestMonitorFailureExponential tests that alerts will trigger
// with an exponential backoff after repeated failures
func TestMonitorFailureExponential(t *testing.T) {
	var alertEveryExp int16 = -1

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
	monitor := Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: &alertEveryExp}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		notice := monitor.failure()
		hasNotice := (notice != nil)

		if hasNotice != c.expectNotice {
			t.Errorf("failure(%v), expected=%t actual=%t", c.name, c.expectNotice, hasNotice)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
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
		monitor Monitor
		expect  expected
		name    string
	}{
		{
			Monitor{Command: CommandOrShell{Command: []string{"echo", "success"}}},
			expected{isSuccess: true, hasNotice: false, lastOutput: "success\n"},
			"Test successful command",
		},
		{
			Monitor{Command: CommandOrShell{ShellCommand: "echo success"}},
			expected{isSuccess: true, hasNotice: false, lastOutput: "success\n"},
			"Test successful command shell",
		},
		{
			Monitor{Command: CommandOrShell{Command: []string{"total", "failure"}}},
			expected{isSuccess: false, hasNotice: true, lastOutput: ""},
			"Test failed command",
		},
		{
			Monitor{Command: CommandOrShell{ShellCommand: "false"}},
			expected{isSuccess: false, hasNotice: true, lastOutput: ""},
			"Test failed command shell",
		},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)

		isSuccess, notice := c.monitor.Check()
		if isSuccess != c.expect.isSuccess {
			t.Errorf("Check(%v) (success), expected=%t actual=%t", c.name, c.expect.isSuccess, isSuccess)
			log.Printf("Case failed: %s", c.name)
		}

		hasNotice := (notice != nil)
		if hasNotice != c.expect.hasNotice {
			t.Errorf("Check(%v) (notice), expected=%t actual=%t", c.name, c.expect.hasNotice, hasNotice)
			log.Printf("Case failed: %s", c.name)
		}

		lastOutput := c.monitor.lastOutput
		if lastOutput != c.expect.lastOutput {
			t.Errorf("Check(%v) (output), expected=%v actual=%v", c.name, c.expect.lastOutput, lastOutput)
			log.Printf("Case failed: %s", c.name)
		}

		log.Println("-----")
	}
}
