package main

import (
	"fmt"
	"strconv"
	"strings"
)

// normalizeString returns a normalized string (trimmed, lowercased).
func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// normalizePort returns a validated numeric port for a protocol.
//
// Behavior:
//   - If port is empty, returns the default port for the protocol (ssh=22, telnet=23).
//   - If port is numeric, validates it's within 1..65535 and returns it.
func normalizePort(port string, protocol string) (string, error) {
	port = strings.TrimSpace(port)
	protocol = normalizeString(protocol)

	if port == "" {
		switch protocol {
		case "ssh":
			return "22", nil
		case "telnet":
			return "23", nil
		}
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid %s port %q", protocol, port)
	}
	if n < 1 || n > 65535 {
		return "", fmt.Errorf("port out of range: %d", n)
	}
	return strconv.Itoa(n), nil
}
