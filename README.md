# [minitor-go](https://git.iamthefij.com/iamthefij/minitor-go)

A minimal monitoring system

## What does it do?

Minitor accepts an HCL configuration file with a set of commands to run and a set of alerts to execute when those commands fail. Minitor has a narow feature set and instead follows a principle to outsource to other command line tools when possible. Thus, it relies on other command line tools to do checks and issue alerts. To make getting started a bit easier, Minitor includes a few scripts to help with common tasks.

## But why?

I'm running a few small services and found Sensu, Consul, Nagios, etc. to all be far too complicated for my usecase.

## So how do I use it?

### Running

Install and execute with:

```bash
go install github.com/iamthefij/minitor-go@latest
minitor
```

If locally developing you can use:

```bash
make run
```

It will read the contents of `sample-config.hcl` and begin its loop. You could also run it directly and provide a new config file via the `-config` argument.


#### Docker

You can pull this repository directly from Docker:

```bash
docker pull iamthefij/minitor-go:latest
```

The Docker image uses a default `config.hcl` copied from `sample-config.hcl`. This won't really do anything for you, so when you run the Docker image, you should supply your own `config.hcl` file:

```bash
docker run -v $PWD/sample-config.hcl:/app/config.hcl iamthefij/minitor-go:latest
```

Images are provided for `amd64`, `arm`, and `arm64` architechtures.

You can configure the timezone for the container by passing a `TZ` env variable. Eg. `TZ=America/Los_Angeles`.

## Configuring

In this repo, you can explore the `sample-config.hcl` file for an example, but the general structure is as follows. It should be noted that environment variable interpolation happens on load of the HCL file.

The global configurations are:

|key|value|
|---|---|
|`check_interval`|Maximum frequency to run checks for each monitor as duration, eg. 1m2s.|
|`default_alert_after`|A default value used as an `alert_after` value for a monitor if not specified or 0.|
|`default_alert_every`|A default value used as an `alert_every` value for a monitor if not specified.|
|`default_alert_down`|Default down alerts to used by a monitor in case none are provided.|
|`default_alert_up`|Default up alerts to used by a monitor in case none are provided.|
|`monitor`|block listing monitors. Detailed description below|
|`alert`|List of all alerts. Detailed description below|

### Monitors

Represent your monitors as blocks with a label indicating the name of the monitor.

```hcl
monitor "example" {
  command = ["echo", "Hello, World!"]
  alert_down = ["log"]
  alert_up = ["log"]
  check_interval = "1m"
  alert_after = 1
  alert_every = -1
}
```

Each monitor allows the following configuration:

|key|value|
|---|---|
|`name`|Name of the monitor running. This will show up in messages and logs.|
|`command`|A list of strings representing a command to be executed. This command's exit value will determine whether the check is successful. This value is mutually exclusive to `shell_command`|
|`shell_command`|A single string that represents a shell command to be executed. This command's exit value will determine whether the check is successful. This value is mutually exclusive to `command`|
|`alert_down`|A list of Alerts to be triggered when the monitor is in a "down" state|
|`alert_up`|A list of Alerts to be triggered when the monitor moves to an "up" state|
|`check_interval`|The interval at which this monitor should be checked. This must be greater than the global `check_interval` value|
|`alert_after`|Allows specifying the number of failed checks before an alert should be triggered. A value of 1 will start sending alerts after the first failure.|
|`alert_every`|Allows specifying how often an alert should be retriggered. There are a few magic numbers here. Defaults to `-1` for an exponential backoff. Setting to `0` disables re-alerting. Positive values will allow retriggering after the specified number of checks|

### Alerts

Represent your alerts as blocks with a lable indicating the name of the alert. The name will be used in your monitor setup in `alert_down` and `alert_up`.

```hcl
monitor "example" {
  command = ["false"]
  alert_down = ["log"]
}

alert "log" {
  shell_command = "echo '{{.MonitorName}} is down!'"
}
```

Each alert allows the following configuration:

|key|value|
|---|---|
|`command`|Specifies the command that should be executed in exec form. This is the command that will be run when the alert is executed. This can be templated with environment variables or the variables shown in the table below. This value is mutually exclusive to `shell_command`|
|`shell_command`|Specifies a shell command as a single string. This is the command that will be run when the alert is executed. This can be templated with environment variables or the variables shown in the table below. This value is mutually exclusive to `command`|

Also, when alerts are executed, they will be passed through Go's format function with arguments for some attributes of the Monitor. The following monitor specific variables can be referenced using Go formatting syntax:

|token|value|
|---|---|
|`{{.AlertCount}}`|Number of times this monitor has alerted|
|`{{.FailureCount}}`|The total number of sequential failed checks for this monitor|
|`{{.LastCheckOutput}}`|The last returned value from the check command to either stderr or stdout|
|`{{.LastSuccess}}`|The datetime of the last successful check as a go Time struct|
|`{{.MonitorName}}`|The name of the monitor that failed and triggered the alert|
|`{{.IsUp}}`|Indicates if the monitor that is alerting is up or not. Can be used in a conditional message template|

To provide flexible formatting, the following non-standard functions are available in templates:

|func|description|
|---|---|
|`ANSIC <Time>`|Formats provided time in ANSIC format|
|`UnixDate <Time>`|Formats provided time in UnixDate format|
|`RubyDate <Time>`|Formats provided time in RubyDate format|
|`RFC822Z <Time>`|Formats provided time in RFC822Z format|
|`RFC850 <Time>`|Formats provided time in RFC850 format|
|`RFC1123 <Time>`|Formats provided time in RFC1123 format|
|`RFC1123Z <Time>`|Formats provided time in RFC1123Z format|
|`RFC3339 <Time>`|Formats provided time in RFC3339 format|
|`RFC3339Nano <Time>`|Formats provided time in RFC3339Nano format|
|`FormatTime <Time> <string template>`|Formats provided time according to provided template|
|`InTZ <Time> <string timezone name>`|Converts provided time to parsed timezone from the provided name|

For more information, check out the [Go documentation for the time module](https://pkg.go.dev/time@go1.20.7#pkg-constants).

#### Running alerts on startup

It's not the best feeling to find out your alerts are broken when you're expecting to be alerted about another failure. To avoid this and provide early insight into broken alerts, it is possible to specify a list of alerts to run when Minitor starts up. This can be done using the command line flag `-startup-alerts`. This flag accepts a comma separated list of strings and will run a test of each of those alerts. Minitor will then respond as it typically does for any failed alert. This can be used to allow you time to correct when initially launching, and to allow schedulers to more easily detect a failed deployment of Minitor.

Eg.

```bash
minitor -startup-alerts=log_down,log_up -config ./config.hcl
```

### Metrics

Minitor supports exporting metrics for [Prometheus](https://prometheus.io/). Prometheus is an open source tool for reading and querying metrics from different sources. Combined with another tool, [Grafana](https://grafana.com/), it allows building of charts and dashboards. You could also opt to just use Minitor to log check results, and instead do your alerting with Grafana.

It is also possible to use the metrics endpoint for monitoring Minitor itself! This allows setting up multiple instances of Minitor on different servers and have them monitor each-other so that you can detect a minitor outage.

To run minitor with metrics, use the `-metrics` flag. The metrics will be served on port `8080` by default, though it can be overriden using `-metrics-port`. They will be accessible on the path `/metrics`. Eg. `localhost:8080/metrics`.

```bash
minitor -metrics
# or
minitor -metrics -metrics-port 3000
```

## Contributing

Whether you're looking to submit a patch or tell me I broke something, you can contribute through the Github mirror and I can merge PRs back to the source repository.

Primary Repo: https://git.iamthefij.com/iamthefij/minitor.git

Github Mirror: https://github.com/IamTheFij/minitor.git

## Original Minitor

This is a reimplementation of [Minitor](https://git.iamthefij.com/iamthefij/minitor) in Go

Minitor is already a minimal monitoring tool. Python 3 was a quick way to get something live, but Python itself comes with a large footprint. Thus Go feels like a better fit for the project, longer term.

Initial target is meant to be roughly compatible requiring only minor changes to configuration. Future iterations may diverge to take advantage of Go specific features.
