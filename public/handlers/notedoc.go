package handlers

import "noteservice/model"

// swagger:route POST /{username}/note notes NoteRequest
// Создать новую заметку.
// security:
//	 - Cookie: []
//   - Bearer: []
// responses:
//   200: SuccessNote
//	 422: UnprocessableEntity
//   500: ServerErrorNotePostResponse

// swagger:parameters NoteRequest
type NoteRequest struct {
	// in:path
	Username string `json:"username"`
	// in:body
	Data model.Input `json:"note"`
}

// swagger:response SuccessNote
type SuccessNote struct {
	// in:body
	Response model.Response `json:"response"`
}

// swagger:response UnprocessableEntity
type UnprocessableEntity struct {
	// in:body
	Response model.Response `json:"response"`
}

// swagger:response ServerErrorNotePostResponse
type ServerErrorNotePostResponse struct {
	// in:body
	Message string `json:"message"`
}

// swagger:route GET /{username}/notes notes NotesRequest
// Получить все заметки пользователя.
// security:
//	 - Cookie: []
//   - Bearer:
// responses:
//   200: SuccessNotes
//   500: ServerErrorNoteGetResponse

// swagger:parameters NotesRequest
type NotesRequest struct {
	// in:path
	Username string `json:"username"`
}

// swagger:response SuccessNotes
type SuccessNotes struct {
	// in:body
	Notes []string `json:"notes"`
}

// swagger:response ServerErrorNoteGetResponse
type ServerErrorNoteGetResponse struct {
	// in:body
	Message string `json:"message"`
}