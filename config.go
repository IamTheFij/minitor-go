package main

import (
	"errors"
	"fmt"
	"time"

	"git.iamthefij.com/iamthefij/slog"
	/*
	 * "github.com/hashicorp/hcl/v2"
	 * "github.com/hashicorp/hcl/v2/gohcl"
	 */
	"github.com/hashicorp/hcl/v2/hclsimple"
)

var errInvalidConfig = errors.New("Invalid configuration")

// Config type is contains all provided user configuration
type Config struct {
	CheckIntervalStr string `hcl:"check_interval"`
	CheckInterval    time.Duration

	DefaultAlertAfter *int       `hcl:"default_alert_after,optional"`
	DefaultAlertEvery *int       `hcl:"default_alert_every,optional"`
	DefaultAlertDown  []string   `hcl:"default_alert_down,optional"`
	DefaultAlertUp    []string   `hcl:"default_alert_up,optional"`
	Monitors          []*Monitor `hcl:"monitor,block"`
	Alerts            []*Alert   `hcl:"alert,block"`

	alertLookup map[string]*Alert
}

func (c Config) GetAlert(name string) (*Alert, bool) {
	if c.alertLookup == nil {
		c.alertLookup = map[string]*Alert{}
		for _, alert := range c.Alerts {
			c.alertLookup[alert.Name] = alert
		}
	}

	v, ok := c.alertLookup[name]

	return v, ok
}

// BuildAllTemplates builds all alert templates
func (c *Config) BuildAllTemplates() (err error) {
	for _, alert := range c.Alerts {
		if err = alert.BuildTemplates(); err != nil {
			return
		}
	}

	return
}

// IsValid checks config validity and returns true if valid
func (config Config) IsValid() (isValid bool) {
	isValid = true

	// Validate alerts
	if len(config.Alerts) == 0 {
		// This should never happen because there is a default alert named 'log' for now
		slog.Errorf("Invalid alert configuration: Must provide at least one alert")

		isValid = false
	}

	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			slog.Errorf("Invalid alert configuration: %+v", alert.Name)

			isValid = false
		}
	}

	// Validate monitors
	if len(config.Monitors) == 0 {
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
				if _, ok := config.GetAlert(alertName); !ok {
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
	config.CheckInterval, err = time.ParseDuration(config.CheckIntervalStr)
	if err != nil {
		return fmt.Errorf("failed to parse top level check_interval duration: %w", err)
	}

	for _, monitor := range config.Monitors {
		// Parse the check_interval string into a time.Duration
		if monitor.CheckIntervalStr != nil {
			monitor.CheckInterval, err = time.ParseDuration(*monitor.CheckIntervalStr)
			if err != nil {
				return fmt.Errorf("failed to parse check_interval duration for monitor %s: %w", monitor.Name, err)
			}
		}

		// Set default values for monitor alerts
		if monitor.AlertAfter == nil {
			monitor.AlertAfter = config.DefaultAlertAfter
		}

		if monitor.AlertEvery == nil {
			monitor.AlertEvery = config.DefaultAlertEvery
		}

		if monitor.AlertDown == nil {
			monitor.AlertDown = config.DefaultAlertDown
		}

		if monitor.AlertUp == nil {
			monitor.AlertUp = config.DefaultAlertUp
		}
	}

	err = config.BuildAllTemplates()

	return
}

// LoadConfig will read config from the given path and parse it
func LoadConfig(filePath string) (config Config, err error) {
	err = hclsimple.DecodeFile(filePath, nil, &config)
	if err != nil {
		return
	}

	slog.Debugf("Config values:\n%v\n", config)

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
