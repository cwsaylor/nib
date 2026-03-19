package model

import (
	"nib/commands"
	"nib/db"
	"nib/messages"
	"nib/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppMode int

const (
	ModeList AppMode = iota
	ModeEdit
	ModeSearch
	ModeConfirmDelete
	ModeTrash
)

type InputMode int

const (
	InputNavigation InputMode = iota
	InputText
)

type Model struct {
	mode      AppMode
	inputMode InputMode
	db        *db.DB
	list      ListModel
	editor    EditorModel
	search    SearchModel
	confirm   ConfirmModel
	trash     TrashModel
	status    string
	showHelp  bool
	width     int
	height    int
	err       error
}

func New(store *db.DB) Model {
	return Model{
		mode:      ModeList,
		inputMode: InputNavigation,
		db:        store,
		list:      NewListModel(),
		editor:    NewEditorModel(),
		search:    NewSearchModel(),
		trash:     NewTrashModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return commands.LoadNotes(m.db)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		m.editor.SetSize(msg.Width, msg.Height)
		m.search.SetSize(msg.Width, msg.Height)
		m.trash.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "?" && m.inputMode == InputNavigation {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

	case messages.ErrMsg:
		m.err = msg.Err
		return m, nil

	case messages.StatusMsg:
		m.status = msg.Text
		return m, nil

	case messages.StatusClearMsg:
		m.status = ""
		return m, nil
	}

	switch m.mode {
	case ModeList:
		return m.updateList(msg)
	case ModeEdit:
		return m.updateEdit(msg)
	case ModeSearch:
		return m.updateSearch(msg)
	case ModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case ModeTrash:
		return m.updateTrash(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string
	switch m.mode {
	case ModeList:
		content = m.list.View(m.width, m.height)
	case ModeEdit:
		content = m.editor.View()
	case ModeSearch:
		content = m.search.View(m.width, m.height)
	case ModeConfirmDelete:
		bg := m.list.View(m.width, m.height)
		dialog := m.confirm.View()
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")))
		_ = bg
	case ModeTrash:
		content = m.trash.View(m.width, m.height)
	}

	// Add status bar if there's a status message
	if m.status != "" {
		statusBar := theme.StatusBar.Width(m.width).Render(m.status)
		content = lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
	}

	if m.err != nil {
		errBar := theme.ErrorText.Width(m.width).Render("Error: " + m.err.Error())
		content = lipgloss.JoinVertical(lipgloss.Left, content, errBar)
	}

	if m.showHelp {
		content = helpView(m.width, m.height)
	}

	return content
}
