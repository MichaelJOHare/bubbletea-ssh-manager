package connect

import (
	"net"
	"time"

	"bubbletea-ssh-manager/internal/config"
)

// ShouldPreflight returns true if the given Target requires a reachability check.
//
// Telnet always requires preflight (host/port).
// SSH requires preflight if a hostname is set (host/port for display & checks).
func ShouldPreflight(t Target) bool {
	switch t.Protocol {
	case config.ProtocolTelnet:
		return true
	case config.ProtocolSSH:
		return t.HostName != ""
	default:
		return false
	}
}

// GenerateHostPort returns the host or host:port string used for preflight.
func GenerateHostPort(t Target) string {
	host := t.HostName
	port := t.Port
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
