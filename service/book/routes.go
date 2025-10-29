package book

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/middleware"
	"perpus_backend/types"
	"perpus_backend/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Handler struct {
	store     types.BookStore
	userStore types.UserStore
}

const (
	COK = http.StatusOK
	OK  = "OK"
)

func NewHandler(s types.BookStore, us types.UserStore) *Handler {
	return &Handler{
		store:     s,
		userStore: us,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/books", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin", "staff", "user")(h.handleGetBooks), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/books/{bookID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin", "staff", "user")(h.handleGetBookByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/books", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin", "staff")(h.handleCreateBook), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/books/{bookID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin", "staff")(h.handleUpdateBook), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/books/{bookID}", middleware.AuthWithJWTToken(middleware.NeededRole(h.userStore, "admin", "staff")(h.handleDeleteBook), h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.store.GetBooks()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   books,
		Status: OK,
	})
}

func (h *Handler) handleGetBookByID(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	book, err := h.store.GetBookByID(bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:   COK,
		Data:   book,
		Status: OK,
	})
}

func (h *Handler) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var (
		payload types.PayloadBook

		fileName string
	)

	if err := r.ParseMultipartForm(8 << 20); err != nil { // 20 = 2 dikalikan sebanyak 20 kali.
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadBook{
		IdBuku:    r.FormValue("id_buku"),
		JudulBuku: r.FormValue("judul_buku"),
		Penulis:   r.FormValue("penulis"),
		Pengarang: r.FormValue("pengarang"),
		Tahun:     r.FormValue("tahun"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetBookByIDBuku(payload.IdBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("id_buku: %s is already exists", payload.IdBuku))
		return
	}

	if _, err := h.store.GetBookByJudulBuku(payload.JudulBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("judul_buku: %s is already exists", payload.JudulBuku))
		return
	}

	file, header, err := r.FormFile("cover_buku")

	if err == http.ErrMissingFile {
		fileName = "-"
	}

	if err == nil {
		defer file.Close()

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		fileName = randomString + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	tahun, err := strconv.Atoi(payload.Tahun)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.CreateBook(&types.Book{
		IdBuku:    payload.IdBuku,
		JudulBuku: payload.JudulBuku,
		CoverBuku: fileName,
		Penulis:   payload.Penulis,
		Pengarang: payload.Pengarang,
		Tahun:     tahun,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "Book Created!",
	})
}

func (h *Handler) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	var (
		payload types.PayloadUpdateBook

		fileName string
	)

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	if err := r.ParseMultipartForm(8 << 20); err != nil { // 20 = 2 dikalikan sebanyak 20 kali.
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	payload = types.PayloadUpdateBook{
		IdBuku:    r.FormValue("id_buku"),
		JudulBuku: r.FormValue("judul_buku"),
		Penulis:   r.FormValue("penulis"),
		Pengarang: r.FormValue("pengarang"),
		Tahun:     r.FormValue("tahun"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	b, err := h.store.GetBookByID(payload.IdBuku)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("id_buku: %s is already exists", payload.IdBuku))
		return
	}

	file, header, err := r.FormFile("cover_buku")

	if err == http.ErrMissingFile {
		fileName = b.CoverBuku
	}

	if err == nil {
		defer file.Close()

		fileImg := "./assets/public/images" + fileName
		if err := os.Remove(fileImg); err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, err)
			return
		}

		randomString := xid.New().String()

		ext := filepath.Ext(header.Filename)
		fileName = randomString + ext

		dst, _ := os.Create("./assets/public/images/" + fileName)
		defer dst.Close()

		io.Copy(dst, file)
	}

	tahun, _ := strconv.Atoi(payload.Tahun)

	if payload.IdBuku != "" {
		b.IdBuku = payload.IdBuku
	}
	if payload.JudulBuku != "" {
		b.JudulBuku = payload.JudulBuku
	}
	if payload.Penulis != "" {
		b.Penulis = payload.Penulis
	}
	if payload.Pengarang != "" {
		b.Pengarang = payload.Pengarang
	}
	if payload.Tahun != "" {
		b.Tahun = tahun
	}

	if err := h.store.UpdateBook(bookID, &types.Book{
		IdBuku:    b.IdBuku,
		JudulBuku: b.JudulBuku,
		CoverBuku: fileName,
		Penulis:   b.Penulis,
		Pengarang: b.Pengarang,
		Tahun:     tahun,
	}); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "Book Updated!",
	})
}

func (h *Handler) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	if err := h.store.DeleteBook(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "Book Deleted!",
	})
}
