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
						Command: []string{"true"},
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
						Command:    []string{"false"},
						AlertAfter: 1,
					},
					&Monitor{
						Name:       "Failure",
						Command:    []string{"false"},
						AlertDown:  []string{"unknown"},
						AlertAfter: 1,
					},
				},
			},
			expectErr: false,
			name:      "Monitor failure, no and unknown alerts",
		},
		{
			config: Config{
				Monitors: []*Monitor{
					&Monitor{
						Name:       "Success",
						Command:    []string{"ls"},
						alertCount: 1,
					},
					&Monitor{
						Name:       "Success",
						Command:    []string{"true"},
						AlertUp:    []string{"unknown"},
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
						Command:    []string{"false"},
						AlertDown:  []string{"good"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"good": &Alert{
						Command: []string{"true"},
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
						Command:    []string{"false"},
						AlertDown:  []string{"bad"},
						AlertAfter: 1,
					},
				},
				Alerts: map[string]*Alert{
					"bad": &Alert{
						Name:    "bad",
						Command: []string{"false"},
					},
				},
			},
			expectErr: true,
			name:      "Monitor failure, bad alert",
		},
	}

	for _, c := range cases {
		c.config.Init()
		err := checkMonitors(&c.config)
		if err == nil && c.expectErr {
			t.Errorf("checkMonitors(%s): Expected panic, the code did not panic", c.name)
		}
	}
}
