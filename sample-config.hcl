check_interval = "5s"

monitor "Fake Website" {
  command = ["curl", "-s", "-o", "/dev/null", "https://minitor.mon"]
  alert_down = ["log_down", "mailgun_down", "sms_down"]
  alert_up = ["log_up", "email_up"]
  check_interval = "10s" # Must be at minimum the global `check_interval`
  alert_after = 3
  alert_every = -1 # Defaults to -1 for exponential backoff. 0 to disable repeating
}

monitor "Real Website" {
  command = ["curl", "-s", "-o", "/dev/null", "https://google.com"]
  alert_down = ["log_down", "mailgun_down", "sms_down"]
  alert_up = ["log_up", "email_up"]
  check_interval = "5s"
  alert_after = 3
  alert_every = -1
}

alert "log_down" {
    command = ["echo", "Minitor failure for {{.MonitorName}}"]
  }
alert "log_up" {
    command = ["echo", "Minitor recovery for {{.MonitorName}}"]
  }
alert "email_up" {
    command = ["sendmail", "me@minitor.mon", "Recovered: {monitor_name}", "We're back!"]
  }
alert "mailgun_down" {
    shell_command = <<EOF
      curl -s -X POST
      -F subject="Alert! {{.MonitorName}} failed"
      -F from="Minitor <minitor@minitor.mon>"
      -F to=me@minitor.mon
      -F text="Our monitor failed"
      https://api.mailgun.net/v3/minitor.mon/messages
      -u "api:${MAILGUN_API_KEY}"
      EOF
    }
alert "sms_down" {
    shell_command = <<EOF
      curl -s -X POST -F "Body=Failure! {{.MonitorName}} has failed"
      -F "From=${AVAILABLE_NUMBER}" -F "To=${MY_PHONE}"
      "https://api.twilio.com/2010-04-01/Accounts/${ACCOUNT_SID}/Messages"
      -u "${ACCOUNT_SID}:${AUTH_TOKEN}"
      EOF
 }
