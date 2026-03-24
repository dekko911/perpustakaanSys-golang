package book

import (
	"errors"
	"fmt"
	"net/http"

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

	r2CoverBookPath = "books/cover_book"
	r2PDFBookPath   = "books/pdf_book"

	size10MB = 10 << 20
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/books", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetBooks, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/book/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleGetBookByID, "admin", "staff", "user"))).Methods(http.MethodGet)

	r.HandleFunc("/book", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleCreateBook, "admin", "staff"))).Methods(http.MethodPost)

	r.HandleFunc("/book/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleUpdateBook, "admin", "staff"))).Methods(http.MethodPut)

	r.HandleFunc("/book/{bookID}", h.jwt.AuthWithJWTToken(h.jwt.RoleGate(h.handleDeleteBook, "admin", "staff"))).Methods(http.MethodDelete)
}

func (h *Handler) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := utils.ParseStringToInt(r.URL.Query().Get("page"))

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	books, lastPage, total, err := h.bookStore.GetBooksWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonResponse{
		Code:      cok,
		Data:      books,
		Page:      page,
		LastPage:  lastPage,
		TotalData: total,
		Status:    http.StatusText(cok),
	})
}

func (h *Handler) handleGetBookByID(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	ctx := r.Context()

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

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

	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := r.ParseMultipartForm(size10MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	payload := types.SetPayloadBook{
		JudulBuku: r.PostForm.Get("judul_buku"),
		Penulis:   r.PostForm.Get("penulis"),
		Pengarang: r.PostForm.Get("pengarang"),
		Tahun:     r.PostForm.Get("tahun"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

		utils.WriteJSONError(w, http.StatusUnprocessableEntity, vErrors)
		return
	}

	if _, err := h.bookStore.GetBookByJudulBuku(ctx, payload.JudulBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("judul_buku: %s is already exists", payload.JudulBuku))
		return
	}

	_, headerImg, errImg := r.FormFile("cover_buku") // get input form file name is "cover_buku"

	_, headerPDF, errPDF := r.FormFile("buku_pdf") // get input form file name is "buku_pdf"

	// fill the cover book, if input form file of cover book is empty
	if errImg == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors.New("required file cover book"))
		filename = "-"
		return
	}

	// fill the pdf file, if input form file of PDF is empty
	if errPDF == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors.New("required file pdf"))
		filePDF = "-"
		return
	}

	if errImg == nil {
		newFilename, err := utils.SetNewFilenameImg(ctx, "original", headerImg, r2CoverBookPath)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filename = newFilename // <- set the filename for cover book image
	}

	if errPDF == nil {
		newFilePDF, err := utils.SetOriginalFilenamePDF(ctx, headerPDF, r2PDFBookPath)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filePDF = newFilePDF // <- set the filename PDF for PDF book
	}

	// TODO: BAGAIMANA AGAR REQUEST TIME UPLOAD 2 FILE TIDAK LEBIH DARI 3 DETIK
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
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	// TODO: buat validasi file dan buatkan function logic untuk menyimpan file ke local storage
	if err := r.ParseMultipartForm(size10MB); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	payload := types.SetPayloadUpdateBook{
		JudulBuku: r.PostForm.Get("judul_buku"),
		Penulis:   r.PostForm.Get("penulis"),
		Pengarang: r.PostForm.Get("pengarang"),
		Tahun:     r.PostForm.Get("tahun"),
	}

	if err := utils.NewValidate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		vErrors := utils.TransformValidationErrorsWithLangIndonesian(errors)

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

	_, headerImg, errImg := r.FormFile("cover_buku")

	_, headerPDF, errPDF := r.FormFile("buku_pdf")

	if errImg == http.ErrMissingFile {
		filename = b.CoverBuku
	}

	if errPDF == http.ErrMissingFile {
		filePDF = b.BukuPDF
	}

	if errImg == nil {
		newFilename, err := utils.UpdateTheFilenameImg(ctx, "random", headerImg, r2CoverBookPath, b.CoverBuku)

		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, err)
			return
		}

		filename = newFilename
	}

	if errPDF == nil {
		newFilePDF, err := utils.UpdateTheOriginalFilenamePDF(ctx, headerPDF, r2PDFBookPath, b.BukuPDF)

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

	if r.Method != http.MethodDelete {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, fmt.Errorf("your method is wrong, current method: %v", r.Method))
		return
	}

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	b, err := h.bookStore.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := utils.DeleteFilepathWithFilename(ctx, b.CoverBuku, b.BukuPDF); err != nil {
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
