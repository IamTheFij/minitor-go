check_interval = "1s"

alert "log_command" {
  command = "should be a list"
}

monitor "Command" {
  command =  ["echo", "$PATH"]
  alert_down =  ["log_command"]
  alert_every =  2
  check_interval =  "10s"
}
