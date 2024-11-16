package main

import (
	"errors"
	"fmt"
	"time"

	"git.iamthefij.com/iamthefij/slog"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

var (
	ErrLoadingConfig  = errors.New("Failed to load or parse configuration")
	ErrConfigInit     = errors.New("Failed to initialize configuration")
	ErrInvalidConfig  = errors.New("Invalid configuration")
	ErrNoAlerts       = errors.New("No alerts provided")
	ErrInvalidAlert   = errors.New("Invalid alert configuration")
	ErrNoMonitors     = errors.New("No monitors provided")
	ErrInvalidMonitor = errors.New("Invalid monitor configuration")
	ErrUnknownAlert   = errors.New("Unknown alert")
)

// Config type is contains all provided user configuration
type Config struct {
	CheckIntervalStr string `hcl:"check_interval"`
	CheckInterval    time.Duration

	DefaultAlertAfter int        `hcl:"default_alert_after,optional"`
	DefaultAlertEvery *int       `hcl:"default_alert_every,optional"`
	DefaultAlertDown  []string   `hcl:"default_alert_down,optional"`
	DefaultAlertUp    []string   `hcl:"default_alert_up,optional"`
	Monitors          []*Monitor `hcl:"monitor,block"`
	Alerts            []*Alert   `hcl:"alert,block"`

	alertLookup map[string]*Alert
}

// Init performs extra initialization on top of loading the config from file
func (config *Config) Init() (err error) {
	config.CheckInterval, err = time.ParseDuration(config.CheckIntervalStr)
	if err != nil {
		return fmt.Errorf("failed to parse top level check_interval duration: %w", err)
	}

	if config.DefaultAlertAfter == 0 {
		minAlertAfter := 1
		config.DefaultAlertAfter = minAlertAfter
	}

	for _, monitor := range config.Monitors {
		if err = monitor.Init(
			config.DefaultAlertAfter,
			config.DefaultAlertEvery,
			config.DefaultAlertDown,
			config.DefaultAlertUp,
		); err != nil {
			return
		}
	}

	err = config.BuildAllTemplates()

	return
}

// IsValid checks config validity and returns true if valid
func (config Config) IsValid() error {
	var err error

	// Validate alerts
	if len(config.Alerts) == 0 {
		err = errors.Join(err, ErrNoAlerts)
	}

	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			err = errors.Join(err, fmt.Errorf("%w: %s", ErrInvalidAlert, alert.Name))
		}
	}

	// Validate monitors
	if len(config.Monitors) == 0 {
		err = errors.Join(err, ErrNoMonitors)
	}

	for _, monitor := range config.Monitors {
		if !monitor.IsValid() {
			err = errors.Join(err, fmt.Errorf("%w: %s", ErrInvalidMonitor, monitor.Name))
		}

		// Check that all Monitor alerts actually exist
		for _, isUp := range []bool{true, false} {
			for _, alertName := range monitor.GetAlertNames(isUp) {
				if _, ok := config.GetAlert(alertName); !ok {
					err = errors.Join(
						err,
						fmt.Errorf("%w: %s. %w: %s", ErrInvalidMonitor, monitor.Name, ErrUnknownAlert, alertName),
					)
				}
			}
		}
	}

	return err
}

// GetAlert returns an alert by name
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

// LoadConfig will read config from the given path and parse it
func LoadConfig(filePath string) (Config, error) {
	var config Config

	if err := hclsimple.DecodeFile(filePath, nil, &config); err != nil {
		slog.Debugf("Failed to load config from %s: %v", filePath, err)
		return config, errors.Join(ErrLoadingConfig, err)
	}

	slog.Debugf("Config values:\n%v\n", config)

	// Finish initializing configuration
	if err := config.Init(); err != nil {
		return config, errors.Join(ErrConfigInit, err)
	}

	if err := config.IsValid(); err != nil {
		return config, errors.Join(ErrInvalidConfig, err)
	}

	return config, nil
}
