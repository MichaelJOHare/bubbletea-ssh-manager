package connect

import (
	"net"
	"strings"
	"time"
)

// ShouldPreflight returns true if the given Target requires a reachability check.
//
// Telnet always requires preflight (host/port).
// SSH requires preflight if a hostname is set (host/port for display / checks).
func ShouldPreflight(t Target) bool {
	switch strings.TrimSpace(t.protocol) {
	case "telnet":
		return true
	case "ssh":
		return strings.TrimSpace(t.HostName) != ""
	default:
		return false
	}
}

// HostPortForPreflight returns the host or host:port string used for preflight.
func HostPortForPreflight(t Target) string {
	host := strings.TrimSpace(t.HostName)
	port := strings.TrimSpace(t.Port)
	if host == "" {
		return ""
	}
	if port == "" {
		return host
	}
	return net.JoinHostPort(host, port)
}

// PreflightDial attempts to open a TCP connection to hostPort within timeout.
func PreflightDial(hostPort string, timeout time.Duration) error {
	hostPort = strings.TrimSpace(hostPort)
	if hostPort == "" {
		return nil
	}

	d := net.Dialer{Timeout: timeout}
	c, err := d.Dial("tcp", hostPort)
	if c != nil {
		_ = c.Close()
	}
	return err
}
