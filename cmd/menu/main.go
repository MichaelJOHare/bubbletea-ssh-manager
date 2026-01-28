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
//       reproduce bug where after logging out of SMG openvms app, sometimes the menu doesn't redraw properly
//          -- ie. the list keeps shifting up on a ~1 second interval
//             -- remember it happening with Console Manager when it was showing BYTLLM errors
//       write down other successful exit status that is sometimes returned when logging out of openvms
//
//       fix remove host status being set right away instead of after confirmation
//          -- to test, open remove host modal, wait ~10 seconds, then cancel - status should show for full 10 seconds
//       maybe make E edit also global (makes editing group names a possibility)?
//          -- actually just make E able to be used on groups too (name only, warn it will change all child aliases)
//       allow changing protocol in edit host form?
//       check model_handle.go handleConnectFinishedMsg comment
//       change color names back to actual colors in theme.go
//       handle immediate ssh errors better (ie. bad config options)
//       make ssh options a select/multi-select field in host form and add all of them?
//
//
//       *** Medium-term ideas ***
//
//       change how current groups are displayed in host form status panel (see note in form_status.go)
//       add real validation to host form inputs (need to make sure the error it shows is clear about what is wrong)
//       add placeholder text to form inputs
//       add group name autocomplete in host form (see huh docs for how to do this - suggestions i think it's called)
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
//       move relayout calls to a better place (not after every modal open/close) - maybe in update loop after handling msg?
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
