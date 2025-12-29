package connect

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"bubbletea-ssh-manager/internal/host"
)

// BuildCommand builds the exec.Cmd to connect to the given Target.
//
// It returns a Target for display/title, and a TailBuffer that captures the last
// part of the command output for error reporting.
func BuildCommand(trgt Target) (cmd *exec.Cmd, tgt Target, tail *TailBuffer, err error) {
	protocol := strings.ToLower(strings.TrimSpace(trgt.Protocol))
	if protocol != "ssh" && protocol != "telnet" {
		name := strings.TrimSpace(trgt.Alias)
		return nil, Target{}, nil, fmt.Errorf("unknown protocol for %s: %q", name, trgt.Protocol)
	}

	programPath, err := PreferredProgramPath(protocol)
	if err != nil {
		return nil, Target{}, nil, fmt.Errorf("%s not found: %w", protocol, err)
	}

	alias := strings.TrimSpace(trgt.Alias)
	user := strings.TrimSpace(trgt.User)
	hostName := strings.TrimSpace(trgt.HostName)
	portRaw := strings.TrimSpace(trgt.Port)

	if alias == "" {
		return nil, Target{}, nil, fmt.Errorf("empty %s alias", protocol)
	}

	tgt = Target{Protocol: protocol, Spec: host.Spec{Alias: alias, User: user, HostName: hostName}}

	var args []string
	switch protocol {
	case "ssh":
		// ssh connects by alias; hostname/port are only for display/preflight
		if hostName != "" {
			p, err := NormalizePort(portRaw, "ssh")
			if err != nil {
				return nil, Target{}, nil, err
			}
			tgt.Port = p
		}
		if user != "" {
			args = append(args, "-l", user, alias)
		} else {
			args = append(args, alias)
		}

	case "telnet":
		// telnet connects by hostname and port
		if hostName == "" {
			return nil, Target{}, nil, fmt.Errorf("telnet %q: empty hostname", alias)
		}
		p, err := NormalizePort(portRaw, "telnet")
		if err != nil {
			return nil, Target{}, nil, err
		}
		tgt.Port = p
		args = []string{hostName, p}
	}

	cmd = exec.Command(programPath, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	tail = NewTailBuffer(4096)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return cmd, tgt, tail, nil
}
