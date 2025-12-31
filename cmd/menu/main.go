package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO:
//
//        *** Short-term ideas ***
//
//       maybe make E edit also global (makes editing group names a possibility)?
// 	 	 check for empty list before opening host details
//       check menu_handle.go handleConnectFinishedMsg comment
//
//
//       *** Medium-term ideas ***
//
//       implement remove host functionality
//           - can be a confirmation prompt
//       add current groups to status display to make adding to groups easier
//       add placeholder text to form inputs
//       add pagination hint to host form when protocol is ssh
//       add confirmation prompt in hostForm
//           - on cancel, "Are you sure you want to cancel? All changes will be lost."
//           - on submit, "Are you sure you want to save these changes?"
//              - make enter not submit forms when validation errors exist
//
//
//        *** Long-term ideas ***
//
//	     maybe have group names become a separate list
//            - should be able to be focused and navigated (maybe tab to switch between)
//       always format host names when displaying in status
//       add icon for executable
//       add config file for environment settings (eg. default user, default port, paths, etc.)
//       fix silent errors in parser.go
//       change model pointer receiver methods to value receivers where possible
//       look into context.Context for managing preflight timeouts/cancellations
//       refactor modeState/formState into separate structs for each mode (discriminated union?)
//       refactor menu package to only contain main
//          - move everything else to internal (eg. internal/tui/model.go, internal/tui/keys(views, forms), etc.)
//       make protocol a type with constants

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
