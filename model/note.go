package model

type Note struct {
	Text     string	`json:"text" db:"text"`
	Username string	`json:"username" db:"username"`
}
