package member

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

func TestHandlerMember(t *testing.T) {
	jwt := &jwt.AuthJWT{}
	mockMemberStore := &types.MockMemberStore{}
	mockUserStore := &types.MockUserStore{}

	h := NewHandler(jwt, mockMemberStore, mockUserStore)

	t.Run("it should be get members", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/members", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/members", h.handleGetMembers).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != cok {
			t.Errorf("expected status code %d, got %d", cok, w.Code)
		}
	})

	t.Run("it should be get member by id", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/members/6918315b-dff4-8324-969f-e43cd434eb3e", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/members/{memberID}", h.handleGetMemberByID).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != cok {
			t.Errorf("expected status code %d, got %d", cok, w.Code)
		}
	})

	t.Run("it should make member", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		payload := types.SetPayloadMember{
			Nama:         "admin",
			JenisKelamin: "L",
			Kelas:        "8i",
			NoTelepon:    "084 428 540 584",
		}

		writer.WriteField("nama", payload.Nama)
		writer.WriteField("jenis_kelamin", payload.JenisKelamin)
		writer.WriteField("kelas", payload.Kelas)
		writer.WriteField("no_telepon", payload.NoTelepon)

		file, err := writer.CreateFormFile("profil", "test.jpeg")
		if err != nil {
			t.Fatal(err)
		}

		file.Write([]byte("fake img content"))

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/members", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/members", h.handleCreateMember).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
