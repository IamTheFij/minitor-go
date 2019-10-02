package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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
	for _, monitor := range config.Monitors {
		if !monitor.IsValid() {
			log.Printf("ERROR: Invalid monitor configuration: %s", monitor.Name)
			isValid = false
		}
	}

	for _, alert := range config.Alerts {
		if !alert.IsValid() {
			log.Printf("ERROR: Invalid alert configuration: %s", alert.Name)
			isValid = false
		}
	}

	return
}

// Init performs extra initialization on top of loading the config from file
func (config *Config) Init() {
	for name, alert := range config.Alerts {
		alert.Name = name
		alert.BuildTemplates()
	}
}

// LoadConfig will read config from the given path and parse it
func LoadConfig(filePath string) (config Config) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// TODO: Decide if this is better expanded here, or only when executing
	envExpanded := os.ExpandEnv(string(data))
	err = yaml.Unmarshal([]byte(envExpanded), &config)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		panic(err)
	}

	log.Printf("config:\n%v\n", config)

	if !config.IsValid() {
		panic("Cannot continue with invalid configuration")
	}

	// Finish initializing configuration
	config.Init()

	return config
}
