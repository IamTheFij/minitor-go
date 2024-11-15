check_interval = "1s"
default_alert_down = ["log_command"]
default_alert_after = 1

monitor "Command" {
  command = ["echo", "$PATH"]
}

alert "log_command" {
  command = ["echo", "default", "'command!!!'", "{{.MonitorName}}"]
}
