package main

import (
	"errors"
	"io/ioutil"
	"time"

	"git.iamthefij.com/iamthefij/slog"
	"gopkg.in/yaml.v2"
)

var errInvalidConfig = errors.New("Invalid configuration")

// Config type is contains all provided user configuration
type Config struct {
	CheckInterval     SecondsOrDuration `yaml:"check_interval"`
	DefaultAlertAfter int16             `yaml:"default_alert_after"`
	DefaultAlertEvery *int16            `yaml:"default_alert_every"`
	DefaultAlertDown  []string          `yaml:"default_alert_down"`
	DefaultAlertUp    []string          `yaml:"default_alert_up"`
	Monitors          []*Monitor
	Alerts            map[string]*Alert
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

// SecondsOrDuration wraps a duration value for parsing a duration or seconds from YAML
// NOTE: This should be removed in favor of only parsing durations once compatibility is broken
type SecondsOrDuration struct {
	value time.Duration
}

// Value returns a duration value
func (sod SecondsOrDuration) Value() time.Duration {
	return sod.value
}

// UnmarshalYAML allows unmarshalling a duration value or seconds if an int was provided
func (sod *SecondsOrDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var seconds int64
	err := unmarshal(&seconds)

	if err == nil {
		sod.value = time.Second * time.Duration(seconds)

		return nil
	}

	// Error indicates that we don't have an int
	err = unmarshal(&sod.value)

	return err
}

// IsValid checks config validity and returns true if valid
func (config Config) IsValid() (isValid bool) {
	isValid = true

	// Validate alerts
	if config.Alerts == nil || len(config.Alerts) == 0 {
		// This should never happen because there is a default alert named 'log' for now
		slog.Errorf("Invalid alert configuration: Must provide at least one alert")

		isValid = false
	}

	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			slog.Errorf("Invalid alert configuration: %+v", alert.Name)

			isValid = false
		} else {
			slog.Debugf("Loaded alert %s", alert.Name)
		}
	}

	// Validate monitors
	if config.Monitors == nil || len(config.Monitors) == 0 {
		slog.Errorf("Invalid monitor configuration: Must provide at least one monitor")

		isValid = false
	}

	for _, monitor := range config.Monitors {
		if !monitor.IsValid() {
			slog.Errorf("Invalid monitor configuration: %s", monitor.Name)

			isValid = false
		}
		// Check that all Monitor alerts actually exist
		for _, isUp := range []bool{true, false} {
			for _, alertName := range monitor.GetAlertNames(isUp) {
				if _, ok := config.Alerts[alertName]; !ok {
					slog.Errorf(
						"Invalid monitor configuration: %s. Unknown alert %s",
						monitor.Name, alertName,
					)

					isValid = false
				}
			}
		}
	}

	return isValid
}

// Init performs extra initialization on top of loading the config from file
func (config *Config) Init() (err error) {
	for _, monitor := range config.Monitors {
		if monitor.AlertAfter == 0 && config.DefaultAlertAfter > 0 {
			monitor.AlertAfter = config.DefaultAlertAfter
		}

		if monitor.AlertEvery == nil && config.DefaultAlertEvery != nil {
			monitor.AlertEvery = config.DefaultAlertEvery
		}

		if len(monitor.AlertDown) == 0 && len(config.DefaultAlertDown) > 0 {
			monitor.AlertDown = config.DefaultAlertDown
		}

		if len(monitor.AlertUp) == 0 && len(config.DefaultAlertUp) > 0 {
			monitor.AlertUp = config.DefaultAlertUp
		}
	}

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

	slog.Debugf("Config values:\n%v\n", config)

	// Add log alert if not present
	if PyCompat {
		// Initialize alerts list if not present
		if config.Alerts == nil {
			config.Alerts = map[string]*Alert{}
		}

		if _, ok := config.Alerts["log"]; !ok {
			config.Alerts["log"] = NewLogAlert()
		}
	}

	// Finish initializing configuration
	if err = config.Init(); err != nil {
		return
	}

	if !config.IsValid() {
		err = errInvalidConfig

		return
	}

	return config, err
}
