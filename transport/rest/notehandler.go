package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"noteservice/model"
	"noteservice/service/yandexspellerservice"
)

func (h Handler) AllNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.log.Infof("Attempt to access non-existent path: %v | %v", r.Method, r.URL.Path)
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}
	val := r.Context().Value(ContextKeyUsername)
	username := val.(string)

	notes, err := h.notes.GetAllNotes(username)
	if err != nil {
		h.log.Errorf("Can't get all notes of user %s: %v", username, err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	result := make([]string, len(notes))
	for k, v := range notes {
		result[k] = v.Text
	}

	b, err := json.Marshal(result)
	if err != nil {
		h.log.Errorf("Can't marshal data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h Handler) NotePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	val := r.Context().Value(ContextKeyUsername)
	username := val.(string)

	result := make([]byte, 0, 10)

	for {
		buf := make([]byte, 10)
		n, err := r.Body.Read(buf)

		result = append(result, buf[:n]...)
		if err == io.EOF {
			break
		}

		if err != nil {
			h.log.Errorf("Can't read input data: %v", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}

	input := model.Input{}
	err := json.Unmarshal(result, &input)
	if err != nil {
		h.log.Errorf("Can't unmarshal data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	b, err := h.speller.CheckText(input.Text)
	if err == yandexspellerservice.ErrSyntax {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(b)
		return
	}
	if err != nil {
		h.log.Errorf("Can't check text by speller: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	err = h.notes.CreateNote(model.Note{
		Text:     input.Text,
		Username: username,
	})
	if err != nil {
		h.log.Errorf("Can't create note: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
