package main

import (
	"log"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		configPath string
		expectErr  bool
		name       string
	}{
		{"./test/valid-config.yml", false, "Valid config file"},
		{"./test/valid-default-log-alert.yml", false, "Valid config file with default log alert"},
		{"./test/does-not-exist", true, "Invalid config path"},
		{"./test/invalid-config-type.yml", true, "Invalid config type for key"},
		{"./test/invalid-config-missing-alerts.yml", true, "Invalid config missing alerts"},
		{"./test/invalid-config-unknown-alert.yml", true, "Invalid config unknown alert"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)
		_, err := LoadConfig(c.configPath)
		hasErr := (err != nil)
		if hasErr != c.expectErr {
			t.Errorf("LoadConfig(%v), expected_error=%v actual=%v", c.name, c.expectErr, err)
			log.Printf("Case failed: %s", c.name)
		}
		log.Println("-----")
	}
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
	actual := config.Monitors[0].Command.ShellCommand
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

	log.Println("-----")
	log.Println("TestMultiLineConfig(parse | string)")
	expected = "echo 'Some string with stuff'\necho '<angle brackets>'\n"
	actual = config.Alerts["log_shell"].Command.ShellCommand
	if expected != actual {
		t.Errorf("TestMultiLineConfig(|) failed")
		t.Logf("string expected=`%v`", expected)
		t.Logf("string actual  =`%v`", actual)
		t.Logf("bytes expected=%v", []byte(expected))
		t.Logf("bytes actual  =%v", []byte(actual))
	}

	log.Println("-----")
	log.Println("TestMultiLineConfig(execute | string)")
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
