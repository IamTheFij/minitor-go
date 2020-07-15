package main

import (
	"log"
	"math"
	"os/exec"
	"time"
)

// Monitor represents a particular periodic check of a command
type Monitor struct {
	// Config values
	Name          string
	Command       CommandOrShell
	AlertDown     []string `yaml:"alert_down"`
	AlertUp       []string `yaml:"alert_up"`
	CheckInterval float64  `yaml:"check_interval"`
	AlertAfter    int16    `yaml:"alert_after"`
	AlertEvery    int16    `yaml:"alert_every"`
	// Other values
	lastCheck    time.Time
	lastOutput   string
	alertCount   int16
	failureCount int16
	lastSuccess  time.Time
}

// IsValid returns a boolean indicating if the Monitor has been correctly
// configured
func (monitor Monitor) IsValid() bool {
	return (!monitor.Command.Empty() &&
		monitor.getAlertAfter() > 0 &&
		monitor.AlertDown != nil)
}

// ShouldCheck returns a boolean indicating if the Monitor is ready to be
// be checked again
func (monitor Monitor) ShouldCheck() bool {
	if monitor.lastCheck.IsZero() {
		return true
	}

	sinceLastCheck := time.Now().Sub(monitor.lastCheck).Seconds()
	return sinceLastCheck >= monitor.CheckInterval
}

// Check will run the command configured by the Monitor and return a status
// and a possible AlertNotice
func (monitor *Monitor) Check() (bool, *AlertNotice) {
	var cmd *exec.Cmd
	if monitor.Command.Command != nil {
		cmd = exec.Command(monitor.Command.Command[0], monitor.Command.Command[1:]...)
	} else {
		cmd = ShellCommand(monitor.Command.ShellCommand)
	}

	output, err := cmd.CombinedOutput()
	monitor.lastCheck = time.Now()
	monitor.lastOutput = string(output)

	var alertNotice *AlertNotice
	isSuccess := (err == nil)
	if isSuccess {
		alertNotice = monitor.success()
	} else {
		alertNotice = monitor.failure()
	}

	if LogDebug {
		log.Printf("DEBUG: Command output: %s", monitor.lastOutput)
	}
	if err != nil {
		if LogDebug {
			log.Printf("DEBUG: Command result: %v", err)
		}
	}

	log.Printf(
		"INFO: %s success=%t, alert=%t",
		monitor.Name,
		isSuccess,
		alertNotice != nil,
	)

	return isSuccess, alertNotice
}

// IsUp returns the status of the current monitor
func (monitor Monitor) IsUp() bool {
	return monitor.alertCount == 0
}

func (monitor *Monitor) success() (notice *AlertNotice) {
	if !monitor.IsUp() {
		// Alert that we have recovered
		notice = monitor.createAlertNotice(true)
	}
	monitor.failureCount = 0
	monitor.alertCount = 0
	monitor.lastSuccess = time.Now()

	return
}

func (monitor *Monitor) failure() (notice *AlertNotice) {
	monitor.failureCount++
	// If we haven't hit the minimum failures, we can exit
	if monitor.failureCount < monitor.getAlertAfter() {
		if LogDebug {
			log.Printf(
				"DEBUG: %s failed but did not hit minimum failures. "+
					"Count: %v alert after: %v",
				monitor.Name,
				monitor.failureCount,
				monitor.getAlertAfter(),
			)
		}
		return
	}

	// Take number of failures after minimum
	failureCount := (monitor.failureCount - monitor.getAlertAfter())

	// Use alert cadence to determine if we should alert
	if monitor.AlertEvery > 0 {
		// Handle integer number of failures before alerting
		if failureCount%monitor.AlertEvery == 0 {
			notice = monitor.createAlertNotice(false)
		}
	} else if monitor.AlertEvery == 0 {
		// Handle alerting on first failure only
		if failureCount == 0 {
			notice = monitor.createAlertNotice(false)
		}
	} else {
		// Handle negative numbers indicating an exponential backoff
		if failureCount >= int16(math.Pow(2, float64(monitor.alertCount))-1) {
			notice = monitor.createAlertNotice(false)
		}
	}

	// If we're going to alert, increment count
	if notice != nil {
		monitor.alertCount++
	}

	return
}

func (monitor Monitor) getAlertAfter() int16 {
	// TODO: Come up with a better way than this method
	// Zero is one!
	if monitor.AlertAfter == 0 {
		return 1
	}
	return monitor.AlertAfter
}

// GetAlertNames gives a list of alert names for a given monitor status
func (monitor Monitor) GetAlertNames(up bool) []string {
	if up {
		return monitor.AlertUp
	}
	return monitor.AlertDown
}

func (monitor Monitor) createAlertNotice(isUp bool) *AlertNotice {
	// TODO: Maybe add something about recovery status here
	return &AlertNotice{
		MonitorName:     monitor.Name,
		AlertCount:      monitor.alertCount,
		FailureCount:    monitor.failureCount,
		LastCheckOutput: monitor.lastOutput,
		LastSuccess:     monitor.lastSuccess,
		IsUp:            isUp,
	}
}
