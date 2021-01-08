package main

import "testing"

func TestCheckMonitors(t *testing.T) {
	cases := []struct {
		config    Config
		expectErr bool
		name      string
	}{
		{
			config:    Config{},
			expectErr: false,
			name:      "Empty",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:    "Success",
						Command: CommandOrShell{Command: []string{"true"}},
					},
				},
			},
			expectErr: false,
			name:      "Monitor success, no alerts",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertAfter: 1,
					},
				},
			},
			expectErr: false,
			name:      "Monitor failure, no alerts",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Success",
						Command:    CommandOrShell{Command: []string{"ls"}},
						alertCount: 1,
					},
				},
			},
			expectErr: false,
			name:      "Monitor recovery, no alerts",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"unknown"},
						AlertAfter: 1,
					},
				},
			},
			expectErr: true,
			name:      "Monitor failure, unknown alerts",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Success",
						Command:    CommandOrShell{Command: []string{"true"}},
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
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"good"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"good": &Alert{
						Command: CommandOrShell{Command: []string{"true"}},
					},
				},
			},
			expectErr: false,
			name:      "Monitor failure, successful alert",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Failure",
						Command:    CommandOrShell{Command: []string{"false"}},
						AlertDown:  []string{"bad"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"bad": &Alert{
						Name:    "bad",
						Command: CommandOrShell{Command: []string{"false"}},
					},
				},
			},
			expectErr: true,
			name:      "Monitor failure, bad alert",
		},
	}

	for _, c := range cases {
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
