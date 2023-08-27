package yandexspellerservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"noteservice/model"
	"strings"
)

const (
	ERROR_UNKNOWN_WORD    = "Слова нет в словаре"
	ERROR_REPEAT_WORD     = "Повтор слова"
	ERROR_CAPITALIZATION  = "Неверное употребление прописных и строчных букв"
	ERROR_TOO_MANY_ERRORS = "Текст содержит слишком много ошибок"

	path = "https://speller.yandex.net/services/spellservice.json/checkText?text="
)

var ErrSyntax = errors.New("SyntaxError")

type spellerService struct {
}

type Speller interface {
	CheckText(text string) ([]byte, error)
}

func NewSpeller() Speller {
	return spellerService{}
}

func (s spellerService) CheckText(text string) ([]byte, error) {
	t := strings.Builder{}

	word := make([]rune, 0, 10)

	for _, v := range text {
		if v == ' ' {
			t.WriteString(string(word))
			t.WriteByte('+')
			word = word[:0]
		} else if v == '\n' {
			t.WriteString(string(word))
			t.WriteByte('+')
			t.WriteString("%0A")
			word = word[:0]
		} else {
			word = append(word, v)
		}
	}
	t.WriteString(string(word))

	req, err := http.NewRequest(http.MethodGet, path+t.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't make request: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read from speller's body: %v", err)
	}

	res := make([]model.Result, 0, 10)
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal from speller's data return: %v", err)
	}

	response := model.Response{}
	var myerr error
	if len(res) != 0 {
		myerr = ErrSyntax
		response.Result = "Syntax error"
		response.Errors = make([]model.IncorrectWord, len(res))

		for k, v := range res {
			switch v.Code {
			case 1:
				response.Errors[k].Error = ERROR_UNKNOWN_WORD
			case 2:
				response.Errors[k].Error = ERROR_REPEAT_WORD
			case 3:
				response.Errors[k].Error = ERROR_CAPITALIZATION
			case 4:
				response.Errors[k].Error = ERROR_TOO_MANY_ERRORS
			}

			response.Errors[k].Word = v.Word
			response.Errors[k].Replacements = v.S
		}
	} else {
		response.Result = "Success"
	}

	b, err = json.Marshal(response)
	if err != nil {
		return nil ,fmt.Errorf("cant parse response while text checking: %v", err)
	}

	return b, myerr
}