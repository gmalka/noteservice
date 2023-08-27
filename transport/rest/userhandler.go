package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"noteservice/model"
	"noteservice/service/tokenservice"
	"time"
)

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var err error
	defer r.Body.Close()

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

	user := model.User{}
	json.Unmarshal(result, &user)
	if user.Username == "" || user.Password == "" {
		h.log.Infof("Register with empty username or password: %s", user.Username)
		http.Error(w, "incorrect input data error", http.StatusBadRequest)
		return
	}

	user.Password, err = h.passhasher.HashPassword(user.Password)
	if err != nil {
		h.log.Errorf("Can't hash password: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	err = h.users.CreateUser(user)
	if err != nil {
		h.log.Errorf("Can't create user: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	message := model.Message{
		Message: "Success Registration, " + user.Username,
	}

	b, err := json.Marshal(message)
	if err != nil {
		h.log.Errorf("Can't marshal data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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

	user := model.User{}
	json.Unmarshal(result, &user)
	if user.Username == "" || user.Password == "" {
		h.log.Infof("Login with empty username or password: %s", user.Username)
		http.Error(w, "incorrect input data error", http.StatusBadRequest)
		return
	}

	wantedUser, err := h.users.GetUser(user.Username)
	if err != nil {
		h.log.Errorf("Can't get user: %v", err)
		http.Error(w, "incorrect password or login", http.StatusForbidden)
		return
	}

	err = h.passhasher.CheckPassword(user.Password, wantedUser.Password)
	if err != nil {
		h.log.Errorf("Can't check password: %v", err)
		http.Error(w, "incorrect password or login", http.StatusForbidden)
		return
	}

	tokens := model.Tokens{}

	tokens.AccessToken, err = h.tokens.CreateToken(model.UserInfo{
		Username: user.Username,
	}, tokenservice.AccessToken)
	if err != nil {
		h.log.Errorf("Can't create access token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	tokens.RefreshToken, err = h.tokens.CreateToken(model.UserInfo{
		Username: user.Username,
	}, tokenservice.RefreshToken)
	if err != nil {
		h.log.Errorf("Can't create refresh token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    tokens.AccessToken,
		Path:     "/" + user.Username,
		HttpOnly: true,
		Expires:  time.Now().Add(tokenservice.ACCESS_TOKEN_TTL * time.Minute),
	}
	http.SetCookie(w, &cookie)

	b, err := json.Marshal(tokens)
	if err != nil {
		h.log.Errorf("Can't marshal data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Errorf("Can't read data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	token := model.Refresh{}
	json.Unmarshal(b, &token)

	claims, err := h.tokens.ParseToken(token.RefreshToken, tokenservice.RefreshToken)
	if err != nil {
		h.log.Errorf("Can't parse token: %v", err)
		http.Error(w, "token error", http.StatusUnauthorized)
		return
	}

	tokens := model.Tokens{}

	tokens.AccessToken, err = h.tokens.CreateToken(model.UserInfo{
		Username: claims.Username,
	}, tokenservice.AccessToken)
	if err != nil {
		h.log.Errorf("Can't create access token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	tokens.RefreshToken, err = h.tokens.CreateToken(model.UserInfo{
		Username: claims.Username,
	}, tokenservice.RefreshToken)
	if err != nil {
		h.log.Errorf("Can't create refresh token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    tokens.AccessToken,
		Path:     "/" + claims.Username,
		HttpOnly: true,
		Expires:  time.Now().Add(tokenservice.ACCESS_TOKEN_TTL * time.Minute),
	}
	http.SetCookie(w, &cookie)

	b, err = json.Marshal(tokens)
	if err != nil {
		h.log.Errorf("Can't marshal data: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
