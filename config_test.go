package main

import (
	"log"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		configPath string
		expectErr  bool
		name       string
	}{
		{
			"./test/valid-config.hcl",
			false,
			"Valid config file",
		},
		{
			"./test/valid-default-log-alert.yml",
			true,
			"Invalid config file no log alert",
		},
		{
			"./test/does-not-exist",
			true,
			"Invalid config path",
		},
		{
			"./test/invalid-config-type.yml",
			true,
			"Invalid config type for key",
		},
		{
			"./test/invalid-config-missing-alerts.yml",
			true,
			"Invalid config missing alerts",
		},
		{
			"./test/invalid-config-unknown-alert.yml",
			true,
			"Invalid config unknown alert",
		},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)
		// Set PyCompat based on compatibility mode
		_, err := LoadConfig(c.configPath)
		hasErr := (err != nil)

		if hasErr != c.expectErr {
			t.Errorf("LoadConfig(%v), expected_error=%v actual=%v", c.name, c.expectErr, err)
			log.Printf("Case failed: %s", c.name)
		}
	}
}

func TestIntervalParsing(t *testing.T) {
	log.Printf("Testing case TestIntervalParsing")

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

	log.Println("-----")
}

// TestMultiLineConfig is a more complicated test stepping through the parsing
// and execution of mutli-line strings presented in YAML
func TestMultiLineConfig(t *testing.T) {
	log.Println("Testing multi-line string config")

	config, err := LoadConfig("./test/valid-verify-multi-line.yml")
	if err != nil {
		t.Fatalf("TestMultiLineConfig(load), expected=no_error actual=%v", err)
	}

	log.Println("-----")
	log.Println("TestMultiLineConfig(parse > string)")

	expected := "echo 'Some string with stuff'; echo \"<angle brackets>\"; exit 1\n"
	actual := config.Monitors[0].ShellCommand

	if expected != actual {
		t.Errorf("TestMultiLineConfig(>) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	log.Println("-----")
	log.Println("TestMultiLineConfig(execute > string)")

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

	log.Println("-----")
	log.Println("TestMultiLineConfig(parse | string)")

	expected = "echo 'Some string with stuff'\necho '<angle brackets>'\n"

	alert, ok := config.Alerts.Get("log_shell")
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

	log.Println("-----")
	log.Println("TestMultiLineConfig(execute | string)")

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
