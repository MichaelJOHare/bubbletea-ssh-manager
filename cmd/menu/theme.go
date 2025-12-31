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

	// Protocol colors
	ProtocolSSH    lipgloss.Color
	ProtocolTelnet lipgloss.Color

	// UI element colors
	SelectedItemBorder lipgloss.Color
	SearchLabel        lipgloss.Color
	SelectedItemTitle  lipgloss.Color
	UsernamePrompt     lipgloss.Color
	GroupName          lipgloss.Color
	PreflightSpinner   lipgloss.Color
	PreflightText      lipgloss.Color
	DetailsHeader      lipgloss.Color
	DetailsBorder      lipgloss.Color
	DetailsLabel       lipgloss.Color
	OptionsLabel       lipgloss.Color

	// Help/key colors
	HelpText  lipgloss.Color
	KeyCursor lipgloss.Color
	KeyBack   lipgloss.Color
	KeyAdd    lipgloss.Color
	KeyQuit   lipgloss.Color
	KeyInfo   lipgloss.Color
	KeyClear  lipgloss.Color
	KeyClose  lipgloss.Color
	KeyEdit   lipgloss.Color
	KeyRemove lipgloss.Color
	KeyEnter  lipgloss.Color
}

// DefaultTheme returns the default Theme with preset color values.
func DefaultTheme() Theme {
	return Theme{
		StatusDefault: lipgloss.Color("#bcbcbc"),
		StatusError:   lipgloss.Color("#e06c75"),
		StatusSuccess: lipgloss.Color("#98c379"),

		ProtocolSSH:    lipgloss.Color("#98c379"),
		ProtocolTelnet: lipgloss.Color("#ea8665"),

		SelectedItemBorder: lipgloss.Color("#882d90"),
		SelectedItemTitle:  lipgloss.Color("#ad58b4"),
		SearchLabel:        lipgloss.Color("#8787ff"),
		UsernamePrompt:     lipgloss.Color("#ddb034"),
		GroupName:          lipgloss.Color("#e48315"),
		PreflightText:      lipgloss.Color("#8d8d8d"),
		PreflightSpinner:   lipgloss.Color("#ff0087"),
		DetailsHeader:      lipgloss.Color("#8787ff"),
		DetailsBorder:      lipgloss.Color("#61afef"),
		DetailsLabel:       lipgloss.Color("#ddb034"),
		OptionsLabel:       lipgloss.Color("#ddb034"),

		HelpText:  lipgloss.Color("#949494"),
		KeyCursor: lipgloss.Color("#98c379"),
		KeyBack:   lipgloss.Color("#c678dd"),
		KeyAdd:    lipgloss.Color("#c678dd"),
		KeyQuit:   lipgloss.Color("#e06c75"),
		KeyInfo:   lipgloss.Color("#61afef"),
		KeyClear:  lipgloss.Color("#ddb034"),
		KeyClose:  lipgloss.Color("#e06c75"),
		KeyEdit:   lipgloss.Color("#98c379"),
		KeyRemove: lipgloss.Color("#e06c75"),
		KeyEnter:  lipgloss.Color("#98c379"),
	}
}

func GreenEnter() string {
	return lipgloss.NewStyle().Foreground(DefaultTheme().KeyEnter).Render("Enter")
}
