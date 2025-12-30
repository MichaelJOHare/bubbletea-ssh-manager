package main

import "github.com/charmbracelet/lipgloss"

// Theme groups UI colors used by the TUI so styling is centralized.
//
// For now this only captures color values (not full lipgloss styles), so callers
// can compose styles locally without introducing global side-effects.
type Theme struct {
	// Status line colors
	StatusDefault lipgloss.Color
	StatusError   lipgloss.Color
	StatusSuccess lipgloss.Color

	// UI element colors
	SearchLabel      lipgloss.Color
	DetailsHeader    lipgloss.Color
	DetailsBorder    lipgloss.Color
	ProtocolSSH      lipgloss.Color
	UsernamePrompt   lipgloss.Color
	PreflightSpinner lipgloss.Color
	ProtocolTelnet   lipgloss.Color
	GroupName        lipgloss.Color
	PreflightText    lipgloss.Color
	OptionsLabel     lipgloss.Color

	// Help/key colors
	HelpText  lipgloss.Color
	KeyCursor lipgloss.Color
	KeyBack   lipgloss.Color
	KeyAdd    lipgloss.Color
	KeyQuit   lipgloss.Color
	KeyInfo   lipgloss.Color
	KeyClear  lipgloss.Color
	KeyEdit   lipgloss.Color
	KeyRemove lipgloss.Color
	KeyEnter  lipgloss.Color
}

func DefaultTheme() Theme {
	return Theme{
		StatusDefault: lipgloss.Color("#bcbcbc"),
		StatusError:   lipgloss.Color("#e06c75"),
		StatusSuccess: lipgloss.Color("#98c379"),

		SearchLabel:      lipgloss.Color("#8787ff"),
		DetailsHeader:    lipgloss.Color("#61afef"),
		DetailsBorder:    lipgloss.Color("#61afef"),
		ProtocolSSH:      lipgloss.Color("#98c379"),
		UsernamePrompt:   lipgloss.Color("#ddb034"),
		ProtocolTelnet:   lipgloss.Color("#ff875f"),
		GroupName:        lipgloss.Color("#e48315"),
		PreflightText:    lipgloss.Color("#8d8d8d"),
		PreflightSpinner: lipgloss.Color("#ff0087"),
		OptionsLabel:     lipgloss.Color("#ddb034"),

		HelpText:  lipgloss.Color("#949494"),
		KeyCursor: lipgloss.Color("#98c379"),
		KeyBack:   lipgloss.Color("#c678dd"),
		KeyAdd:    lipgloss.Color("#c678dd"),
		KeyQuit:   lipgloss.Color("#e06c75"),
		KeyInfo:   lipgloss.Color("#61afef"),
		KeyClear:  lipgloss.Color("#ddb034"),
		KeyEdit:   lipgloss.Color("#98c379"),
		KeyRemove: lipgloss.Color("#e06c75"),
		KeyEnter:  lipgloss.Color("#98c379"),
	}
}
