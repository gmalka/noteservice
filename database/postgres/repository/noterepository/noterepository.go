package noterepository

import (
	"fmt"
	"noteservice/model"

	"github.com/jmoiron/sqlx"
)

type noteRepository struct {
	db *sqlx.DB
}

type NoteStore interface {
	CreateNote(note model.Note) error
	GetAllNotes(username string) ([]model.Note, error)
}

func NewNoteStore(db *sqlx.DB) NoteStore {
	return noteRepository{db: db}
}

func (n noteRepository) CreateNote(note model.Note) error {
	_, err := n.db.Exec("INSERT INTO notes(text, username) VALUES($1,$2)", note.Text, note.Username)

	return err
}

func (n noteRepository) GetAllNotes(username string) ([]model.Note, error) {
	rows, err := n.db.Query("SELECT text FROM notes WHERE username=$1", username)
	if err != nil {
		return nil, fmt.Errorf("can't get all notes: %v", err)
	}

	notes := make([]model.Note, 0, 10)

	for rows.Next() {
		var t string

		rows.Scan(&t)
		notes = append(notes, model.Note{
			Text:     t,
			Username: username,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("can't scan notes from postgres: %v", err)
	}

	return notes, nil
}
