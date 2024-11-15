check_interval = "1s"

monitor "Command" {
    command = ["echo", "$PATH"]
    alert_down = [ "alert_down", "log_shell", "log_command" ]
    alert_every = 0
}
