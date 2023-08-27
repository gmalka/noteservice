package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"noteservice/model"
	mymock "noteservice/transport/rest/mock"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name             string
		user             model.User
		initHasherMock   func(*mymock.MockPasswordHasher)
		initUsersMock    func(*mymock.MockUserService)
		wantedStatusCode int
	}{
		{
			name: "/signup: OK",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().HashPassword("123").Times(1).Return("123", nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().CreateUser(model.User{Username: "gmalka", Password: "123"}).Times(1).Return(nil)
			},
			wantedStatusCode: http.StatusOK,
		},
		{
			name: "/signup: Bad input data(empty username)",
			user: model.User{
				Username: "",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
			},
			initUsersMock: func(mus *mymock.MockUserService) {
			},
			wantedStatusCode: http.StatusBadRequest,
		},
		{
			name: "/signup: Bad input data(empty username)",
			user: model.User{
				Username: "gmalka",
				Password: "",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
			},
			initUsersMock: func(mus *mymock.MockUserService) {
			},
			wantedStatusCode: http.StatusBadRequest,
		},
		{
			name: "/signup: Error on PasswordHasher function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().HashPassword("123").Times(1).Return("", errors.New("some error"))
			},
			initUsersMock: func(mus *mymock.MockUserService) {
			},
			wantedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "/signup: Error on CreateUser function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().HashPassword("123").Times(1).Return("123", nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().CreateUser(model.User{Username: "gmalka", Password: "123"}).Times(1).Return(errors.New("some error"))
			},
			wantedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			log.SetOutput(ioutil.Discard)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			hasher := mymock.NewMockPasswordHasher(ctrl)
			users := mymock.NewMockUserService(ctrl)

			tt.initHasherMock(hasher)
			tt.initUsersMock(users)

			h := Handler{
				log:        log,
				users:      users,
				passhasher: hasher,
			}

			val, err := json.Marshal(tt.user)
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(val))
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			router := h.InitRouter()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantedStatusCode {
				t.Errorf("Expected %d, but got %d", tt.wantedStatusCode, recorder.Code)
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name             string
		user             model.User
		initHasherMock   func(*mymock.MockPasswordHasher)
		initUsersMock    func(*mymock.MockUserService)
		initTokensMock   func(*mymock.MockTokenManager)
		checkCookie      func(t *testing.T, recoder *httptest.ResponseRecorder)
		checkTokens      func(t *testing.T, recoder *httptest.ResponseRecorder)
		wantedStatusCode int
	}{
		{
			name: "/signin: OK",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().CheckPassword("123", "1234").Times(1).Return(nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{Username: "gmalka", Password: "1234"}, nil)
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("access_token", nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 1).Times(1).Return("refresh_token", nil)
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				result := strings.Split(cookie, " ")

				if len(result) < 2 {
					t.Errorf("Unespected cookie %v: ", cookie)
				}

				if result[0] != "token=access_token;" {
					t.Errorf("Expected token=access_token;, but got %v: ", result[0])
				}

				if result[1] != "Path=/gmalka;" {
					t.Errorf("Expected Path=/gmalka;, but got %v: ", result[1])
				}
			},
			checkTokens: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				b, err := io.ReadAll(recoder.Body)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				tokens := model.Tokens{}
				json.Unmarshal(b, &tokens)

				if tokens.AccessToken != "access_token" {
					t.Errorf("Expected access_token, but got %s", tokens.AccessToken)
				}

				if tokens.RefreshToken != "refresh_token" {
					t.Errorf("Expected refresh_token, but got %s", tokens.RefreshToken)
				}
			},
			wantedStatusCode: http.StatusOK,
		},
		{
			name: "/signin: Bad input data(empty password)",
			user: model.User{
				Username: "gmalka",
				Password: "",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
			},
			initUsersMock: func(mus *mymock.MockUserService) {
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens: func(t *testing.T, recoder *httptest.ResponseRecorder) {
			},
			wantedStatusCode: http.StatusBadRequest,
		},
		{
			name: "/signin: Bad input data(empty username)",
			user: model.User{
				Username: "",
				Password: "123",
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
			},
			initUsersMock: func(mus *mymock.MockUserService) {
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens: func(t *testing.T, recoder *httptest.ResponseRecorder) {
			},
			wantedStatusCode: http.StatusBadRequest,
		},
		{
			name: "/signin: Error on GetUser function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, errors.New("some error"))
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {},
			initTokensMock: func(mtm *mymock.MockTokenManager) {},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusForbidden,
		},
		{
			name: "/signin: Error on CheckPassword function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{Username: "gmalka", Password: "1234"}, nil)
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().CheckPassword("123", "1234").Times(1).Return(errors.New("some error"))
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusForbidden,
		},
		{
			name: "/signin: Error on Access token generate function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{Username: "gmalka", Password: "1234"}, nil)
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().CheckPassword("123", "1234").Times(1).Return(nil)
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("", errors.New("some error"))
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "/signin: Error on Refresh token generate function",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{Username: "gmalka", Password: "1234"}, nil)
			},
			initHasherMock: func(mph *mymock.MockPasswordHasher) {
				mph.EXPECT().CheckPassword("123", "1234").Times(1).Return(nil)
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("access_token", nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 1).Times(1).Return("", errors.New("some error"))
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				if cookie != "" {
					t.Errorf("Expected nothing, but got %v", cookie)
				}
			},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			log.SetOutput(ioutil.Discard)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			hasher := mymock.NewMockPasswordHasher(ctrl)
			users := mymock.NewMockUserService(ctrl)
			tokens := mymock.NewMockTokenManager(ctrl)

			tt.initHasherMock(hasher)
			tt.initUsersMock(users)
			tt.initTokensMock(tokens)

			h := Handler{
				log:        log,
				users:      users,
				passhasher: hasher,
				tokens:     tokens,
			}

			val, err := json.Marshal(tt.user)
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/signin", bytes.NewReader(val))
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			router := h.InitRouter()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantedStatusCode {
				t.Errorf("Expected %d, but got %d", tt.wantedStatusCode, recorder.Code)
			}

			tt.checkCookie(t, recorder)
			tt.checkTokens(t, recorder)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	tests := []struct {
		name             string
		user             model.User
		getToken         func() []byte
		initTokensMock   func(*mymock.MockTokenManager)
		checkCookie      func(t *testing.T, recoder *httptest.ResponseRecorder)
		checkTokens      func(t *testing.T, recoder *httptest.ResponseRecorder)
		wantedStatusCode int
	}{
		{
			name: "/refresh: OK",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			getToken: func() []byte {
				refresh := model.Refresh{
					RefreshToken: "123",
				}

				b, err := json.Marshal(refresh)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("123", 1).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("access_token", nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 1).Times(1).Return("refresh_token", nil)
			},
			checkCookie: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				cookie := recoder.Header().Get("Set-Cookie")

				result := strings.Split(cookie, " ")

				if len(result) < 2 {
					t.Errorf("Unespected cookie %v: ", cookie)
				}

				if result[0] != "token=access_token;" {
					t.Errorf("Expected token=access_token;, but got %v: ", result[0])
				}

				if result[1] != "Path=/gmalka;" {
					t.Errorf("Expected Path=/gmalka;, but got %v: ", result[1])
				}
			},
			checkTokens: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				b, err := io.ReadAll(recoder.Body)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				tokens := model.Tokens{}
				json.Unmarshal(b, &tokens)

				if tokens.AccessToken != "access_token" {
					t.Errorf("Expected access_token, but got %s", tokens.AccessToken)
				}

				if tokens.RefreshToken != "refresh_token" {
					t.Errorf("Expected refresh_token, but got %s", tokens.RefreshToken)
				}
			},
			wantedStatusCode: http.StatusOK,
		},
		{
			name: "/refresh: error on ParseToken",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			getToken: func() []byte {
				refresh := model.Refresh{
					RefreshToken: "123",
				}

				b, err := json.Marshal(refresh)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("123", 1).Times(1).Return(model.UserClaims{Username: "gmalka"}, errors.New("some error"))
			},
			checkCookie:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "/refresh: error on Access generate",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			getToken: func() []byte {
				refresh := model.Refresh{
					RefreshToken: "123",
				}

				b, err := json.Marshal(refresh)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("123", 1).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("access_token", errors.New("some error"))
			},
			checkCookie:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "/refresh: error on Refresh generate",
			user: model.User{
				Username: "gmalka",
				Password: "123",
			},
			getToken: func() []byte {
				refresh := model.Refresh{
					RefreshToken: "123",
				}

				b, err := json.Marshal(refresh)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("123", 1).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 0).Times(1).Return("access_token", nil)
				mtm.EXPECT().CreateToken(model.UserInfo{Username: "gmalka"}, 1).Times(1).Return("refresh_token", errors.New("some error"))
			},
			checkCookie:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			checkTokens:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			log.SetOutput(ioutil.Discard)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokens := mymock.NewMockTokenManager(ctrl)

			tt.initTokensMock(tokens)

			h := Handler{
				log:    log,
				tokens: tokens,
			}

			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(tt.getToken()))
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			router := h.InitRouter()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantedStatusCode {
				t.Errorf("Expected %d, but got %d", tt.wantedStatusCode, recorder.Code)
			}

			tt.checkCookie(t, recorder)
			tt.checkTokens(t, recorder)
		})
	}
}
