package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: always format host names when displaying in status
//       implement edit/remove/add host functionality
//       make ? bring up S, E, A, R help menu which we'll use to show details, add, edit, or remove hosts
//          -- Add and Edit can use textinputs components to gather info (see examples in bubbletea repo)
//                - include nickname, groupname, hostname, and the rest of host.Spec
//          -- Remove can be a confirmation prompt
//       make help color scheme easier to read
//       swap from group.HOST to GROUP.host for consistency

// print clear screen escape on quitting?

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
