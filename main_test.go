package main_test

import (
	"testing"

	m "git.iamthefij.com/iamthefij/minitor-go/v2"
)

// TestCheckConfig tests the checkConfig function
// It also tests results for potentially invalid configuration. For example, no alerts
func TestCheckMonitors(t *testing.T) {
	cases := []struct {
		config             m.Config
		expectFailureError bool
		expectRecoverError bool
		name               string
	}{
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Monitors: []*m.Monitor{
					{
						Name: "Success",
					},
				},
			},
			expectFailureError: false,
			expectRecoverError: false,
			name:               "No alerts",
		},
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Monitors: []*m.Monitor{
					{
						Name:       "Failure",
						AlertDown:  []string{"unknown"},
						AlertUp:    []string{"unknown"},
						AlertAfter: 1,
					},
				},
			},
			expectFailureError: true,
			expectRecoverError: true,
			name:               "Unknown alerts",
		},
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Monitors: []*m.Monitor{
					{
						Name:       "Failure",
						AlertDown:  []string{"good"},
						AlertUp:    []string{"good"},
						AlertAfter: 1,
					},
				},
				Alerts: []*m.Alert{{
					Name:    "good",
					Command: []string{"true"},
				}},
			},
			expectFailureError: false,
			expectRecoverError: false,
			name:               "Successful alert",
		},
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Monitors: []*m.Monitor{
					{
						Name:       "Failure",
						AlertDown:  []string{"bad"},
						AlertUp:    []string{"bad"},
						AlertAfter: 1,
					},
				},
				Alerts: []*m.Alert{{
					Name:    "bad",
					Command: []string{"false"},
				}},
			},
			expectFailureError: true,
			expectRecoverError: true,
			name:               "Failing alert",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Init()
			if err != nil {
				t.Errorf("checkMonitors(%s): unexpected error reading config: %v", c.name, err)
			}

			for _, check := range []struct {
				shellCmd  string
				name      string
				expectErr bool
			}{
				{"false", "Failure", c.expectFailureError}, {"true", "Success", c.expectRecoverError},
			} {
				// Set the shell command for this check
				c.config.Monitors[0].ShellCommand = check.shellCmd

				// Run the check
				err = m.CheckMonitors(&c.config)

				// Check the results
				if err == nil && check.expectErr {
					t.Errorf("checkMonitors(%s:%s): Expected error, the code did not error", c.name, check.name)
				} else if err != nil && !check.expectErr {
					t.Errorf("checkMonitors(%s:%s): Did not expect an error, but we got one anyway: %v", c.name, check.name, err)
				}
			}
		})
	}
}

func TestFirstRunAlerts(t *testing.T) {
	cases := []struct {
		config        m.Config
		expectErr     bool
		startupAlerts []string
		name          string
	}{
		{
			config: m.Config{
				CheckIntervalStr: "1s",
			},
			expectErr:     true,
			startupAlerts: []string{"missing"},
			name:          "Unknown",
		},
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Alerts: []*m.Alert{
					{
						Name:    "good",
						Command: []string{"true"},
					},
				},
			},
			expectErr:     false,
			startupAlerts: []string{"good"},
			name:          "Successful alert",
		},
		{
			config: m.Config{
				CheckIntervalStr: "1s",
				Alerts: []*m.Alert{
					{
						Name:    "bad",
						Command: []string{"false"},
					},
				},
			},
			expectErr:     true,
			startupAlerts: []string{"bad"},
			name:          "Failed alert",
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Init()
			if err != nil {
				t.Errorf("sendFirstRunAlerts(%s): unexpected error reading config: %v", c.name, err)
			}

			err = m.SendStartupAlerts(&c.config, c.startupAlerts)
			if err == nil && c.expectErr {
				t.Errorf("sendFirstRunAlerts(%s): Expected error, the code did not error", c.name)
			} else if err != nil && !c.expectErr {
				t.Errorf("sendFirstRunAlerts(%s): Did not expect an error, but we got one anyway: %v", c.name, err)
			}
		})
	}
}
