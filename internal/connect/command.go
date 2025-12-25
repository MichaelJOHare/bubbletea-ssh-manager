package connect

import (
	"bubbletea-ssh-manager/internal/host"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Request struct {
	Protocol    string // "ssh" or "telnet"
	DisplayName string // optional; used only for friendlier error messages
	host.Spec          // shared host fields (alias/hostname/port/user)
}

// BuildCommand builds the exec.Cmd to connect to the given Request.
//
// It returns a Target for display/title, and a TailBuffer that captures the last
// part of the command output for error reporting.
func BuildCommand(req Request) (cmd *exec.Cmd, tgt Target, tail *TailBuffer, err error) {
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	if protocol != "ssh" && protocol != "telnet" {
		name := strings.TrimSpace(req.DisplayName)
		if name == "" {
			name = strings.TrimSpace(req.Alias)
		}
		return nil, Target{}, nil, fmt.Errorf("unknown protocol for %s: %q", name, req.Protocol)
	}

	programPath, err := PreferredProgramPath(protocol)
	if err != nil {
		return nil, Target{}, nil, fmt.Errorf("%s not found: %w", protocol, err)
	}

	alias := strings.TrimSpace(req.Alias)
	user := strings.TrimSpace(req.User)
	hostName := strings.TrimSpace(req.HostName)
	portRaw := strings.TrimSpace(req.Port)

	if alias == "" {
		return nil, Target{}, nil, fmt.Errorf("empty %s alias", protocol)
	}

	tgt = Target{protocol: protocol, Spec: host.Spec{Alias: alias, User: user, HostName: hostName}}

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
