package model

type Response struct {
	Result string          `json:"result"`
	Errors []IncorrectWord `json:"errors,omitempty"`
}

type IncorrectWord struct {
	Error        string   `json:"error"`
	Word         string   `json:"word"`
	Replacements []string `json:"replacements,omitempty"`
}

type Message struct {
	Message string `json:"message"`
}