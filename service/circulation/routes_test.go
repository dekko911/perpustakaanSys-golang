package circulation

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"

	"github.com/gorilla/mux"
)

func TestHandlerCirculation(t *testing.T) {
	jwt := &jwt.AuthJWT{}
	mockCirculationStore := &types.MockCirculationStore{}
	mockUserStore := &types.MockUserStore{}

	h := NewHandler(jwt, mockCirculationStore, mockUserStore)

	t.Run("it should get circulations", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/circulations", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/circulations", h.handleGetCirculations).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("it should be get member by ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/circulations/6918315b-dff4-8324-969f-e43cd434eb3e", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/circulations/{cID}", h.handleGetCirculationByID).Methods(http.MethodGet)
		r.ServeHTTP(w, req)

		// t.Log(w.Body)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("it should create a circulation", func(t *testing.T) {
		form := url.Values{}
		payload := types.SetPayloadCirculation{
			BukuID:        "6918315b-dff4-8324-969f-e43cd434eb3e",
			Peminjam:      "miko",
			TanggalPinjam: "2025-12-02",
			JatuhTempo:    "2025-12-12",
			Denda:         "10000",
		}

		form.Add("buku_id", payload.BukuID)
		form.Add("peminjam", payload.Peminjam)
		form.Add("tanggal_pinjam", payload.TanggalPinjam)
		form.Add("jatuh_tempo", payload.JatuhTempo)
		form.Add("denda", payload.Denda)

		req, err := http.NewRequest(http.MethodPost, "/circulations", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		r := mux.NewRouter()

		r.HandleFunc("/circulations", h.handleCreateCirculation).Methods(http.MethodPost)
		r.ServeHTTP(w, req)

		// t.Log(w.Body) // for debug

		if w.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
		}
	})
}
