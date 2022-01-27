package main

import "testing"

func Ptr[T any](v T) *T {
	return &v
}

func TestCheckMonitors(t *testing.T) {
	cases := []struct {
		config    Config
		expectErr bool
		name      string
	}{
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:    "Success",
						Command: []string{"true"},
					},
				},
			},
			expectErr: false,
			name:      "Monitor success, no alerts",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    []string{"false"},
						AlertAfter: Ptr(1),
					},
				},
			},
			expectErr: true,
			name:      "Monitor failure, no alerts",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Success",
						Command:    []string{"ls"},
						alertCount: 1,
					},
				},
			},
			expectErr: false,
			name:      "Monitor recovery, no alerts",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    []string{"false"},
						AlertDown:  []string{"unknown"},
						AlertAfter: Ptr(1),
					},
				},
			},
			expectErr: true,
			name:      "Monitor failure, unknown alerts",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Success",
						Command:    []string{"true"},
						AlertUp:    []string{"unknown"},
						alertCount: 1,
					},
				},
			},
			expectErr: true,
			name:      "Monitor recovery, unknown alerts",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    []string{"false"},
						AlertDown:  []string{"good"},
						AlertAfter: Ptr(1),
					},
				},
				Alerts: []*Alert{{
					Name:    "good",
					Command: []string{"true"},
				}},
			},
			expectErr: false,
			name:      "Monitor failure, successful alert",
		},
		{
			config: Config{
				CheckIntervalStr: "1s",
				Monitors: []*Monitor{
					{
						Name:       "Failure",
						Command:    []string{"false"},
						AlertDown:  []string{"bad"},
						AlertAfter: Ptr(1),
					},
				},
				Alerts: []*Alert{{
					Name:    "bad",
					Command: []string{"false"},
				}},
			},
			expectErr: true,
			name:      "Monitor failure, bad alert",
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

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
		})
	}
}

func TestFirstRunAlerts(t *testing.T) {
	cases := []struct {
		config        Config
		expectErr     bool
		startupAlerts []string
		name          string
	}{
		{
			config:        Config{},
			expectErr:     false,
			startupAlerts: []string{},
			name:          "Empty",
		},
		{
			config:        Config{},
			expectErr:     true,
			startupAlerts: []string{"missing"},
			name:          "Unknown",
		},
		{
			config: Config{
				Alerts: []*Alert{
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
			config: Config{
				Alerts: []*Alert{
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
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Init()
			if err != nil {
				t.Errorf("sendFirstRunAlerts(%s): unexpected error reading config: %v", c.name, err)
			}

			err = sendStartupAlerts(&c.config, c.startupAlerts)
			if err == nil && c.expectErr {
				t.Errorf("sendFirstRunAlerts(%s): Expected error, the code did not error", c.name)
			} else if err != nil && !c.expectErr {
				t.Errorf("sendFirstRunAlerts(%s): Did not expect an error, but we got one anyway: %v", c.name, err)
			}
		})
	}
}
