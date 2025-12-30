package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO:
// always format host names when displaying in status
//
//       put A add on main help text, remove from details help
//         since add should be global and always available (confusing to have it only in details)
//       maybe make E edit also global (makes editing group names a possibility)
//
//       implement remove host functionality
//           - can be a confirmation prompt
//       implement arrow keys for navigating HostForm fields (require protocol selection first on addHost)
//           - or maybe make the whole form a selectable list and have protocol default to ssh on addHost
//              - make esc the only cancel key so left/right can move cursor in input fields
//       add full help text for HostForm
//           - add "up/down arrow keys to navigate fields", "enter submits", "esc cancels" etc. hints
//           - add Validation to form fields and show errors in the title
//       remove ssh options when adding a telnet host
//           - make them a different form section that only shows for ssh protocol (give option to skip?)
//       add confirmation prompt in hostForm
//           - on cancel, "Are you sure you want to cancel? All changes will be lost."
//           - on submit, "Are you sure you want to save these changes?"
//
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
