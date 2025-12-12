package main_test

import (
	"errors"
	"testing"
	"time"

	m "git.iamthefij.com/iamthefij/minitor-go"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		configPath  string
		expectedErr error
		name        string
	}{
		{"./test/does-not-exist", m.ErrLoadingConfig, "Invalid config path"},
		{"./test/invalid-config-wrong-hcl-type.hcl", m.ErrLoadingConfig, "Incorrect HCL type"},
		{"./test/invalid-config-missing-alerts.hcl", m.ErrNoAlerts, "Invalid config missing alerts"},
		{"./test/invalid-config-missing-alerts.hcl", m.ErrInvalidConfig, "Invalid config general"},
		{"./test/invalid-config-invalid-duration.hcl", m.ErrConfigInit, "Invalid config type for key"},
		{"./test/invalid-config-unknown-alert.hcl", m.ErrUnknownAlert, "Invalid config unknown alert"},
		{"./test/valid-config-default-values.hcl", nil, "Valid config file with default values"},
		{"./test/valid-config.hcl", nil, "Valid config file"},
	}
	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			_, err := m.LoadConfig(c.configPath)
			hasErr := (err != nil)
			expectErr := (c.expectedErr != nil)

			if hasErr != expectErr || !errors.Is(err, c.expectedErr) {
				t.Errorf("LoadConfig(%v), expected_error=%v actual=%v", c.name, c.expectedErr, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cases := []struct {
		configPath     string
		expectedResult m.Config
		name           string
	}{
		{
			"./test/valid-config-default-values.hcl",
			m.Config{
				CheckInterval:     1 * time.Second,
				DefaultAlertAfter: 2,
				DefaultAlertEvery: Ptr(0),
				DefaultAlertDown:  []string{"log_command"},
			},
			"override defaults",
		},
		{
			"./test/valid-config.hcl",
			m.Config{
				CheckInterval:     30 * time.Second,
				DefaultAlertAfter: 1,
				DefaultAlertEvery: Ptr(-1),
				DefaultAlertDown:  []string{},
			},
			"default defaults",
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			config, err := m.LoadConfig(c.configPath)
			if err != nil {
				t.Errorf("Got error when loading config file %q: %s", c.configPath, err)
			}

			// Test Config has default values
			if config.DefaultAlertAfter != c.expectedResult.DefaultAlertAfter {
				t.Errorf("Got unexpected DefaultAlertAfter from file %q: expected=%v actual=%v", c.configPath, c.expectedResult.DefaultAlertAfter, config.DefaultAlertAfter)
			}

			if *config.DefaultAlertEvery != *c.expectedResult.DefaultAlertEvery {
				t.Errorf("Got unexpected DefaultAlertEvery from file %q: expected=%v actual=%v", c.configPath, *c.expectedResult.DefaultAlertEvery, *config.DefaultAlertEvery)
			}

			if !m.EqualSliceString(config.DefaultAlertUp, c.expectedResult.DefaultAlertUp) {
				t.Errorf("Got unexpected DefaultAlertUp from file %q: expected=%v actual=%v", c.configPath, c.expectedResult.DefaultAlertUp, config.DefaultAlertUp)
			}

			if !m.EqualSliceString(config.DefaultAlertDown, c.expectedResult.DefaultAlertDown) {
				t.Errorf("Got unexpected DefaultAlertDown from file %q: expected=%v actual=%v", c.configPath, c.expectedResult.DefaultAlertDown, config.DefaultAlertDown)
			}

			// Check that monitor defaults propagate
			var defaultMonitor *m.Monitor
			for _, monitor := range config.Monitors {
				if monitor.Name == "Default" {
					defaultMonitor = monitor
				}
			}

			if defaultMonitor == nil {
				t.Errorf("failed to find default monitor in %q", c.configPath)
			}

			if defaultMonitor.AlertAfter != c.expectedResult.DefaultAlertAfter {
				t.Errorf("Got unexpected AlertAfter from file %q: expected=%v actual=%v", c.configPath, c.expectedResult.DefaultAlertAfter, defaultMonitor.AlertAfter)
			}

			if *defaultMonitor.AlertEvery != *c.expectedResult.DefaultAlertEvery {
				t.Errorf("Got unexpected AlertEvery from file %q: expected=%v actual=%v", c.configPath, *c.expectedResult.DefaultAlertEvery, *defaultMonitor.AlertEvery)
			}

			if !m.EqualSliceString(defaultMonitor.AlertUp, c.expectedResult.DefaultAlertUp) {
				t.Errorf("Got unexpected AlertUp from file %q: expected=%v actual=%v", c.configPath, c.expectedResult.DefaultAlertUp, defaultMonitor.AlertUp)
			}

			// NOTE: Can't compare AlertDown because default is empty and that is invalid
		})
	}
}

// TestMultiLineConfig is a more complicated test stepping through the parsing
// and execution of mutli-line strings presented in YAML
func TestMultiLineConfig(t *testing.T) {
	t.Parallel()

	config, err := m.LoadConfig("./test/valid-verify-multi-line.hcl")
	if err != nil {
		t.Fatalf("TestMultiLineConfig(load), expected=no_error actual=%v", err)
	}

	t.Run("Test Monitor with Indented Multi-Line String", func(t *testing.T) {
		// Verify indented heredoc is as expected
		expected := "echo 'Some string with stuff'\necho \"<angle brackets>\"\nexit 1\n"
		actual := config.Monitors[0].ShellCommand

		if expected != actual {
			t.Error("Heredoc mismatch")
			t.Errorf("string expected=`%v`", expected)
			t.Errorf("string actual  =`%v`", actual)
		}

		// Run the monitor and verify the output
		_, notice := config.Monitors[0].Check()
		if notice == nil {
			t.Fatal("Did not receive an alert notice and should have")
		}

		// Verify the output of the monitor is as expected
		expected = "Some string with stuff\n<angle brackets>\n"
		actual = notice.LastCheckOutput

		if expected != actual {
			t.Error("Output mismatch")
			t.Errorf("string expected=`%v`", expected)
			t.Errorf("string actual  =`%v`", actual)
		}
	})

	t.Run("Test Alert with Multi-Line String", func(t *testing.T) {
		alert, ok := config.GetAlert("log_shell")
		if !ok {
			t.Fatal("Could not find expected alert 'log_shell'")
		}

		expected := "    echo 'Some string with stuff'\n    echo '<angle brackets>'\n"
		actual := alert.ShellCommand

		if expected != actual {
			t.Error("Heredoc mismatch")
			t.Errorf("string expected=`%v`", expected)
			t.Errorf("string actual  =`%v`", actual)
		}

		actual, err = alert.Send(m.AlertNotice{})
		if err != nil {
			t.Fatal("Execution of alert failed")
		}

		expected = "Some string with stuff\n<angle brackets>\n"
		if expected != actual {
			t.Error("Output mismatch")
			t.Errorf("string expected=`%v`", expected)
			t.Errorf("string actual  =`%v`", actual)
		}
	})
}
