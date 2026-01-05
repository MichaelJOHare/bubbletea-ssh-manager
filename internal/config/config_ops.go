package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetConfigPath returns the full path to a config file under the user's home directory.
//
// On Windows/MSYS2, prefer $HOME so this matches where MSYS2/OpenSSH tools
// look for config files (eg. ~/.ssh/config).
func GetConfigPath(parts ...string) (string, error) {
	home, err := getHomeDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, parts...)...), nil
}

// GetConfigPathForProtocol returns the root config file for the given protocol.
//
//   - ssh: ~/.ssh/config
//   - telnet: ~/.telnet/config
func GetConfigPathForProtocol(protocol Protocol) (string, error) {
	switch protocol {
	case ProtocolSSH:
		return GetConfigPath(".ssh", "config")
	case ProtocolTelnet:
		return GetConfigPath(".telnet", "config")
	default:
		return "", fmt.Errorf("unknown protocol: %q", protocol)
	}
}

// GetConfigPathForAlias returns the config file path that should be edited
// for the given alias.
//
// If the alias was defined in an included file, this returns that included file.
// If not found, it returns ("", nil).
func GetConfigPathForAlias(protocol Protocol, alias string) (string, error) {
	root, err := GetConfigPathForProtocol(protocol)
	if err != nil {
		return "", err
	}
	entry, err := FindHostEntry(root, alias)
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
func AddHostToRootConfig(protocol Protocol, spec Spec, opts SSHOptions) error {
	root, err := GetConfigPathForProtocol(protocol)
	if err != nil {
		return err
	}
	return AddHostEntry(root, EntryFromSpec(spec, opts, root))
}

// UpdateHostInConfig updates an existing host entry.
//
// It uses Include-aware resolution so edits land in the file that originally
// defined oldAlias.
func UpdateHostInConfig(protocol Protocol, oldAlias string, updated Spec, opts SSHOptions) error {
	configPath, err := GetConfigPathForAlias(protocol, oldAlias)
	if err != nil {
		return err
	}
	if strings.TrimSpace(configPath) == "" {
		return os.ErrNotExist
	}
	return UpdateHostEntry(configPath, oldAlias, EntryFromSpec(updated, opts, configPath))
}

// RemoveHostFromConfig removes an alias from the config file that defined it.
//
// It uses Include-aware resolution so removals land in the correct include file.
func RemoveHostFromConfig(protocol Protocol, alias string) error {
	configPath, err := GetConfigPathForAlias(protocol, alias)
	if err != nil {
		return err
	}
	if strings.TrimSpace(configPath) == "" {
		return os.ErrNotExist
	}
	err = RemoveHostEntry(configPath, alias)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	return err
}
