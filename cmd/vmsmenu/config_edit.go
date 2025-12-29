package main

import (
	"bubbletea-ssh-manager/internal/config"
	"bubbletea-ssh-manager/internal/host"
	"bubbletea-ssh-manager/internal/sshopts"
	"errors"
	"fmt"
	"os"
	"strings"
)

// getProtocolConfigPath returns the root config file for the given protocol.
//
//   - ssh: ~/.ssh/config
//   - telnet: ~/.telnet/config
func getProtocolConfigPath(protocol string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "ssh":
		return config.GetConfigPath(".ssh", "config")
	case "telnet":
		return config.GetConfigPath(".telnet", "config")
	default:
		return "", fmt.Errorf("unknown protocol: %q", protocol)
	}
}

// getConfigPathForAlias returns the config file path that should be edited
// for the given alias.
//
// If the alias was defined in an included file, this returns that included file.
// If not found, it returns ("", nil).
func getConfigPathForAlias(protocol, alias string) (string, error) {
	root, err := getProtocolConfigPath(protocol)
	if err != nil {
		return "", err
	}
	entry, err := config.FindHostEntry(root, alias)
	if err != nil || entry == nil {
		return "", err
	}
	source := entry.SourcePath
	return source, nil
}

// AddHostToRootConfig appends a new host block to the protocol's root config.
//
// This intentionally writes to the root config even if it contains Include
// directives; may be extended in the future to support writing to included files.
func AddHostToRootConfig(protocol string, spec host.Spec, opts sshopts.Options) error {
	root, err := getProtocolConfigPath(protocol)
	if err != nil {
		return err
	}
	return config.AddHostEntry(root, config.EntryFromSpec(spec, opts, root))
}

// UpdateHostInConfig updates an existing host entry.
//
// It uses Include-aware resolution so edits land in the file that originally
// defined oldAlias.
func UpdateHostInConfig(protocol, oldAlias string, updated host.Spec, opts sshopts.Options) error {
	configPath, err := getConfigPathForAlias(protocol, oldAlias)
	if err != nil {
		return err
	}
	if strings.TrimSpace(configPath) == "" {
		return os.ErrNotExist
	}
	return config.UpdateHostEntry(configPath, oldAlias, config.EntryFromSpec(updated, opts, configPath))
}

// RemoveHostFromConfig removes an alias from the config file that defined it.
//
// It uses Include-aware resolution so removals land in the correct include file.
func RemoveHostFromConfig(protocol, alias string) error {
	configPath, err := getConfigPathForAlias(protocol, alias)
	if err != nil {
		return err
	}
	if strings.TrimSpace(configPath) == "" {
		return os.ErrNotExist
	}
	err = config.RemoveHostEntry(configPath, alias)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	return err
}
