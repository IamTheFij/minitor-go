package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		configPath string
		expectErr  bool
		name       string
	}{
		{"./test/does-not-exist", true, "Invalid config path"},
		// {"./test/invalid-config-missing-alerts.yml", true, "Invalid config missing alerts"},
		// {"./test/invalid-config-type.yml", true, "Invalid config type for key"},
		// {"./test/invalid-config-unknown-alert.yml", true, "Invalid config unknown alert"},
		// {"./test/valid-config-default-values.yml", false, "Valid config file with default values"},
		{"./test/valid-config.hcl", false, "Valid config file"},
		// {"./test/valid-default-log-alert.yml", true, "Invalid config file no log alert"},
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

// TestMultiLineConfig is a more complicated test stepping through the parsing
// and execution of mutli-line strings presented in YAML
func TestMultiLineConfig(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig("./test/valid-verify-multi-line.hcl")
	if err != nil {
		t.Fatalf("TestMultiLineConfig(load), expected=no_error actual=%v", err)
	}

	expected := "echo 'Some string with stuff';\necho \"<angle brackets>\";\nexit 1\n"
	actual := config.Monitors[0].ShellCommand

	if expected != actual {
		t.Errorf("TestMultiLineConfig(>) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	_, notice := config.Monitors[0].Check()
	if notice == nil {
		t.Fatal("Did not receive an alert notice")
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

	alert, ok := config.GetAlert("log_shell")
	if !ok {
		t.Fatal("Could not find expected alert 'log_shell'")
	}

	actual = alert.ShellCommand
	if expected != actual {
		t.Errorf("TestMultiLineConfig(|) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	actual, err = alert.Send(AlertNotice{})
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
