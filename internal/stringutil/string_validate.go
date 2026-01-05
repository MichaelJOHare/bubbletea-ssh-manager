package stringutil

import (
	"errors"
	"strings"

	"bubbletea-ssh-manager/internal/config"
)

// ValidateHostNickname checks if the given nickname is valid.
func ValidateHostNickname(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("nickname is required")
	}
	if strings.ContainsAny(s, "*?!") {
		return errors.New("nicknames with wildcard characters are not supported")
	}
	if strings.Contains(s, ".") {
		return errors.New("nicknames cannot contain '.'")
	}
	return nil
}

// ValidateHostGroup checks if the given group name is valid.
//
// It throws an error if the group name contains wildcard characters or a dot.
func ValidateHostGroup(s string) error {
	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, "*?!") {
		return errors.New("group names with wildcard characters are not supported")
	}
	if strings.Contains(s, ".") {
		return errors.New("group names cannot contain '.'")
	}
	return nil
}

// ValidateHostName checks if the given hostname is valid for the specified protocol.
func ValidateHostName(protocol config.Protocol, s string) error {
	s = strings.TrimSpace(s)
	if protocol == config.ProtocolTelnet && s == "" {
		return errors.New("hostname is required for telnet")
	}
	return nil
}

// ValidateHostPort checks if the given port is valid for the specified protocol.
func ValidateHostPort(protocol config.Protocol, s string) error {
	_, err := NormalizePort(s, protocol)
	return err
}
