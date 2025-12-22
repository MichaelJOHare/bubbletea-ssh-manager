package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: always format host names when displaying in status
//       implement saving username for hosts
//       implement edit/remove/add host functionality
//       improve ? info display formatting (and location and content)
//       make help color scheme easier to read
//       swap from group.HOST to GROUP.host for consistency

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
