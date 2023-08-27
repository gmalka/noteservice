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
	"noteservice/service/yandexspellerservice"
	mymock "noteservice/transport/rest/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

func TestHandler_AllNotes(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		token            string
		initTokensMock   func(*mymock.MockTokenManager)
		initNotesMock    func(*mymock.MockNoteService)
		initUsersMock    func(*mymock.MockUserService)
		checkReturn      func(t *testing.T, recoder *httptest.ResponseRecorder)
		wantedStatusCode int
	}{
		{
			name:  "/${username}/notes AllNotes OK",
			path:  "/gmalka/notes",
			token: "111",
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initNotesMock: func(mns *mymock.MockNoteService) {
				mns.EXPECT().GetAllNotes("gmalka").Times(1).Return([]model.Note{{Text: "first note", Username: "gmalka"}, {Text: "second note", Username: "gmalka"}}, nil)
			},
			checkReturn: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				b, err := io.ReadAll(recoder.Body)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				result := make([]string, 0)
				err = json.Unmarshal(b, &result)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				wanted := []string{"first note", "second note"}
				if len(wanted) != len(result) {
					t.Errorf("Expected %v, but got %v", wanted, result)
				}

				for k, v := range wanted {
					if v != result[k] {
						t.Errorf("Expected %v, but got %v", wanted, result)
					}
				}
			},
			wantedStatusCode: http.StatusOK,
		},
		{
			name:  "/${username}/notes AllNotes Wrong username in token",
			path:  "/alka/notes",
			token: "111",
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock:    func(mus *mymock.MockUserService) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes AllNotes error on ParseToken",
			path:  "/gmalka/notes",
			token: "111",
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, errors.New("some error"))
			},
			initUsersMock:    func(mus *mymock.MockUserService) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes AllNotes error on GetUser while auth",
			path:  "/gmalka/notes",
			token: "111",
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, errors.New("some error"))
			},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes AllNotes error on GetAllNotes",
			path:  "/gmalka/notes",
			token: "111",
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initNotesMock: func(mns *mymock.MockNoteService) {
				mns.EXPECT().GetAllNotes("gmalka").Times(1).Return(nil, errors.New("some error"))
			},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
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
			notes := mymock.NewMockNoteService(ctrl)
			users := mymock.NewMockUserService(ctrl)

			tt.initTokensMock(tokens)
			tt.initNotesMock(notes)
			tt.initUsersMock(users)

			h := Handler{
				log:    log,
				users:  users,
				notes:  notes,
				tokens: tokens,
			}

			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, tt.path, nil)
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			request.Header.Set("Authorization", "Bearer "+tt.token)

			router := h.InitRouter()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantedStatusCode {
				t.Errorf("Expected %d, but got %d", tt.wantedStatusCode, recorder.Code)
			}

			tt.checkReturn(t, recorder)
		})
	}
}

func TestHandler_NotePost(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		token            string
		input            func() []byte
		initTokensMock   func(*mymock.MockTokenManager)
		initSpellerMock  func(*mymock.MockSpeller)
		initNotesMock    func(*mymock.MockNoteService)
		initUsersMock    func(*mymock.MockUserService)
		checkReturn      func(t *testing.T, recoder *httptest.ResponseRecorder)
		wantedStatusCode int
	}{
		{
			name:  "/${username}/notes NotePost OK",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initSpellerMock: func(ms *mymock.MockSpeller) {
				ms.EXPECT().CheckText("some text").Times(1).Return([]byte("some text"), nil)
			},
			initNotesMock: func(mns *mymock.MockNoteService) {
				mns.EXPECT().CreateNote(model.Note{Text: "some text", Username: "gmalka"}).Times(1).Return(nil)
			},
			checkReturn: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				b, err := io.ReadAll(recoder.Body)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				wanted := []byte("some text")
				if len(b) != len(wanted) {
					t.Errorf("Expected %v, but got %v", []byte("some text"), b)
				}
				for k, v := range b {
					if v != wanted[k] {
						t.Errorf("Expected %v, but got %v", []byte("some text"), b)
					}
				}
			},
			wantedStatusCode: http.StatusOK,
		},
		{
			name:  "/${username}/notes NotePost Wrong username in token",
			path:  "/alka/note",
			token: "111",
			input: func() []byte { return nil },
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock:    func(mus *mymock.MockUserService) {},
			initSpellerMock:  func(ms *mymock.MockSpeller) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes NotePost error on ParseToken",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{}, errors.New("some errpr"))
			},
			initUsersMock:    func(mus *mymock.MockUserService) {},
			initSpellerMock:  func(ms *mymock.MockSpeller) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes NotePost error on GetUser",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, errors.New("some errpr"))
			},
			initSpellerMock:  func(ms *mymock.MockSpeller) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnauthorized,
		},
		{
			name:  "/${username}/notes NotePost error on CheckText",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initSpellerMock: func(ms *mymock.MockSpeller) {
				ms.EXPECT().CheckText("some text").Times(1).Return(nil, errors.New("some text"))
			},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusInternalServerError,
		},
		{
			name:  "/${username}/notes NotePost error on CheckText, founded syntax errors",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initSpellerMock: func(ms *mymock.MockSpeller) {
				ms.EXPECT().CheckText("some text").Times(1).Return(nil, yandexspellerservice.ErrSyntax)
			},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
			wantedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:  "/${username}/notes NotePost error on CreateNote",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				text := model.Input{Text: "some text"}

				b, err := json.Marshal(text)
				if err != nil {
					t.Errorf("Expected no error, but got %b", err)
				}

				return b
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initSpellerMock: func(ms *mymock.MockSpeller) {
				ms.EXPECT().CheckText("some text").Times(1).Return([]byte("some text"), nil)
			},
			initNotesMock: func(mns *mymock.MockNoteService) {
				mns.EXPECT().CreateNote(model.Note{Text: "some text", Username: "gmalka"}).Times(1).Return(errors.New("some error"))
			},
			checkReturn: func(t *testing.T, recoder *httptest.ResponseRecorder) {
			},
			wantedStatusCode: http.StatusInternalServerError,
		},
		{
			name:  "/${username}/notes NotePost empty input",
			path:  "/gmalka/note",
			token: "111",
			input: func() []byte {
				return []byte{}
			},
			initTokensMock: func(mtm *mymock.MockTokenManager) {
				mtm.EXPECT().ParseToken("111", 0).Times(1).Return(model.UserClaims{Username: "gmalka"}, nil)
			},
			initUsersMock: func(mus *mymock.MockUserService) {
				mus.EXPECT().GetUser("gmalka").Times(1).Return(model.User{}, nil)
			},
			initSpellerMock:  func(ms *mymock.MockSpeller) {},
			initNotesMock:    func(mns *mymock.MockNoteService) {},
			checkReturn:      func(t *testing.T, recoder *httptest.ResponseRecorder) {},
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
			notes := mymock.NewMockNoteService(ctrl)
			users := mymock.NewMockUserService(ctrl)
			speller := mymock.NewMockSpeller(ctrl)

			tt.initTokensMock(tokens)
			tt.initNotesMock(notes)
			tt.initUsersMock(users)
			tt.initSpellerMock(speller)

			h := Handler{
				log:     log,
				users:   users,
				notes:   notes,
				tokens:  tokens,
				speller: speller,
			}

			recorder := httptest.NewRecorder()

			b := tt.input()

			request, err := http.NewRequest(http.MethodPost, tt.path, bytes.NewReader(b))
			if err != nil {
				t.Errorf("Expected no error, but got %b", err)
			}

			request.Header.Set("Authorization", "Bearer "+tt.token)

			router := h.InitRouter()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantedStatusCode {
				t.Errorf("Expected %d, but got %d", tt.wantedStatusCode, recorder.Code)
			}

			tt.checkReturn(t, recorder)
		})
	}
}
