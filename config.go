package main

import (
	"errors"
	"time"

	"git.iamthefij.com/iamthefij/slog"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

var errInvalidConfig = errors.New("Invalid configuration")

// Config type is contains all provided user configuration
type Config struct {
	CheckInterval time.Duration  `hcl:"check_interval"`
	Monitors      []*Monitor     `hcl:"monitor,block"`
	Alerts        AlertContainer `hcl:"alerts,block"`
}

func (p *Parser) decodeConfig(block *hcl.Block, ectx *hcl.EvalContext) (*Config, hcl.Diagnostics) {
	var b struct {
		CheckInterval string         `hcl:"check_interval"`
		Monitors      []*Monitor     `hcl:"monitor,block"`
		Alerts        AlertContainer `hcl:"alerts,block"`
	}
	diags := gohcl.DecodeBody(block.Body, ectx, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	config := &Config{
		Monitors: b.Monitors,
		Alerts:   b.Alerts,
	}

	// Set a default for CheckInterval
	if b.CheckInterval == "" {
		b.CheckInterval = "30s"
	}

	checkInterval, err := time.ParseDuration(b.CheckInterval)
	if err != nil {
		return nil, append(diags, &hcl.Diagnostic{
			Summary:  "Failed to parse check_interval duration",
			Severity: hcl.DiagError,
			Detail:   err.Error(),
			Subject:  &block.DefRange,
		})
	}
	config.CheckInterval = checkInterval

	return config, diags
}

// AlertContainer is struct wrapping map access to alerts
type AlertContainer struct {
	Alerts      []*Alert `hcl:"alert,block"`
	alertLookup map[string]*Alert
}

// Get returns an alert based on it's name
func (ac AlertContainer) Get(name string) (*Alert, bool) {
	// Build lookup map on first run
	if ac.alertLookup == nil {
		ac.alertLookup = map[string]*Alert{}
		for _, alert := range ac.Alerts {
			ac.alertLookup[alert.Name] = alert
		}
	}

	v, ok := ac.alertLookup[name]

	return v, ok
}

// IsEmpty checks if there are any defined alerts
func (ac AlertContainer) IsEmpty() bool {
	return ac.Alerts == nil || len(ac.Alerts) == 0
}

// BuildAllTemplates builds all alert templates
func (ac *AlertContainer) BuildAllTemplates() (err error) {
	for _, alert := range ac.Alerts {
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
	if config.Alerts.IsEmpty() {
		// This should never happen because there is a default alert named 'log' for now
		slog.Errorf("Invalid alert configuration: Must provide at least one alert")

		isValid = false
	}

	for _, alert := range config.Alerts.Alerts {
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
				if _, ok := config.Alerts.Get(alertName); !ok {
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
	err = config.Alerts.BuildAllTemplates()

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
