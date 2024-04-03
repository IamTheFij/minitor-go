package main

import (
	"testing"
)

func TestNewHealthCheck(t *testing.T) {
	monitors := []*Monitor{
		{Name: "Test Monitor"},
	}
	hc := NewHealthCheckHandler(monitors)

	monitors[0].alertCount++

	if healthy, _ := hc.MinitorHealthCheck(); healthy {
		t.Errorf("Initial hc state should be unhealthy until some successful alert is sent")
	}

	if healthy, _ := hc.MonitorsHealthCheck(); healthy {
		t.Errorf("Faking an alert on the monitor pointer should make this unhealthy")
	}
}

func TestMinitorHealthCheck(t *testing.T) {
	monitors := []*Monitor{
		{Name: "Test Monitor"},
	}
	hc := NewHealthCheckHandler(monitors)

	t.Run("MinitorHealthCheck(healthy)", func(t *testing.T) {
		hc.MinitorHealthy(true)
		healthy, body := hc.MinitorHealthCheck()
		if !healthy {
			t.Errorf("Expected healthy check")
		}
		if body != "OK" {
			t.Errorf("Expected OK response")
		}
	})

	t.Run("MinitorHealthCheck(unhealthy)", func(t *testing.T) {
		hc.MinitorHealthy(false)
		healthy, body := hc.MinitorHealthCheck()
		if healthy {
			t.Errorf("Expected healthy check")
		}
		if body != "UNHEALTHY" {
			t.Errorf("Expected UNHEALTHY response")
		}
	})
}

func TestMonitorsHealthCheck(t *testing.T) {
	monitors := []*Monitor{
		{Name: "Test Monitor"},
	}
	hc := NewHealthCheckHandler(monitors)

	t.Run("MonitorsHealthCheck(healthy)", func(t *testing.T) {
		healthy, body := hc.MonitorsHealthCheck()
		if !healthy {
			t.Errorf("Expected healthy check")
		}
		if body != "OK" {
			t.Errorf("Expected OK response")
		}
	})

	t.Run("MonitorsHealthCheck(unhealthy)", func(t *testing.T) {
		monitors[0].alertCount++
		healthy, body := hc.MonitorsHealthCheck()
		if healthy {
			t.Errorf("Expected healthy check")
		}
		if body != "UNHEALTHY: The following monitors are unhealthy: Test Monitor" {
			t.Errorf("Expected UNHEALTHY response")
		}
	})
}
