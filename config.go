package main

import (
	"errors"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config type is contains all provided user configuration
type Config struct {
	CheckInterval int64 `yaml:"check_interval"`
	Monitors      []*Monitor
	Alerts        map[string]*Alert
}

// IsValid checks config validity and returns true if valid
func (config Config) IsValid() (isValid bool) {
	isValid = true

	// Validate monitors
	if config.Monitors == nil || len(config.Monitors) == 0 {
		log.Errorf("Invalid monitor configuration: Must provide at least one monitor")
		isValid = false
	}
	for _, monitor := range config.Monitors {
		if !monitor.IsValid() {
			log.Errorf("Invalid monitor configuration: %s", monitor.Name)
			isValid = false
		}
		// Check that all Monitor alerts actually exist
		for _, isUp := range []bool{true, false} {
			for _, alertName := range monitor.GetAlertNames(isUp) {
				if _, ok := config.Alerts[alertName]; !ok {
					log.Errorf(
						"Invalid monitor configuration: %s. Unknown alert %s",
						monitor.Name, alertName,
					)
					isValid = false
				}
			}
		}
	}

	// Validate alerts
	if config.Alerts == nil || len(config.Alerts) == 0 {
		log.Errorf("Invalid alert configuration: Must provide at least one alert")
		isValid = false
	}
	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			log.Errorf("Invalid alert configuration: %s", alert.Name)
			isValid = false
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

	// TODO: Decide if this is better expanded here, or only when executing
	envExpanded := os.ExpandEnv(string(data))
	err = yaml.Unmarshal([]byte(envExpanded), &config)
	if err != nil {
		return
	}

	log.Debugf("Config values:\n%v\n", config)

	if !config.IsValid() {
		err = errors.New("Invalid configuration")
		return
	}

	// Finish initializing configuration
	err = config.Init()

	return
}
