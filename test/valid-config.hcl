check_interval = "1s"

alert "log_command" {
  command = ["echo", "regular", "'command!!!'", "{{.MonitorName}}"]
}

alert "log_shell" {
  shell_command = "echo \"Failure on {{.MonitorName}} User is $USER\""
}

monitor "Default" {
  command = ["echo"]
  alert_down = ["log_command"]
}

monitor "Command" {
  command =  ["echo", "$PATH"]
  alert_down =  ["log_command", "log_shell"]
  alert_every =  2
  check_interval =  "10s"
}

monitor "Shell" {
  shell_command = <<-EOF
    echo 'Some string with stuff'
    echo 'another line'
    echo $PATH
    exit 1
  EOF
  alert_down = ["log_command", "log_shell"]
  alert_after = 5
  alert_every = 0
  check_interval = "1m"
}
