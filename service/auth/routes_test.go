package auth

import (
	"bytes"
	"mime/multipart"
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

	h := NewHandler(jwt, userStore)

	t.Run("it should fail register, because use wrong email format", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		payload := types.SetPayloadUser{
			Name:     "admin",
			Email:    "asd",
			Password: "asd",
		}

		writer.WriteField("name", payload.Name)
		writer.WriteField("email", payload.Email)
		writer.WriteField("password", payload.Password)

		file, err := writer.CreateFormFile("avatar", "test.png")
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

		r.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for check error in body

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status code %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("it should register an user", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		payload := types.SetPayloadUser{
			Name:     "admin",
			Email:    "admin@gmail.com",
			Password: "admin12345",
		}

		writer.WriteField("name", payload.Name)
		writer.WriteField("email", payload.Email)
		writer.WriteField("password", payload.Password)
		writer.WriteField("avatar", "-")

		// file, err := writer.CreateFormFile("avatar", "alah.png")
		// if err != nil {
		// 	t.Fatal(err)
		// }

		// file.Write([]byte("fake img content"))

		writer.Close()

		req, err := http.NewRequest(http.MethodPost, "/register", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

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
