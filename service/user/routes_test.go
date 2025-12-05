package user

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"perpus_backend/types"
	"testing"

	"github.com/gorilla/mux"
)

func TestHandlerUser(t *testing.T) {
	us := &types.MockUserStore{}
	h := NewHandler(us)

	t.Run("should get users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users", h.handleGetUsers).Methods(http.MethodGet)

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("it should get user by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/6918315b-dff4-8324-969f-e43cd434eb3e", nil)

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users/{userID}", h.handleGetUserWithRolesByID).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // check the error

		if w.Code != cok {
			t.Errorf("expected status code %d, got %d", cok, w.Code)
		}
	})

	t.Run("should be created user", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		payload := types.SetPayloadUser{
			Name:     "miko",
			Email:    "miko@gmail.com",
			Password: "miko12345",
		}

		writer.WriteField("name", payload.Name)
		writer.WriteField("email", payload.Email)
		writer.WriteField("password", payload.Password)

		file, err := writer.CreateFormFile("avatar", "test.jpeg")
		if err != nil {
			t.Fatal(err)
		}

		file.Write([]byte("fake img content"))

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users", h.handleCreateUser).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
