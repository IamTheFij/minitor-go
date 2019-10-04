package main

import (
	"fmt"
	"log"
	"time"
)

func checkMonitors(config *Config) {
	for _, monitor := range config.Monitors {
		if monitor.ShouldCheck() {
			_, alertNotice := monitor.Check()

			// Should probably consider refactoring everything below here
			if alertNotice != nil {
				log.Printf("DEBUG: Recieved an alert notice from %s", alertNotice.MonitorName)
				alertNames := monitor.GetAlertNames(alertNotice.IsUp)
				if alertNames == nil {
					// TODO: Should this be a panic? Should this be validated against? Probably
					log.Printf(
						"WARNING: Recieved alert, but no alert mechanisms exist. MonitorName=%s IsUp=%t",
						alertNotice.MonitorName, alertNotice.IsUp,
					)
				}
				for _, alertName := range alertNames {
					if alert, ok := config.Alerts[alertName]; ok {
						output, err := alert.Send(*alertNotice)
						if err != nil {
							log.Printf(
								"ERROR: Alert '%s' failed. result=%v: output=%s",
								alert.Name,
								err,
								output,
							)
							// TODO: Maybe return this error instead of panicking here
							panic(fmt.Errorf(
								"ERROR: Unsuccessfully triggered alert '%s'. "+
									"Crashing to avoid false negatives: %v",
								alert.Name,
								err,
							))
						}
					} else {
						// TODO: Maybe panic here. Also, probably validate up front
						log.Printf("ERROR: Alert with name '%s' not found", alertName)
					}
				}
			}
		}
	}
}

func main() {
	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start main loop
	for {
		checkMonitors(&config)

		sleepTime := time.Duration(config.CheckInterval) * time.Second
		time.Sleep(sleepTime)
	}
}
