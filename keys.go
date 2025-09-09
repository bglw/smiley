package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Send key.Binding
}

var CurrentKeyMap = KeyMap{
	Send: key.NewBinding(key.WithKeys("tab", "ctrl+j")),
}
