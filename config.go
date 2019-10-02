package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	CheckInterval int64 `yaml:"check_interval"`
	Monitors      []*Monitor
	Alerts        map[string]*Alert
}

func LoadConfig(filePath string) (config Config) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// TODO: Decide if this is better expanded here, or only when executing
	env_expanded := os.ExpandEnv(string(data))

	err = yaml.Unmarshal([]byte(env_expanded), &config)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		panic(err)
	}

	log.Printf("config:\n%v\n", config)

	return config
}
