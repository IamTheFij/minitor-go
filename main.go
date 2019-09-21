package main

import (
	"time"
)

func main() {
	config := LoadConfig("config.yml")

	for {
		for _, monitor := range config.Monitors {
			if monitor.ShouldCheck() {
				monitor.Check()
			}
		}

		sleepTime := time.Duration(config.CheckInterval) * time.Second
		time.Sleep(sleepTime)
	}
}
