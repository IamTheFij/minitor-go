# minitor-go

A reimplementation of [Minitor](https://git.iamthefij.com/iamthefij/minitor) in Go

Minitor is already a minimal monitoring tool. Python 3 was a quick way to get something live, but Python itself comes with a large footprint. Thus Go feels like a better fit for the project, longer term.

Initial target is meant to be roughly compatible requiring only minor changes to configuration. Future iterations may diverge to take advantage of Go specific features.

## Differences from Python version


Templating for Alert messages has been updated. In the Python version, `str.format(...)` was used with certain keys passed in that could be used to format messages. In the Go version, we use a struct, `AlertNotice` defined in `alert.go` and the built in Go templating format. Eg.

minitor-py:
```yaml
alerts:
  log_command:
    command: ['echo', '{monitor_name}']
  log_shell:
    command: 'echo {monitor_name}'
```

minitor-go:
```yaml
alerts:
  log_command:
    command: ['echo', '{{.MonitorName}}']
  log_shell:
    command: 'echo {{.MonitorName}}'
```

Finally, newlines in a shell command don't terminate a particular command. Semicolons must be used and continuations should not.

minitor-py:
```yaml
alerts:
  log_shell:
    command: >
      echo "line 1"
      echo "line 2"
      echo "continued" \
        "line"
```

minitor-go:
```yaml
alerts:
  log_shell:
    command: >
      echo "line 1";
      echo "line 2";
      echo "continued"
        "line"
```

## To do
There are two sets of task lists. The first is to get rough parity on key features with the Python version. The second is to make some improvements to the framework.

Pairity:

  - [x] Run monitor commands
  - [x] Run monitor commands in a shell
  - [x] Run alert commands
  - [x] Run alert commands in a shell
  - [x] Allow templating of alert commands
  - [x] Implement Prometheus client to export metrics
  - [x] Test coverage
  - [x] Integration testing (manual or otherwise)
  - [x] Allow commands and shell commands in the same config key

Improvement (potentially breaking):

  - [ ] Implement leveled logging (maybe glog or logrus)
  - [ ] Consider switching from YAML to TOML
  - [ ] Consider value of templating vs injecting values into Env variables
  - [ ] Consider dropping `alert_up` and `alert_down` in favor of using Go templates that offer more control of messaging
  - [ ] Async checking
  - [ ] Use durations rather than seconds checked in event loop
  - [ ] Revisit metrics and see if they all make sense
