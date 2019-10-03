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
		{Monitor{Command: []string{"echo", "test"}}, true, "Command only"},
		{Monitor{CommandShell: "echo test"}, true, "CommandShell only"},
		{Monitor{}, false, "No commands"},
		{
			Monitor{Command: []string{"echo", "test"}, CommandShell: "echo test"},
			false,
			"Both commands",
		},
		{Monitor{AlertAfter: -1}, false, "Invalid alert threshold, -1"},
		{Monitor{AlertAfter: 0}, false, "Invalid alert threshold, 0"},
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
		{Monitor{lastCheck: timeNow, CheckInterval: 15}, false, "Just checked"},
		{Monitor{lastCheck: timeTenSecAgo, CheckInterval: 15}, false, "-10s"},
		{Monitor{lastCheck: timeTwentySecAgo, CheckInterval: 15}, true, "-20s"},
	}

	for _, c := range cases {
		actual := c.monitor.ShouldCheck()
		if actual != c.expected {
			t.Errorf("ShouldCheck(%v), expected=%t actual=%t", c.name, c.expected, actual)
		}
	}
}

// TestMonitorIsUp tests the Monitor.isUp()
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
		actual := c.monitor.isUp()
		if actual != c.expected {
			t.Errorf("isUp(%v), expected=%t actual=%t", c.name, c.expected, actual)
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

// TestMonitorFailureAlertAfter tests that alerts will not trigger until
// hitting the threshold provided by AlertAfter
func TestMonitorFailureAlertAfter(t *testing.T) {
	cases := []struct {
		monitor      Monitor
		expectNotice bool
		name         string
	}{
		{Monitor{AlertAfter: 1}, true, "Empty"}, // Defaults to true because and AlertEvery default to 0
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: 1}, true, "Alert after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: 1}, true, "Alert after 1: second failure"},
		{Monitor{failureCount: 0, AlertAfter: 20, AlertEvery: 1}, false, "Alert after 20: first failure"},
		{Monitor{failureCount: 19, AlertAfter: 20, AlertEvery: 1}, true, "Alert after 20: 20th failure"},
		{Monitor{failureCount: 20, AlertAfter: 20, AlertEvery: 1}, true, "Alert after 20: 21st failure"},
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
		{Monitor{AlertAfter: 1}, true, "Empty"}, // Defaults to true because AlertAfter and AlertEvery default to 0
		// Alert first time only, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: 0}, true, "Alert first time only after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: 0}, false, "Alert first time only after 1: second failure"},
		{Monitor{failureCount: 2, AlertAfter: 1, AlertEvery: 0}, false, "Alert first time only after 1: third failure"},
		// Alert every time, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: 1}, true, "Alert every time after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: 1}, true, "Alert every time after 1: second failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: 1}, true, "Alert every time after 1: third failure"},
		// Alert every other time, after 1
		{Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: 2}, true, "Alert every other time after 1: first failure"},
		{Monitor{failureCount: 1, AlertAfter: 1, AlertEvery: 2}, false, "Alert every other time after 1: second failure"},
		{Monitor{failureCount: 2, AlertAfter: 1, AlertEvery: 2}, true, "Alert every other time after 1: third failure"},
		{Monitor{failureCount: 3, AlertAfter: 1, AlertEvery: 2}, false, "Alert every other time after 1: fourth failure"},
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
	monitor := Monitor{failureCount: 0, AlertAfter: 1, AlertEvery: -1}
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
