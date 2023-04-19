package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"git.iamthefij.com/iamthefij/slog"
)

var (
	// ExportMetrics will track whether or not we want to export metrics to prometheus
	ExportMetrics = false
	// MetricsPort is the port to expose metrics on
	MetricsPort = 8080
	// Metrics contains all active metrics
	Metrics = NewMetrics()

	// version of minitor being run
	version = "dev"

	errUnknownAlert = errors.New("unknown alert")
)

func sendAlerts(config *Config, monitor *Monitor, alertNotice *AlertNotice) error {
	slog.Debugf("Received an alert notice from %s", alertNotice.MonitorName)
	alertNames := monitor.GetAlertNames(alertNotice.IsUp)

	if alertNames == nil {
		// This should only happen for a recovery alert. AlertDown is validated not empty
		slog.Warningf(
			"Received alert, but no alert mechanisms exist. MonitorName=%s IsUp=%t",
			alertNotice.MonitorName, alertNotice.IsUp,
		)

		return nil
	}

	for _, alertName := range alertNames {
		if alert, ok := config.Alerts[alertName]; ok {
			output, err := alert.Send(*alertNotice)
			if err != nil {
				slog.Errorf(
					"Alert '%s' failed. result=%v: output=%s",
					alert.Name,
					err,
					output,
				)

				return err
			}

			// Count alert metrics
			Metrics.CountAlert(monitor.Name, alert.Name)
		} else {
			// This case should never actually happen since we validate against it
			slog.Errorf("Unknown alert for monitor %s: %s", alertNotice.MonitorName, alertName)

			return fmt.Errorf("unknown alert for monitor %s: %s: %w", alertNotice.MonitorName, alertName, errUnknownAlert)
		}
	}

	return nil
}

func checkMonitors(config *Config) error {
	// TODO: Run this in goroutines and capture exceptions
	for _, monitor := range config.Monitors {
		if monitor.ShouldCheck() {
			success, alertNotice := monitor.Check()
			hasAlert := alertNotice != nil

			// Track status metrics
			Metrics.SetMonitorStatus(monitor.Name, monitor.IsUp())
			Metrics.CountCheck(monitor.Name, success, monitor.LastCheckMilliseconds(), hasAlert)

			if alertNotice != nil {
				err := sendAlerts(config, monitor, alertNotice)
				// If there was an error in sending an alert, exit early and bubble it up
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	showVersion := flag.Bool("version", false, "Display the version of minitor and exit")
	configPath := flag.String("config", "config.yml", "Alternate configuration path (default: config.yml)")

	flag.BoolVar(&slog.DebugLevel, "debug", false, "Enables debug logs (default: false)")
	flag.BoolVar(&ExportMetrics, "metrics", false, "Enables prometheus metrics exporting (default: false)")
	flag.IntVar(&MetricsPort, "metrics-port", MetricsPort, "The port that Prometheus metrics should be exported on, if enabled. (default: 8080)")
	flag.Parse()

	// Print version if flag is provided
	if *showVersion {
		fmt.Println("Minitor version:", version)

		return
	}

	// Load configuration
	config, err := LoadConfig(*configPath)
	slog.OnErrFatalf(err, "Error loading config: %v", err)

	// Serve metrics exporter, if specified
	if ExportMetrics {
		slog.Infof("Exporting metrics to Prometheus on port %d", MetricsPort)

		go ServeMetrics()
	}

	// Start main loop
	for {
		err = checkMonitors(&config)
		slog.OnErrPanicf(err, "Error checking monitors")

		time.Sleep(config.CheckInterval)
	}
}
