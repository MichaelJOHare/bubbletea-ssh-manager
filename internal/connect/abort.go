package connect

import (
	"errors"
	"strings"
)

const connectionAbortedExitStatus = "exit status 512"

// ErrAborted is a sentinel error used when the app explicitly cancels a connect
// attempt (for example, cancelling a preflight check).
//
// Prefer wrapping/returning this error over string-matching when possible.
var ErrAborted = errors.New("btms: aborted")

// IsConnectionAborted returns true if the given error indicates
// that the connection was aborted by user (eg. Ctrl+C).
func IsConnectionAborted(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAborted) {
		return true
	}

	s := strings.TrimSpace(err.Error())
	if s == connectionAbortedExitStatus {
		return true
	}
	if strings.Contains(s, connectionAbortedExitStatus) {
		return true
	}
	ls := strings.ToLower(s)
	if strings.Contains(ls, "signal: interrupt") {
		return true
	}

	return false
}
