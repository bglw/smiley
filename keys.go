package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Send    key.Binding
	History key.Binding
	Log     key.Binding
	Quit    key.Binding
}

var CurrentKeyMap = KeyMap{
	Send:    key.NewBinding(key.WithKeys("tab", "ctrl+j")),
	History: key.NewBinding(key.WithKeys("ctrl+h")),
	Quit:    key.NewBinding(key.WithKeys("ctrl+c")),
	Log:     key.NewBinding(key.WithKeys("ctrl+l")),
}
