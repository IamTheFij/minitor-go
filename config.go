package main

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Config type is contains all provided user configuration
type Config struct {
	CheckInterval int64 `yaml:"check_interval"`
	Monitors      []*Monitor
	Alerts        map[string]*Alert
}

// CommandOrShell type wraps a string or list of strings
// for executing a command directly or in a shell
type CommandOrShell struct {
	ShellCommand string
	Command      []string
}

// Empty checks if the Command has a value
func (cos CommandOrShell) Empty() bool {
	return (cos.ShellCommand == "" && cos.Command == nil)
}

// UnmarshalYAML allows unmarshalling either a string or slice of strings
// and parsing them as either a command or a shell command.
func (cos *CommandOrShell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cmd []string
	err := unmarshal(&cmd)
	// Error indicates this is shell command
	if err != nil {
		var shellCmd string
		err := unmarshal(&shellCmd)
		if err != nil {
			return err
		}
		cos.ShellCommand = shellCmd
	} else {
		cos.Command = cmd
	}
	return nil
}

// IsValid checks config validity and returns true if valid
func (config Config) IsValid() (isValid bool) {
	isValid = true

	// Validate alerts
	if config.Alerts == nil || len(config.Alerts) == 0 {
		// This should never happen because there is a default alert named 'log' for now
		log.Printf("ERROR: Invalid alert configuration: Must provide at least one alert")
		isValid = false
	}
	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			log.Printf("ERROR: Invalid alert configuration: %s", alert.Name)
			isValid = false
		}
	}

	// Validate monitors
	if config.Monitors == nil || len(config.Monitors) == 0 {
		log.Printf("ERROR: Invalid monitor configuration: Must provide at least one monitor")
		isValid = false
	}
	for _, monitor := range config.Monitors {
		if !monitor.IsValid() {
			log.Printf("ERROR: Invalid monitor configuration: %s", monitor.Name)
			isValid = false
		}
		// Check that all Monitor alerts actually exist
		for _, isUp := range []bool{true, false} {
			for _, alertName := range monitor.GetAlertNames(isUp) {
				if _, ok := config.Alerts[alertName]; !ok {
					log.Printf(
						"ERROR: Invalid monitor configuration: %s. Unknown alert %s",
						monitor.Name, alertName,
					)
					isValid = false
				}
			}
		}
	}

	return
}

// Init performs extra initialization on top of loading the config from file
func (config *Config) Init() (err error) {
	for name, alert := range config.Alerts {
		alert.Name = name
		if err = alert.BuildTemplates(); err != nil {
			return
		}
	}

	return
}

// LoadConfig will read config from the given path and parse it
func LoadConfig(filePath string) (config Config, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return
	}

	if LogDebug {
		log.Printf("DEBUG: Config values:\n%v\n", config)
	}

	// Add log alert if not present
	if PyCompat {
		// Intialize alerts list if not present
		if config.Alerts == nil {
			config.Alerts = map[string]*Alert{}
		}
		if _, ok := config.Alerts["log"]; !ok {
			config.Alerts["log"] = NewLogAlert()
		}
	}

	if !config.IsValid() {
		err = errors.New("Invalid configuration")
		return
	}

	// Finish initializing configuration
	err = config.Init()

	return
}
