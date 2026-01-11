package book

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	bookStore types.BookStore
	userStore types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, bs types.BookStore, us types.UserStore) *Handler {
	return &Handler{
		bookStore: bs,
		userStore: us,
		jwt:       jwt,
	}
}

const (
	cok = http.StatusOK

	publicCoverBookFilePath = "./assets/public/images/cover/"
	publicPDFBookFilePath   = "./assets/private/pdf/"

	size10MB = 10 << 20
	size8MB  = 8 << 20
	size1MB  = 1 << 20
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/books", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetBooks, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/books/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetBookByID, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/books", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateBook, "admin", "staff"))).Methods(http.MethodPost)

	r.HandleFunc("/books/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateBook, "admin", "staff"))).Methods(http.MethodPut)

	r.HandleFunc("/books/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteBook, "admin", "staff"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := utils.ParseStringToInt(r.URL.Query().Get("page"))

	books, lastPage, err := h.bookStore.GetBooksWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:     cok,
		Data:     books,
		Page:     page,
		LastPage: lastPage,
		Status:   http.StatusText(cok),
	})
}

func (h *Handler) handleGetBookByID(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	ctx := r.Context()

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	book, err := h.bookStore.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:   cok,
		Data:   book,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var filename, filePDF string

	r.Body = http.MaxBytesReader(w, r.Body, size10MB)

	if err := r.ParseMultipartForm(size10MB); err != nil { // 20 = 2 dikalikan sebanyak 20 kali.
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload := types.SetPayloadBook{
		JudulBuku: r.FormValue("judul_buku"),
		Penulis:   r.FormValue("penulis"),
		Pengarang: r.FormValue("pengarang"),
		Tahun:     r.FormValue("tahun"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	if _, err := h.bookStore.GetBookByJudulBuku(ctx, payload.JudulBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("judul_buku: %s is already exists", payload.JudulBuku))
		return
	}

	fileImg, headerImg, errImg := r.FormFile("cover_buku") // get input form file name is "cover_buku"

	fileBookPDF, headerPDF, errPDF := r.FormFile("buku_pdf") // get input form file name is "buku_pdf"

	// fill the cover book, if input form file of cover book is empty
	if errImg == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("required file cover book"))
		filename = "-"
		return
	}

	// fill the pdf file, if input form file of PDF is empty
	if errPDF == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("required file pdf"))
		filePDF = "-"
		return
	}

	if errImg == nil {
		defer fileImg.Close()

		newFilename, err := utils.SetNewFilenameImg("random", fileImg, publicCoverBookFilePath, headerImg.Filename, headerImg.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filename = newFilename // <- set the filename for cover book image
	}

	if errPDF == nil {
		defer fileBookPDF.Close()

		newFilePDF, err := utils.SetOriginalFilenamePDF(fileBookPDF, publicPDFBookFilePath, headerPDF.Filename, headerPDF.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filePDF = newFilePDF // <- set the filename PDF for PDF book
	}

	err := h.bookStore.CreateBook(ctx, &types.Book{
		JudulBuku: payload.JudulBuku,
		CoverBuku: filename,
		BukuPDF:   filePDF,
		Penulis:   payload.Penulis,
		Pengarang: payload.Pengarang,
		Tahun:     utils.ParseStringToInt(payload.Tahun),
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		// in this line, it should be exist remove file if the err was triggered
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.JsonResponse{
		Code:    http.StatusCreated,
		Message: "Book Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	ctx := r.Context()

	var filename, filePDF string

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, size10MB)

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	// TODO: buat validasi file dan buatkan function logic untuk menyimpan file ke local storage
	if err := r.ParseMultipartForm(size10MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload := types.SetPayloadUpdateBook{
		JudulBuku: r.FormValue("judul_buku"),
		Penulis:   r.FormValue("penulis"),
		Pengarang: r.FormValue("pengarang"),
		Tahun:     r.FormValue("tahun"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesia(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	b, err := h.bookStore.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
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
		b.Tahun = utils.ParseStringToInt(payload.Tahun)
	}

	fileImg, headerImg, errImg := r.FormFile("cover_buku")

	fileBookPDF, headerPDF, errPDF := r.FormFile("buku_pdf")

	if errImg == http.ErrMissingFile {
		filename = b.CoverBuku
	}

	if errPDF == http.ErrMissingFile {
		filePDF = b.BukuPDF
	}

	if errImg == nil {
		defer fileImg.Close()

		newFilename, err := utils.UpdateTheFilenameImg("random", fileImg, publicCoverBookFilePath, b.CoverBuku, headerImg.Filename, headerImg.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filename = newFilename
	}

	if errPDF == nil {
		defer fileBookPDF.Close()

		newFilePDF, err := utils.UpdateTheOriginalFilenamePDF(fileBookPDF, publicPDFBookFilePath, b.BukuPDF, headerPDF.Filename, headerPDF.Size)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filePDF = newFilePDF
	}

	err = h.bookStore.UpdateBook(ctx, bookID, &types.Book{
		JudulBuku: b.JudulBuku,
		CoverBuku: filename,
		BukuPDF:   filePDF,
		Penulis:   b.Penulis,
		Pengarang: b.Pengarang,
		Tahun:     b.Tahun,
	})
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		// it should be exist a remove file, but i don't know yet how to remove it
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Book Updated!",
		Status:  http.StatusText(cok),
	})
}

func (h *Handler) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	ctx := r.Context()

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	b, err := h.bookStore.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// file cover book
	fileImg := filepath.Join(publicCoverBookFilePath, b.CoverBuku)

	// file pdf book
	filePDF := filepath.Join(publicPDFBookFilePath, b.BukuPDF)

	if err := utils.DeleteFilepathWithFilename(fileImg, filePDF); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.bookStore.DeleteBook(ctx, bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:    cok,
		Message: "Book Deleted!",
		Status:  http.StatusText(cok),
	})
}
