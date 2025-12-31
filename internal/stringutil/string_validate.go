package stringutil

import (
	"errors"
	"strings"
)

// ValidateHostNickname checks if the given nickname is valid.
func ValidateHostNickname(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("nickname is required")
	}
	if strings.ContainsAny(s, "*?!") {
		return errors.New("nickname patterns are not supported")
	}
	if strings.Contains(s, ".") {
		return errors.New("nickname cannot contain '.'")
	}
	return nil
}

// ValidateHostGroup checks if the given group name is valid.
func ValidateHostGroup(s string) error {
	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, "*?!") {
		return errors.New("group patterns are not supported")
	}
	if strings.Contains(s, ".") {
		return errors.New("group cannot contain '.'")
	}
	return nil
}

// ValidateHostName checks if the given hostname is valid for the specified protocol.
func ValidateHostName(protocol string, s string) error {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	s = strings.TrimSpace(s)
	if protocol == "telnet" && s == "" {
		return errors.New("hostname is required for telnet")
	}
	return nil
}

// ValidateHostPort checks if the given port is valid for the specified protocol.
func ValidateHostPort(protocol string, s string) error {
	_, err := NormalizePort(s, protocol)
	return err
}
