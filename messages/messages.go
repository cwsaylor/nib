package messages

import "socnotes/types"

type NotesLoadedMsg struct{ Notes []types.Note }
type NoteSavedMsg struct{ Note types.Note }
type NoteDeletedMsg struct{ ID int }
type SearchResultsMsg struct{ Results []types.Note }
type EditNoteMsg struct{ Note types.Note }
type ConfirmDeleteMsg struct{ Note types.Note }
type YankDoneMsg struct{}
type StatusMsg struct{ Text string }
type DebounceTickMsg struct{}
type StatusClearMsg struct{}
type ErrMsg struct{ Err error }
type TrashedNotesMsg struct{ Notes []types.Note }
type NoteRestoredMsg struct{ ID int }
type NotePurgedMsg struct{ ID int }
type NoteCreatedMsg struct{ Note types.Note }
