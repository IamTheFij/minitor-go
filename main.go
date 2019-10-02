package main

import (
	"log"
	"time"
)

func main() {
	config := LoadConfig("config.yml")

	for {
		for _, monitor := range config.Monitors {
			if monitor.ShouldCheck() {
				_, alertNotice := monitor.Check()
				if alertNotice != nil {
					//log.Printf("Recieved an alert notice: %v", alertNotice)
					var alerts []string
					if alertNotice.IsUp {
						alerts = monitor.AlertUp
						log.Printf("Alert up: %v", monitor.AlertUp)
					} else {
						alerts = monitor.AlertDown
						log.Printf("Alert down: %v", monitor.AlertDown)
					}
					if alerts == nil {
						log.Printf("WARNING: Found alert, but no alert mechanism: %v", alertNotice)
					}
					for _, alertName := range alerts {
						if alert, ok := config.Alerts[alertName]; ok {
							alert.Send(*alertNotice)
						} else {
							log.Printf("WARNING: Could not find alert for %s", alertName)
						}
					}
				}
			}
		}

		sleepTime := time.Duration(config.CheckInterval) * time.Second
		time.Sleep(sleepTime)
	}
}
