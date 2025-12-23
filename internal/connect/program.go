package connect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PreferredProgramPath returns the preferred full path to the given program name.
//
// On Windows, it prefers MSYS2 binaries if available.
// On other platforms, it looks in the system PATH.
func PreferredProgramPath(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty program name")
	}

	// prefer MSYS2 binaries when running on Windows
	if runtime.GOOS == "windows" {
		roots := []string{}
		// default install location
		roots = append(roots, `C:\msys64`)
		for _, root := range roots {
			p := filepath.Join(root, "usr", "bin", name+".exe")
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}

	// fallback to PATH lookup
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return p, nil
}
