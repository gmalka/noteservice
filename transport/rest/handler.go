package rest

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"noteservice/model"
	"noteservice/service/tokenservice"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ContextKey string

const ContextKeyUsername ContextKey = "username"

type UserService interface {
	CreateUser(user model.User) error
	GetUser(username string) (model.User, error)
}

type NoteService interface {
	CreateNote(note model.Note) error
	GetAllNotes(username string) ([]model.Note, error)
}

type TokenManager interface {
	CreateToken(userinfo model.UserInfo, kind int) (string, error)
	ParseToken(inputToken string, kind int) (model.UserClaims, error)
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(verifiable, wanted string) error
}

type Speller interface {
	CheckText(text string) ([]byte, error)
}

type Handler struct {
	log        *logrus.Logger
	users      UserService
	notes      NoteService
	tokens     TokenManager
	passhasher PasswordHasher
	speller    Speller
}

func NewHandler(log *logrus.Logger, users UserService, notes NoteService, tokens TokenManager, passhasher PasswordHasher, speller Speller) Handler {
	return Handler{
		log:        log,
		users:      users,
		notes:      notes,
		tokens:     tokens,
		passhasher: passhasher,
		speller:    speller,
	}
}

// @title Orders API
// @version 1.0
// @description This is a sample service for managing orders
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email soberkoder@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8081
// @BasePath /
func (h Handler) InitRouter() http.Handler {
	r := http.NewServeMux()

	r.Handle("/", h.Logging(http.HandlerFunc(h.MainMenu)))
	r.Handle("/signin", h.Logging(http.HandlerFunc(h.Login)))
	r.Handle("/signup", h.Logging(http.HandlerFunc(h.Register)))
	r.Handle("/refresh", h.Logging(http.HandlerFunc(h.Refresh)))

	r.HandleFunc("/swagger", h.swaggerUI)
	r.HandleFunc("/public/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))).ServeHTTP(w, r)
	})

	return r
}

func (h Handler) swaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fp := "./templates/swagger.html"
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		h.log.Errorf("Can't parse file: %v", err)
		return
	}

	err = tmpl.Execute(w, struct {
		Time int64
	}{
		Time: time.Now().Unix(),
	})
	if err != nil {
		h.log.Errorf("Can't execute swagger.html: %v", err)
		return
	}
}

func (h Handler) MainMenu(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) > 5 {
		h.log.Infof("Attempt to access non-existent path: %v | %v", r.Method, r.URL.Path)
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}
	result := make([]string, 0, len(path))
	for _, v := range path {
		if v != "" {
			result = append(result, v)
		}
	}

	if len(result) != 2 {
		h.log.Infof("Attempt to access non-existent path: %v | %v", r.Method, r.URL.Path)
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), ContextKeyUsername, result[0]))
	err := h.Auth(r)
	if err != nil {
		h.log.Errorf("Authentication error: %v", err)
		http.Error(w, "Auth error", http.StatusUnauthorized)
		return
	}

	if result[1] == "note" && r.Method == http.MethodPost {
		h.NotePost(w, r)
		return
	} else if result[1] == "notes" {
		h.AllNotes(w, r)
		return
	} else {
		h.log.Infof("Attempt to access non-existent path: %v | %v", r.Method, r.URL.Path)
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}
}

func (h Handler) Auth(r *http.Request) error {
	token, err := GetToken(r)
	if err != nil {
		return err
	}

	claims, err := h.tokens.ParseToken(token, tokenservice.AccessToken)

	val := r.Context().Value(ContextKeyUsername)
	username := val.(string)

	if username != claims.Username {
		return fmt.Errorf("incorrect auth token")
	}

	if err != nil {
		return err
	}

	_, err = h.users.GetUser(claims.Username)
	if err != nil {
		return err
	}

	return nil
}

func GetToken(r *http.Request) (string, error) {
	var token string
	cookie, err := r.Cookie("token")
	if err != nil {
		if err != http.ErrNoCookie {
			return "", fmt.Errorf("cant parse cookie")
		}
	} else {
		token = cookie.Value
	}

	if token == "" {
		t := r.Header.Get("Authorization")
		val := strings.Split(t, " ")
		if len(val) != 2 || val[0] != "Bearer" {
			return "", fmt.Errorf("incorrect auth format")
		} else {
			token = val[1]
		}
	}

	if token == "" {
		return "", fmt.Errorf("can't find auth token")
	}

	return token, nil
}

type customWriter struct {
	http.ResponseWriter
	code int
}

func (w *customWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (h Handler) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &customWriter{w, http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)
		h.log.Infof("%s %s %s Status %d", r.Method, r.RequestURI, time.Since(start), wrappedWriter.code)
	})
}
