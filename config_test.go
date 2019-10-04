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
		{"./test/does-not-exist", true, "Invalid config path"},
		{"./test/invalid-config-type.yml", true, "Invalid config type for key"},
		{"./test/invalid-config-missing-alerts.yml", true, "Invalid config missing alerts"},
	}

	for _, c := range cases {
		log.Printf("Testing case %s", c.name)
		_, err := LoadConfig(c.configPath)
		hasErr := (err != nil)
		if hasErr != c.expectErr {
			t.Errorf("LoadConfig(%v), expected=%v actual=%v", c.name, "Err", err)
			log.Printf("Case failed: %s", c.name)
		}
		log.Println("-----")
	}
}
