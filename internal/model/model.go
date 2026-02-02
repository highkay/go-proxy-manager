package model

import (
	"time"
)

// Proxy represents a proxy server instance
type Proxy struct {
	URL        string        `json:"url"`         // e.g., http://1.2.3.4:8080
	Protocol   string        `json:"protocol"`    // http, https, socks5
	IP         string        `json:"ip"`
	Port       int           `json:"port"`
	Latency    time.Duration `json:"latency"`
	LastCheck  time.Time     `json:"last_check"`
	FailCount  int           `json:"fail_count"`
	Source     string        `json:"source"`
}

// Source represents a proxy provider
type Source struct {
	URL      string        `yaml:"url"`
	Type     string        `yaml:"type"`     // text, json
	Interval time.Duration `yaml:"interval"` // Update interval
}
