# minitor-go

A reimplementation of [Minitor](https://git.iamthefij.com/iamthefij/minitor) in Go

Minitor is already a minimal monitoring tool. Python 3 was a quick way to get something live, but Python itself comes with a large footprint. Thus Go feels like a better fit for the project, longer term.

Initial target is meant to be roughly compatible requiring only minor changes to configuration. Future iterations may diverge to take advantage of Go specific features.

## Differences from Python version

Templating for Alert messages has been updated. In the Python version, `str.format(...)` was used with certain keys passed in that could be used to format messages. In the Go version, we use a struct, `AlertNotice` defined in `alert.go` and the built in Go templating format. Eg.

minitor-py:
```yaml
alerts:
  log:
    command: 'echo {monitor_name}'
```

minitor-go:
```yaml
alerts:
  log:
    command: 'echo {{.MonitorName}}'
```

For the time being, legacy configs for the Python version of Minitor should be compatible if you apply the `-py-compat` flag when running Minitor. Eventually, this flag will go away when later breaking changes are introduced.

## Future

Future, potentially breaking changes

  - [ ] Implement leveled logging (maybe glog or logrus)
  - [ ] Consider value of templating vs injecting values into Env variables
  - [ ] Async checking
  - [ ] Revisit metrics and see if they all make sense
  - [ ] Consider dropping `alert_up` and `alert_down` in favor of using Go templates that offer more control of messaging (Breaking)
  - [ ] Use durations rather than seconds checked in event loop (Potentially breaking)
