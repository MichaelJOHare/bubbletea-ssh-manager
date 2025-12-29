package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: always format host names when displaying in status
//       implement remove host functionality
//          	-- can be a confirmation prompt
//       implement arrow keys for navigating HostForm fields
// 	 	 check for empty list before opening host details
//       fix double newline when adding/editing hosts
//       consolidate color styles

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
