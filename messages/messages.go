package messages

import "nib/types"

type NotesLoadedMsg struct {
	Notes   []types.Note
	HasMore bool
}
type MoreNotesLoadedMsg struct {
	Notes   []types.Note
	HasMore bool
}
type NoteSavedMsg struct{ Note types.Note }
type NoteDeletedMsg struct{ ID int }
type SearchResultsMsg struct {
	Results []types.Note
	HasMore bool
}
type MoreSearchResultsMsg struct {
	Results []types.Note
	HasMore bool
}
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
