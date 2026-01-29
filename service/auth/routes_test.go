package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"

	"github.com/gorilla/mux"
)

func TestAuthHandler(t *testing.T) {
	jwt := &jwt.AuthJWT{}
	userStore := &types.MockUserStore{}

	// TODO: UBAH METHOD FORM NYA MENJADI JSON
	h := NewHandler(jwt, userStore)

	t.Run("it should fail register, because use wrong email format", func(t *testing.T) {
		payload := types.SetPayloadJSONRegister{
			Name:     "admin",
			Email:    "asd",
			Password: "asd",
		}

		marshalled, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for check error in body

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status code %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("it should register an user", func(t *testing.T) {
		payload := types.SetPayloadJSONRegister{
			Name:     "admin",
			Email:    "admin@gmail.com",
			Password: "admin12345",
		}

		marshalled, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for check error in body

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
