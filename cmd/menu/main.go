package main

import (
	"fmt"
	"os"

	tui "bubbletea-ssh-manager/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// TODO:
//
//        *** Short-term ideas ***
//
//       maybe make E edit also global (makes editing group names a possibility)?
//          -- actually just make E and R able to be used on groups too (maybe not, deleting groups could be messy)
//       allow changing protocol in edit host form?
// 	 	 check for empty list before opening host details
//       check menu_handle.go handleConnectFinishedMsg comment
//       change color names back to actual colors in theme.go
//       handle immediate ssh errors better (especially multi-line errors)
//
//
//       *** Medium-term ideas ***
//
//       change how current groups are displayed in host form status panel (see note in form_status.go)
//       add real validation to host form inputs (need to make sure the error it shows is clear about what is wrong)
//       add placeholder text to form inputs
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
//          - move everything else to internal (eg. internal/tui/model.go, internal/tui/keys(views, forms), etc.)
//       make protocol a type with constants
//       add --version flag

func main() {
	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
