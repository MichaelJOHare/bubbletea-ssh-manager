package connect

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Request struct {
	Protocol    string // "ssh" or "telnet"
	DisplayName string // optional; used only for friendlier error messages
	Alias       string // ssh-style Host alias from the config
	User        string // optional user name
	Host        string // hostname or IP address
	Port        string // port number as string
}

// BuildCommand builds the exec.Cmd to connect.
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
	host := strings.TrimSpace(req.Host)
	portRaw := strings.TrimSpace(req.Port)

	if alias == "" {
		return nil, Target{}, nil, fmt.Errorf("empty %s alias", protocol)
	}

	tgt = Target{protocol: protocol, alias: alias, user: user, host: host}

	var args []string
	switch protocol {
	case "ssh":
		// ssh connects by alias; hostname/port are only for display/preflight
		if host != "" {
			p, err := NormalizePort(portRaw, "ssh")
			if err != nil {
				return nil, Target{}, nil, err
			}
			tgt.port = p
		}
		if user != "" {
			args = []string{"-l", user, alias}
		} else {
			args = []string{alias}
		}

	case "telnet":
		// telnet connects by hostname and port
		if host == "" {
			return nil, Target{}, nil, fmt.Errorf("telnet %q: empty hostname", alias)
		}
		p, err := NormalizePort(portRaw, "telnet")
		if err != nil {
			return nil, Target{}, nil, err
		}
		tgt.port = p
		args = []string{host, p}
	}

	cmd = exec.Command(programPath, args...)
	cmd.Stdin = os.Stdin

	tail = NewTailBuffer(4096)
	cmd.Stdout = io.MultiWriter(os.Stdout, tail)
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	return cmd, tgt, tail, nil
}
