package book

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"perpus_backend/types"
	"testing"

	"github.com/gorilla/mux"
)

func TestHandlerBook(t *testing.T) {
	mockBookStore := &types.MockBookStore{}
	mockUserStore := &types.MockUserStore{}
	h := NewHandler(mockBookStore, mockUserStore)

	t.Run("it should get books", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/books", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/books", h.handleGetBooks).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("it should get book by ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/books/6918315b-dff4-8324-969f-e43cd434eb3e", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/books/{bookID}", h.handleGetBookByID).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("it should make a book", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		payload := types.SetPayloadBook{
			JudulBuku: "wleee",
			Penulis:   "si itu",
			Pengarang: "si ini",
			Tahun:     "2025",
		}

		writer.WriteField("judul_buku", payload.JudulBuku)
		writer.WriteField("penulis", payload.Penulis)
		writer.WriteField("pengarang", payload.Pengarang)
		writer.WriteField("tahun", payload.Tahun)

		img, err := writer.CreateFormFile("cover_buku", "test.jpg")
		if err != nil {
			t.Fatal(err)
		}

		img.Write([]byte("fake img file"))

		pdf, err := writer.CreateFormFile("buku_pdf", "test.pdf")
		if err != nil {
			t.Fatal(err)
		}

		pdf.Write([]byte("fake pdf file"))

		writer.Close()

		req, err := http.NewRequest(http.MethodPost, "/books", body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/books", h.handleCreateBook).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
