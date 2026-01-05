package connect

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"bubbletea-ssh-manager/internal/config"
	str "bubbletea-ssh-manager/internal/stringutil"
)

// preferredProgramPath returns the preferred full path to the ssh/telnet program for the given protocol name.
//
// On Windows, it prefers MSYS2 binaries if available.
// On other platforms, it looks in the system PATH.
func preferredProgramPath(protocol config.Protocol) (string, error) {
	if protocol == "" {
		return "", fmt.Errorf("empty program name")
	}

	// prefer MSYS2 binaries when running on Windows
	if runtime.GOOS == "windows" {
		roots := []string{}
		// default install location
		roots = append(roots, `C:\msys64`)
		for _, root := range roots {
			p := filepath.Join(root, "usr", "bin", string(protocol)+".exe")
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}

	// fallback to PATH lookup
	p, err := exec.LookPath(string(protocol))
	if err != nil {
		return "", err
	}
	return p, nil
}

// BuildCommand builds the exec.Cmd to connect to the given Target.
//
// It returns a Target for display/title, and a TailBuffer that captures the last
// part of the command output for error reporting.
func BuildCommand(trgt Target) (cmd *exec.Cmd, tgt Target, tail *TailBuffer, err error) {
	// Boundary normalization: callers should already provide normalized specs
	// (from parsed config or form submit), but normalize defensively once here.
	trgt.Spec = trgt.Spec.Normalized()

	if trgt.Protocol != config.ProtocolSSH && trgt.Protocol != config.ProtocolTelnet {
		name := trgt.Alias
		return nil, Target{}, nil, fmt.Errorf("unknown protocol for %s: %q", name, trgt.Protocol)
	}

	programPath, err := preferredProgramPath(trgt.Protocol)
	if err != nil {
		return nil, Target{}, nil, fmt.Errorf("%s not found: %w", trgt.Protocol, err)
	}

	alias := trgt.Alias
	user := trgt.User
	hostName := trgt.HostName
	portRaw := trgt.Port

	if alias == "" {
		return nil, Target{}, nil, fmt.Errorf("empty %s alias", trgt.Protocol)
	}

	tgt = Target{Protocol: trgt.Protocol, Spec: config.Spec{Alias: alias, User: user, HostName: hostName}}

	var args []string
	switch trgt.Protocol {
	case config.ProtocolSSH:
		// ssh connects by alias; hostname/port are only for display/preflight
		if hostName != "" {
			p, err := str.NormalizePort(portRaw, config.ProtocolSSH)
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

	case config.ProtocolTelnet:
		// telnet connects by hostname and port
		if hostName == "" {
			return nil, Target{}, nil, fmt.Errorf("telnet %q: empty hostname", alias)
		}
		p, err := str.NormalizePort(portRaw, config.ProtocolTelnet)
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
