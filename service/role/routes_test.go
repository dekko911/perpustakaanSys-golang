package role

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"perpus_backend/types"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestHandlerRole(t *testing.T) {
	mockRoleStore := types.MockRoleStore{}
	mockUserStore := types.MockUserStore{}
	h := NewHandler(mockRoleStore, mockUserStore)

	t.Run("it should be get roles data", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/roles", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/roles", h.handleGetRoles).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug if error

		if w.Code != cok {
			t.Errorf("expected status code: %d, got %d", cok, w.Code)
		}
	})

	t.Run("it should be get role by id", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/roles/6918315b-dff4-8324-969f-e43cd434eb3e", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/roles/{roleID}", h.handleGetRoleByID).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug if error

		if w.Code != cok {
			t.Errorf("expected status code: %d, got %d", cok, w.Code)
		}
	})

	t.Run("it should correct and make role", func(t *testing.T) {
		form := &url.Values{}
		payload := types.SetPayloadRole{
			Name: "admin",
		}

		form.Add("name", payload.Name)

		req, err := http.NewRequest(http.MethodPost, "/roles", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/roles", h.handleCreateRole).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
