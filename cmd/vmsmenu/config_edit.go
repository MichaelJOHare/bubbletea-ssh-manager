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

// protocolRootConfigPath returns the root config file for the given protocol.
//
// - ssh: ~/.ssh/config
// - telnet: ~/.telnet/config
func protocolRootConfigPath(protocol string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "ssh":
		return config.UserConfigPath(".ssh", "config")
	case "telnet":
		return config.UserConfigPath(".telnet", "config")
	default:
		return "", fmt.Errorf("unknown protocol: %q", protocol)
	}
}

// resolveConfigPathForAlias returns the config file path that should be edited
// for the given alias.
//
// If the alias was defined in an included file, this returns that included file.
// If not found, it returns ("", nil).
func resolveConfigPathForAlias(protocol, alias string) (string, error) {
	root, err := protocolRootConfigPath(protocol)
	if err != nil {
		return "", err
	}
	source, err := config.ResolveHostSourcePath(root, alias)
	if err != nil {
		return "", err
	}
	return source, nil
}

func entryFromSpec(spec host.Spec, opts sshopts.Options, sourcePath string) config.HostEntry {
	return config.HostEntry{
		Spec:       spec,
		SSHOptions: opts,
		SourcePath: sourcePath,
	}
}

// AddHostToRootConfig appends a new host block to the protocol's root config.
//
// This intentionally writes to the root config even if it contains Include
// directives; the TUI can later be extended to let users pick a target include.
func AddHostToRootConfig(protocol string, spec host.Spec, opts sshopts.Options) error {
	root, err := protocolRootConfigPath(protocol)
	if err != nil {
		return err
	}
	return config.AddHostEntry(root, entryFromSpec(spec, opts, root))
}

// UpdateHostInConfig updates an existing host entry.
//
// It uses Include-aware resolution so edits land in the file that originally
// defined oldAlias.
func UpdateHostInConfig(protocol, oldAlias string, updated host.Spec, opts sshopts.Options) error {
	configPath, err := resolveConfigPathForAlias(protocol, oldAlias)
	if err != nil {
		return err
	}
	if strings.TrimSpace(configPath) == "" {
		return os.ErrNotExist
	}
	return config.UpdateHostEntry(configPath, oldAlias, entryFromSpec(updated, opts, configPath))
}

// RemoveHostFromConfig removes an alias from the config file that defined it.
//
// It uses Include-aware resolution so removals land in the correct include file.
func RemoveHostFromConfig(protocol, alias string) error {
	configPath, err := resolveConfigPathForAlias(protocol, alias)
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
