package user

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"perpus_backend/types"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetUsers(t *testing.T) {
	userStore := &mockUserStore{}
	handler := NewHandler(userStore)

	t.Run("should get users", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/users", nil)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users", handler.handleGetUsers).Methods(http.MethodGet)

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("should get user, because there is invalid param ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/users/d21e510f-7e9b-4120-b0f3-b8f4bbf15f2b", nil)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/users/{userID}", handler.handleGetUserWithRolesByID).Methods(http.MethodGet)

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
	return &types.User{}, fmt.Errorf("user not found")
}

func (m *mockUserStore) GetUserWithRolesByEmail(email string) (*types.User, error) {
	return &types.User{}, fmt.Errorf("user not found")
}

func (m *mockUserStore) CreateUser(*types.User) error {
	return fmt.Errorf("can't create an user")
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
