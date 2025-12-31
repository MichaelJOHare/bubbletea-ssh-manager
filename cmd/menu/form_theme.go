package main

import "github.com/charmbracelet/huh"

// hostFormTheme returns a huh.Theme customized for the host add/edit form
func hostFormTheme(app Theme) *huh.Theme {
	t := huh.ThemeCharm()

	t.Focused.Title = t.Focused.Title.Foreground(app.SelectedItemTitle)
	t.Blurred.Title = t.Blurred.Title.Foreground(app.StatusDefault)

	// "|" active form item
	t.Focused.Base = t.Focused.Base.BorderForeground(app.SelectedItemBorder)

	// protocol selector option color
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(app.StatusSuccess)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(app.StatusDefault)

	// ">" indicator for select + multiselect
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(app.PreflightSpinner)

	// ">" prompt for text inputs
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(app.PreflightSpinner)
	t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(app.PreflightSpinner)

	// cursor color
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(app.StatusDefault)
	t.Focused.TextInput.CursorText = t.Focused.TextInput.CursorText.Foreground(app.StatusDefault)
	t.Blurred.TextInput.Cursor = t.Blurred.TextInput.Cursor.Foreground(app.StatusDefault)
	t.Blurred.TextInput.CursorText = t.Blurred.TextInput.CursorText.Foreground(app.StatusDefault)

	return t
}

// confirmFormTheme returns a huh.Theme customized for confirmation dialogs
func confirmFormTheme(app Theme) *huh.Theme {
	t := huh.ThemeCharm()

	t.Focused.Title = t.Focused.Title.Foreground(app.StatusError)
	t.Blurred.Title = t.Blurred.Title.Foreground(app.StatusDefault)

	t.Focused.Description = t.Focused.Description.Foreground(app.StatusDefault)
	t.Blurred.Description = t.Blurred.Description.Foreground(app.StatusDefault)

	// selected/unselected option colors for Yes/No
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(app.StatusSuccess)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(app.StatusDefault)

	// ">" indicator
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(app.PreflightSpinner)

	return t
}
