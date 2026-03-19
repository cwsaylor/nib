package commands

import (
	"os/exec"
	"runtime"
	"socnotes/db"
	"socnotes/messages"
	"socnotes/types"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadNotes(s *db.DB) tea.Cmd {
	return func() tea.Msg {
		notes, err := s.ListActive()
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NotesLoadedMsg{Notes: notes}
	}
}

func SaveNote(s *db.DB, n types.Note) tea.Cmd {
	return func() tea.Msg {
		saved, err := s.Upsert(n)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NoteSavedMsg{Note: saved}
	}
}

func CreateNote(s *db.DB) tea.Cmd {
	return func() tea.Msg {
		note, err := s.Create("", "")
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NoteCreatedMsg{Note: note}
	}
}

func DeleteNote(s *db.DB, id int) tea.Cmd {
	return func() tea.Msg {
		if err := s.SoftDelete(id); err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NoteDeletedMsg{ID: id}
	}
}

func SearchNotes(s *db.DB, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := s.Search(query)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.SearchResultsMsg{Results: results}
	}
}

func LoadTrashed(s *db.DB) tea.Cmd {
	return func() tea.Msg {
		notes, err := s.ListTrashed()
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.TrashedNotesMsg{Notes: notes}
	}
}

func RestoreNote(s *db.DB, id int) tea.Cmd {
	return func() tea.Msg {
		if err := s.Restore(id); err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NoteRestoredMsg{ID: id}
	}
}

func PurgeNote(s *db.DB, id int) tea.Cmd {
	return func() tea.Msg {
		if err := s.HardDelete(id); err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.NotePurgedMsg{ID: id}
	}
}

func LoadNote(s *db.DB, id int) tea.Cmd {
	return func() tea.Msg {
		note, err := s.GetNote(id)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.EditNoteMsg{Note: note}
	}
}

func Yank(body string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		case "linux":
			cmd = exec.Command("xclip", "-selection", "clipboard")
		default:
			cmd = exec.Command("clip.exe")
		}
		cmd.Stdin = strings.NewReader(body)
		if err := cmd.Run(); err != nil {
			return messages.ErrMsg{Err: err}
		}
		return messages.YankDoneMsg{}
	}
}

func StatusFlash(text string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return messages.StatusMsg{Text: text} },
		tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return messages.StatusClearMsg{}
		}),
	)
}
