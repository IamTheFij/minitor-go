package main

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		configPath string
		expectErr  bool
		name       string
		pyCompat   bool
	}{
		{"./test/valid-config.yml", false, "Valid config file", false},
		{"./test/valid-config-default-values.yml", false, "Valid config file with default values", false},
		{"./test/valid-default-log-alert.yml", true, "Invalid config file no log alert", false},
		{"./test/does-not-exist", true, "Invalid config path", false},
		{"./test/invalid-config-type.yml", true, "Invalid config type for key", false},
		{"./test/invalid-config-missing-alerts.yml", true, "Invalid config missing alerts", false},
		{"./test/invalid-config-unknown-alert.yml", true, "Invalid config unknown alert", false},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadConfig(c.configPath)
			hasErr := (err != nil)

			if hasErr != c.expectErr {
				t.Errorf("LoadConfig(%v), expected_error=%v actual=%v", c.name, c.expectErr, err)
			}
		})
	}
}

func TestIntervalParsing(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig("./test/valid-config.yml")
	if err != nil {
		t.Errorf("Failed loading config: %v", err)
	}

	oneSecond := time.Second
	tenSeconds := 10 * time.Second
	oneMinute := time.Minute

	// validate top level interval seconds represented as an int
	if config.CheckInterval != oneSecond {
		t.Errorf("Incorrectly parsed int seconds. expected=%v actual=%v", oneSecond, config.CheckInterval)
	}

	if config.Monitors[0].CheckInterval != tenSeconds {
		t.Errorf("Incorrectly parsed seconds duration. expected=%v actual=%v", oneSecond, config.CheckInterval)
	}

	if config.Monitors[1].CheckInterval != oneMinute {
		t.Errorf("Incorrectly parsed seconds duration. expected=%v actual=%v", oneSecond, config.CheckInterval)
	}
}

// TestMultiLineConfig is a more complicated test stepping through the parsing
// and execution of mutli-line strings presented in YAML
func TestMultiLineConfig(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig("./test/valid-verify-multi-line.yml")
	if err != nil {
		t.Fatalf("TestMultiLineConfig(load), expected=no_error actual=%v", err)
	}

	expected := "echo 'Some string with stuff'; echo \"<angle brackets>\"; exit 1\n"
	actual := config.Monitors[0].Command.ShellCommand

	if expected != actual {
		t.Errorf("TestMultiLineConfig(>) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	_, notice := config.Monitors[0].Check()
	if notice == nil {
		t.Fatalf("Did not receive an alert notice")
	}

	expected = "Some string with stuff\n<angle brackets>\n"
	actual = notice.LastCheckOutput

	if expected != actual {
		t.Errorf("TestMultiLineConfig(execute > string) check failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	expected = "echo 'Some string with stuff'\necho '<angle brackets>'\n"
	actual = config.Alerts["log_shell"].Command.ShellCommand

	if expected != actual {
		t.Errorf("TestMultiLineConfig(|) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	actual, err = config.Alerts["log_shell"].Send(AlertNotice{})
	if err != nil {
		t.Errorf("Execution of alert failed")
	}

	expected = "Some string with stuff\n<angle brackets>\n"
	if expected != actual {
		t.Errorf("TestMultiLineConfig(execute | string) check failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}
}
