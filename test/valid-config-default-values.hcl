check_interval = "1s"
default_alert_down = ["log_command"]
default_alert_every = 0
default_alert_after = 2

monitor "Default" {
  command = ["echo"]
}

monitor "Command" {
  command = ["echo", "$PATH"]
}

alert "log_command" {
  command = ["echo", "default", "'command!!!'", "{{.MonitorName}}"]
}
