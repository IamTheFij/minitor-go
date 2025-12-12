package main

import (
	"errors"
	"fmt"
	"math"
	"os/exec"
	"time"

	"git.iamthefij.com/iamthefij/slog"
)

// Monitor represents a particular periodic check of a command
type Monitor struct { //nolint:maligned
	// Config values
	CheckIntervalStr *string `hcl:"check_interval,optional"`
	CheckInterval    time.Duration

	Name         string `hcl:"name,label"`
	AlertCount   int
	AlertAfter   int      `hcl:"alert_after,optional"`
	AlertEvery   *int     `hcl:"alert_every,optional"`
	AlertDown    []string `hcl:"alert_down,optional"`
	AlertUp      []string `hcl:"alert_up,optional"`
	Command      []string `hcl:"command,optional"`
	ShellCommand string   `hcl:"shell_command,optional"`

	// Other values
	failureCount      int
	lastCheck         time.Time
	lastSuccess       time.Time
	lastOutput        string
	lastCheckDuration time.Duration
}

// Init initializes the Monitor with default values
func (monitor *Monitor) Init(defaultAlertAfter int, defaultAlertEvery *int, defaultAlertDown []string, defaultAlertUp []string) error {
	// Parse the check_interval string into a time.Duration
	if monitor.CheckIntervalStr != nil {
		var err error

		monitor.CheckInterval, err = time.ParseDuration(*monitor.CheckIntervalStr)
		if err != nil {
			return fmt.Errorf("failed to parse check_interval duration for monitor %s: %w", monitor.Name, err)
		}
	}

	// Set default values for monitor alerts
	if monitor.AlertAfter == 0 {
		minAlertAfter := 1
		monitor.AlertAfter = max(defaultAlertAfter, minAlertAfter)
	}

	if monitor.AlertEvery == nil {
		monitor.AlertEvery = defaultAlertEvery
	}

	if len(monitor.AlertDown) == 0 {
		monitor.AlertDown = defaultAlertDown
	}

	if len(monitor.AlertUp) == 0 {
		monitor.AlertUp = defaultAlertUp
	}

	return nil
}

// Validate checks that the Monitor is properly configured and returns errors if not
func (monitor Monitor) Validate() error {
	hasCommand := len(monitor.Command) > 0
	hasShellCommand := monitor.ShellCommand != ""
	hasValidAlertAfter := monitor.AlertAfter > 0
	hasAlertDown := len(monitor.AlertDown) > 0

	var err error

	hasAtLeastOneCommand := hasCommand || hasShellCommand
	if !hasAtLeastOneCommand {
		err = errors.Join(err, fmt.Errorf(
			"%w: monitor %s has no command or shell_command configured",
			ErrInvalidMonitor,
			monitor.Name,
		))
	}

	hasAtMostOneCommand := !(hasCommand && hasShellCommand)
	if !hasAtMostOneCommand {
		err = errors.Join(err, fmt.Errorf(
			"%w: monitor %s has both command and shell_command configured",
			ErrInvalidMonitor,
			monitor.Name,
		))
	}

	if !hasValidAlertAfter {
		err = errors.Join(err, fmt.Errorf(
			"%w: monitor %s has invalid alert_after value %d. Must be greater than 0",
			ErrInvalidMonitor,
			monitor.Name,
			monitor.AlertAfter,
		))
	}

	if !hasAlertDown {
		err = errors.Join(err, fmt.Errorf(
			"%w: monitor %s has no alert_down configured. Configure one here or add a default_alert_down",
			ErrInvalidMonitor,
			monitor.Name,
		))
	}

	return err
}

func (monitor Monitor) LastOutput() string {
	return monitor.lastOutput
}

// ShouldCheck returns a boolean indicating if the Monitor is ready to be be checked again
func (monitor Monitor) ShouldCheck() bool {
	if monitor.lastCheck.IsZero() || monitor.CheckInterval == 0 {
		return true
	}

	sinceLastCheck := time.Since(monitor.lastCheck)

	return sinceLastCheck >= monitor.CheckInterval
}

// Check will run the command configured by the Monitor and return a status and a possible AlertNotice
func (monitor *Monitor) Check() (bool, *AlertNotice) {
	var cmd *exec.Cmd
	if len(monitor.Command) > 0 {
		cmd = exec.Command(monitor.Command[0], monitor.Command[1:]...)
	} else if monitor.ShellCommand != "" {
		cmd = ShellCommand(monitor.ShellCommand)
	} else {
		slog.Fatalf("Monitor %s has no command configured", monitor.Name)
	}

	checkStartTime := time.Now()
	output, err := cmd.CombinedOutput()
	monitor.lastCheck = time.Now()
	monitor.lastOutput = string(output)
	monitor.lastCheckDuration = monitor.lastCheck.Sub(checkStartTime)

	var alertNotice *AlertNotice

	isSuccess := (err == nil)
	if isSuccess {
		alertNotice = monitor.Success()
	} else {
		alertNotice = monitor.Failure()
	}

	slog.Debugf("Command output: %s", monitor.lastOutput)
	slog.OnErrWarnf(err, "Command result: %v", err)

	slog.Infof(
		"%s success=%t, alert=%t",
		monitor.Name,
		isSuccess,
		alertNotice != nil,
	)

	return isSuccess, alertNotice
}

// GetAlertNames gives a list of alert names for a given monitor status
func (monitor Monitor) GetAlertNames(up bool) []string {
	if up {
		return monitor.AlertUp
	}

	return monitor.AlertDown
}

// IsUp returns the status of the current monitor
func (monitor Monitor) IsUp() bool {
	return monitor.AlertCount == 0
}

// LastCheckMilliseconds gives number of miliseconds the last check ran for
func (monitor Monitor) LastCheckMilliseconds() int64 {
	return monitor.lastCheckDuration.Milliseconds()
}

func (monitor *Monitor) Success() (notice *AlertNotice) {
	if !monitor.IsUp() {
		// Alert that we have recovered
		notice = monitor.createAlertNotice(true)
	}

	monitor.failureCount = 0
	monitor.AlertCount = 0
	monitor.lastSuccess = time.Now()

	return
}

func (monitor *Monitor) Failure() (notice *AlertNotice) {
	monitor.failureCount++
	// If we haven't hit the minimum failures, we can exit
	if monitor.failureCount < monitor.AlertAfter {
		slog.Debugf(
			"%s failed but did not hit minimum failures. "+
				"Count: %v alert after: %v",
			monitor.Name,
			monitor.failureCount,
			monitor.AlertAfter,
		)

		return
	}

	// Take number of failures after minimum
	failureCount := (monitor.failureCount - monitor.AlertAfter)

	// Use alert cadence to determine if we should alert
	switch {
	case monitor.AlertEvery == nil, *monitor.AlertEvery == 0:
		// Handle alerting on first failure only
		if failureCount == 0 {
			notice = monitor.createAlertNotice(false)
		}
	case *monitor.AlertEvery > 0:
		// Handle integer number of failures before alerting
		if failureCount%*monitor.AlertEvery == 0 {
			notice = monitor.createAlertNotice(false)
		}
	default:
		// Handle negative numbers indicating an exponential backoff
		if failureCount >= int(math.Pow(2, float64(monitor.AlertCount))-1) { //nolint:mnd
			notice = monitor.createAlertNotice(false)
		}
	}

	// If we're going to alert, increment count
	if notice != nil {
		monitor.AlertCount++
		notice.AlertCount = monitor.AlertCount
	}

	return notice
}

func (monitor Monitor) createAlertNotice(isUp bool) *AlertNotice {
	// TODO: Maybe add something about recovery status here
	return &AlertNotice{
		MonitorName:     monitor.Name,
		AlertCount:      monitor.AlertCount,
		FailureCount:    monitor.failureCount,
		LastCheckOutput: monitor.lastOutput,
		LastSuccess:     monitor.lastSuccess,
		IsUp:            isUp,
	}
}
