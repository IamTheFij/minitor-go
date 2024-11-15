check_interval = "1s"

monitor "Shell" {
  shell_command = <<-EOF
    echo 'Some string with stuff'
    echo "<angle brackets>"
    exit 1
  EOF
  alert_down = ["log_shell"]
  alert_after =  1
  alert_every =  0
}

alert "log_shell" {
  shell_command = <<EOF
    echo 'Some string with stuff'
    echo '<angle brackets>'
EOF
}
