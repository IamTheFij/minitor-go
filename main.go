package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	// LogDebug will control whether debug messsages should be logged
	LogDebug = false

	// version of minitor being run
	version = "dev"
)

func checkMonitors(config *Config) error {
	for _, monitor := range config.Monitors {
		if monitor.ShouldCheck() {
			_, alertNotice := monitor.Check()

			// Should probably consider refactoring everything below here
			if alertNotice != nil {
				if LogDebug {
					log.Printf("DEBUG: Recieved an alert notice from %s", alertNotice.MonitorName)
				}
				alertNames := monitor.GetAlertNames(alertNotice.IsUp)
				if alertNames == nil {
					// This should only happen for a recovery alert. AlertDown is validated not empty
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
							return fmt.Errorf(
								"Unsuccessfully triggered alert '%s'. "+
									"Crashing to avoid false negatives: %v",
								alert.Name,
								err,
							)
						}
					} else {
						// This case should never actually happen since we validate against it
						log.Printf("ERROR: Unknown alert for monitor %s: %s", alertNotice.MonitorName, alertName)
						return fmt.Errorf("Unknown alert for monitor %s: %s", alertNotice.MonitorName, alertName)
					}
				}
			}
		}
	}

	return nil
}

func main() {
	// Get debug flag
	flag.BoolVar(&LogDebug, "debug", false, "Enables debug logs (default: false)")
	var showVersion = flag.Bool("version", false, "Display the version of minitor and exit")
	flag.Parse()

	// Print version if flag is provided
	if *showVersion {
		fmt.Println("Minitor version:", version)
		return
	}

	// Load configuration
	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start main loop
	for {
		err = checkMonitors(&config)
		if err != nil {
			panic(err)
		}

		sleepTime := time.Duration(config.CheckInterval) * time.Second
		time.Sleep(sleepTime)
	}
}
