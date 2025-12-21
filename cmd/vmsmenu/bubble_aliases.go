package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

type textinputModel = textinput.Model // text input model from bubbletea/bubbles

type listModel = list.Model // list model from bubbletea/bubbles

type menuDelegate struct {
	list.DefaultDelegate // embed default delegate from bubbletea/bubbles/list to extend it
	groupHints           map[*menuItem]string
}
