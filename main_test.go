package main

import "testing"

func TestCheckMonitors(t *testing.T) {
	cases := []struct {
		config      Config
		expectErr   bool
		name        string
		selfMonitor bool
	}{
		{
			config:    Config{},
			expectErr: false,
			name:      "Empty",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:    "Success",
						Command: CommandOrShell{Command: []string{"true"}},
					},
				},
			},
			expectErr:   false,
			name:        "Monitor success, no alerts",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertAfter: 1,
					},
				},
			},
			expectErr:   false,
			name:        "Monitor failure, no alerts",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Success",
						Command:    CommandOrShell{Command: []string{"ls"}},
						alertCount: 1,
					},
				},
			},
			expectErr:   false,
			name:        "Monitor recovery, no alerts",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"unknown"},
						AlertAfter: 1,
					},
				},
			},
			expectErr:   true,
			name:        "Monitor failure, unknown alerts",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Success",
						Command:    CommandOrShell{Command: []string{"true"}},
						AlertUp:    []string{"unknown"},
						alertCount: 1,
					},
				},
			},
			expectErr:   true,
			name:        "Monitor recovery, unknown alerts",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Success",
						Command:    CommandOrShell{Command: []string{"true"}},
						AlertUp:    []string{"unknown"},
						alertCount: 1,
					},
				},
			},
			expectErr:   false,
			name:        "Monitor recovery, unknown alerts, with Health Check",
			selfMonitor: true,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"good"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"good": {
						Command: CommandOrShell{Command: []string{"true"}},
					},
				},
			},
			expectErr:   false,
			name:        "Monitor failure, successful alert",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"bad"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"bad": {
						Name:    "bad",
						Command: CommandOrShell{Command: []string{"false"}},
					},
				},
			},
			expectErr:   true,
			name:        "Monitor failure, bad alert",
			selfMonitor: false,
		},
		{
			config: Config{
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"bad"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"bad": {
						Name:    "bad",
						Command: CommandOrShell{Command: []string{"false"}},
					},
				},
			},
			expectErr:   false,
			name:        "Monitor failure, bad alert, with Health Check",
			selfMonitor: true,
		},
	}

	for _, c := range cases {
		SelfMonitor = c.selfMonitor

		err := c.config.Init()
		if err != nil {
			t.Errorf("checkMonitors(%s): unexpected error reading config: %v", c.name, err)
		}

		err = checkMonitors(&c.config)
		if err == nil && c.expectErr {
			t.Errorf("checkMonitors(%s): Expected panic, the code did not panic", c.name)
		} else if err != nil && !c.expectErr {
			t.Errorf("checkMonitors(%s): Did not expect an error, but we got one anyway: %v", c.name, err)
		}
	}
}
