package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: always format host names when displaying in status
//       implement edit/remove/add host functionality
//          -- Add and Edit can use textinputs components to gather info (see examples in bubbletea repo)
//                - include nickname, groupname, hostname, and the rest of host.Spec
//          -- Remove can be a confirmation prompt
//       swap from group.HOST to GROUP.host for consistency?

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
