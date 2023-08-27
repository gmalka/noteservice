package noteservice

import "noteservice/model"

type NoteStore interface {
	CreateNote(note model.Note) error
	GetAllNotes(username string) ([]model.Note, error)
}

type NoteService interface {
	CreateNote(note model.Note) error
	GetAllNotes(username string) ([]model.Note, error)
}

type noteService struct {
	store NoteStore
}

func NewNoteService(store NoteStore) NoteService {
	return noteService{
		store: store,
	}
}

func (n noteService) CreateNote(note model.Note) error {
	return n.store.CreateNote(note)
}

func (n noteService) GetAllNotes(username string) ([]model.Note, error) {
	return n.store.GetAllNotes(username)
}