package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Send     key.Binding
	History  key.Binding
	Log      key.Binding
	Quit     key.Binding
	Switch   key.Binding
	Modal    key.Binding
	Followup key.Binding
}

var CurrentKeyMap = KeyMap{
	Send:     key.NewBinding(key.WithKeys("tab", "ctrl+j")),
	History:  key.NewBinding(key.WithKeys("ctrl+h")),
	Quit:     key.NewBinding(key.WithKeys("ctrl+c")),
	Log:      key.NewBinding(key.WithKeys("ctrl+l")),
	Switch:   key.NewBinding(key.WithKeys("shift+tab")),
	Modal:    key.NewBinding(key.WithKeys("ctrl+k")),
	Followup: key.NewBinding(key.WithKeys("ctrl+n")),
}
