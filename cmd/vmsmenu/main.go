package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO: Always format host names when displaying in status
//       Add hosts in groups to searchable list
//       Implement edit/remove/add host functionality
//       Improve ? info display formatting
//       Adjust help color scheme for better visibility

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
