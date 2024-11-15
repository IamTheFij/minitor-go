check_interval = "1s"

monitor "Command" {
  command = ["echo", "$PATH"]
  alert_down = ["not_log"]
  alert_every = 0
}


alert "log" {
  command = ["true"]
}
