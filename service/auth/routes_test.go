package auth

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"perpus_backend/types"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestAuthHandler(t *testing.T) {
	userStore := &mockUserStore{}
	handler := NewHandler(userStore)

	t.Run("it should register an user", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("name", "admin")
		writer.WriteField("email", "admin@admin.com")
		writer.WriteField("password", "admin12345")

		file, err := writer.CreateFormFile("avatar", "test.jpg")
		if err != nil {
			t.Fatal(err)
		}

		file.Write([]byte("fake img content"))

		writer.Close()

		req, err := http.NewRequest(http.MethodPost, "/register", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/register", handler.handleRegister).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("it should login", func(t *testing.T) {
		payload := url.Values{}
		payload.Add("email", "admin@admin.com")
		payload.Add("password", "admin12345")

		req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(payload.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/login", handler.handleLogin).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}

type mockUserStore struct{}

func (m *mockUserStore) GetUsers() ([]*types.User, error) {
	return nil, nil
}

func (m *mockUserStore) GetUserWithRolesByID(id string) (*types.User, error) {
	return nil, nil
}

func (m *mockUserStore) GetUserWithRolesByEmail(email string) (*types.User, error) {
	return nil, nil
}

func (m *mockUserStore) CreateUser(*types.User) error {
	return nil
}

func (m *mockUserStore) UpdateUser(id string, u *types.User) error {
	return nil
}

func (m *mockUserStore) DeleteUser(id string) error {
	return nil
}

func (m *mockUserStore) IncrementTokenVersion(id string) error {
	return nil
}
