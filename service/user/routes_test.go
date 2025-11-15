package user

import (
	"net/http"
	"net/http/httptest"
	"perpus_backend/types"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetUsers(tst *testing.T) {
	userStore := &mockUserStore{}
	handler := NewHandler(userStore)

	tst.Run("should get users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users", handler.handleGetUsers).Methods(http.MethodGet)

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	tst.Run("it should get user with param ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/hah", nil)

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users/{userID}", handler.handleGetUserWithRolesByID).Methods(http.MethodGet)

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// t.Run("should be created user", func(t *testing.T) {
	// 	req := httptest.NewRequest(http.MethodPost, "/users", ?)
	// })
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
