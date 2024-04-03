package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HealthCheckHandler struct {
	isMinitorHealthy bool
	monitors         []*Monitor
}

func NewHealthCheckHandler(monitors []*Monitor) *HealthCheckHandler {
	return &HealthCheckHandler{
		false,
		monitors,
	}
}

func (hch *HealthCheckHandler) MinitorHealthy(healthy bool) {
	hch.isMinitorHealthy = healthy
}

func (hch HealthCheckHandler) MinitorHealthCheck() (bool, string) {
	if hch.isMinitorHealthy {
		return true, "OK"
	} else {
		return false, "UNHEALTHY"
	}
}

func (hch HealthCheckHandler) MonitorsHealthCheck() (bool, string) {
	downMonitors := []string{}

	for _, monitor := range hch.monitors {
		if !monitor.IsUp() {
			downMonitors = append(downMonitors, monitor.Name)
		}
	}

	if len(downMonitors) == 0 {
		return true, "OK"
	} else {
		return false, fmt.Sprintf("UNHEALTHY: The following monitors are unhealthy: %s", strings.Join(downMonitors, ", "))
	}
}

func (hch HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var healthy bool

	var body string

	if monitors := r.URL.Query().Get("monitors"); monitors != "" {
		healthy, body = hch.MonitorsHealthCheck()
	} else {
		healthy, body = hch.MinitorHealthCheck()
	}

	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_, _ = io.WriteString(w, body)
}

func HandleHealthCheck() {
	http.Handle("/metrics", HealthChecks)
}
