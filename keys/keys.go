package keys

import "github.com/charmbracelet/bubbles/key"

type ListKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	New    key.Binding
	Edit   key.Binding
	Search key.Binding
	Delete key.Binding
	Yank   key.Binding
	Trash  key.Binding
	Help   key.Binding
	Quit   key.Binding
}

var ListKeys = ListKeyMap{
	Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Edit:   key.NewBinding(key.WithKeys("e", "enter"), key.WithHelp("e/↵", "edit")),
	Search: key.NewBinding(key.WithKeys("s", "/"), key.WithHelp("s//", "search")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Yank:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yank")),
	Trash:  key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "trash")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:   key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

type EditorKeyMap struct {
	Save    key.Binding
	Back    key.Binding
	Discard key.Binding
}

var EditorKeys = EditorKeyMap{
	Save:    key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "save & back")),
	Discard: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "discard")),
}

type TrashKeyMap struct {
	Restore key.Binding
	Purge   key.Binding
	Back    key.Binding
	Up      key.Binding
	Down    key.Binding
}

var TrashKeys = TrashKeyMap{
	Restore: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restore")),
	Purge:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "purge")),
	Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Up:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:    key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
}
