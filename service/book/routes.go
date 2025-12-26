package book

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/perpus_backend/pkg/jwt"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Handler struct {
	store     types.BookStore
	userStore types.UserStore

	jwt *jwt.AuthJWT
}

func NewHandler(jwt *jwt.AuthJWT, s types.BookStore, us types.UserStore) *Handler {
	return &Handler{
		store:     s,
		userStore: us,
		jwt:       jwt,
	}
}

const (
	cok = http.StatusOK

	dirCoverBookPath = "./assets/public/images/cover/"
	dirPDFBookPath   = "./assets/private/pdf/"

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

	books, lastPage, err := h.store.GetBooksWithPagination(ctx, page)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
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

	book, err := h.store.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:   cok,
		Data:   book,
		Status: http.StatusText(cok),
	})
}

func (h *Handler) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()

		fileName, filePDF  string
		coverPath, pdfPath string
		extCover, extPDF   string
		sizeCover, sizePDF int64
	)

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

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	if _, err := h.store.GetBookByJudulBuku(ctx, payload.JudulBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("judul_buku: %s is already exists", payload.JudulBuku))
		return
	}

	fileCoverBook, headerCB, errCB := r.FormFile("cover_buku") // get input form file name is "cover_buku"

	filePDFbook, headerPDF, errPDF := r.FormFile("buku_pdf") // get input form file name is "buku_pdf"

	// fill the cover book, if input form file of cover book is empty
	if errCB == http.ErrMissingFile {
		fileName = "-"
	}

	// fill the pdf file, if input form file of PDF is empty
	if errPDF == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("required pdf"))
		filePDF = "-"
		return
	}

	extCover = filepath.Ext(headerCB.Filename) // get the extension in input form file cover book
	sizeCover = headerCB.Size                  // get the actual size from input form file "cover_buku"

	extPDF = filepath.Ext(headerPDF.Filename) // get the extension in input form file book pdf
	sizePDF = headerPDF.Size                  // get the actual size from input form file "buku_pdf"

	// doing some validation
	// check if extension is not same with expected want
	if extCover != ".png" && extCover != ".jpg" && extCover != ".jpeg" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only support png, jpg, and jpeg"))
		return
	}

	// check if the size file cover_buku over 1mb
	if sizeCover > size1MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file cover under 1mb"))
		return
	}

	// check if input form file PDF doesn't exist
	if extPDF != ".pdf" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("convert to pdf first"))
		return
	}

	// check if size file pdf over 8mb
	if sizePDF > size8MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file pdf under 8mb"))
		return
	}

	// if cover_buku pass all the validation, then create a file and put in at dir was it set before
	if errCB == nil {
		defer fileCoverBook.Close()

		randomString := xid.New().String()

		fileName = randomString + extCover
		coverPath = dirCoverBookPath + fileName // make it 2 in 1

		dst, _ := os.Create(coverPath)
		defer dst.Close()

		io.Copy(dst, fileCoverBook)
	}

	// if buku_pdf pass all the validation, then create a file and put in at dir was it set before
	if errPDF == nil {
		defer filePDFbook.Close()

		filePDF = headerPDF.Filename
		pdfPath = dirPDFBookPath + filePDF // make it 2 in 1

		dest, _ := os.Create(pdfPath)
		defer dest.Close()

		io.Copy(dest, filePDFbook)
	}

	err := h.store.CreateBook(ctx, &types.Book{
		JudulBuku: payload.JudulBuku,
		CoverBuku: fileName,
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

	utils.WriteJSON(w, http.StatusCreated, utils.JsonData{
		Code:    http.StatusCreated,
		Message: "Book Created!",
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	var (
		ctx = r.Context()

		fileName, filePDF  string
		coverPath, pdfPath string
		extCov, extPdf     string
		sizeCover, sizePDF int64
	)

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, size10MB)

	if err := uuid.Validate(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

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

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, errors)
		return
	}

	b, err := h.store.GetBookByID(ctx, bookID)
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

	// it same goes like the upper, at handleCreateBook()
	fileCoverBook, headerCB, errCB := r.FormFile("cover_buku")

	filePDFBook, headerPDF, errPDF := r.FormFile("buku_pdf")

	if errCB == http.ErrMissingFile {
		fileName = b.CoverBuku
	}

	if errPDF == http.ErrMissingFile {
		filePDF = b.BukuPDF
	}

	extCov = filepath.Ext(headerCB.Filename)
	sizeCover = headerCB.Size

	extPdf = filepath.Ext(headerPDF.Filename)
	sizePDF = headerPDF.Size

	if extCov != ".png" && extCov != ".jpg" && extCov != ".jpeg" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only support png, jpg, and jpeg"))
		return
	}

	if sizeCover > size1MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file cover under 1mb"))
		return
	}

	if extPdf != ".pdf" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("convert to pdf first"))
		return
	}

	if sizePDF > size8MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file pdf under 8mb"))
		return
	}

	if errCB == nil {
		defer fileCoverBook.Close()

		fileImg := dirCoverBookPath + b.CoverBuku
		info, err := os.Stat(fileImg)

		if err == nil {
			if !info.IsDir() {
				os.Remove(fileImg) // remove file old first, for avoid accident stack a goddamn storage memory
			}
		}

		randomString := xid.New().String()

		fileName = randomString + extCov

		coverPath = dirCoverBookPath + fileName

		dst, _ := os.Create(coverPath)
		defer dst.Close()

		io.Copy(dst, fileCoverBook)
	}

	if errPDF == nil {
		defer filePDFBook.Close()

		filePDFOld := dirPDFBookPath + b.BukuPDF
		info, err := os.Stat(filePDFOld)

		if err == nil {
			if !info.IsDir() {
				os.Remove(filePDFOld)
			}
		}

		filePDF = headerPDF.Filename

		pdfPath = dirPDFBookPath + filePDF

		dest, _ := os.Create(pdfPath)
		defer dest.Close()

		io.Copy(dest, fileCoverBook)
	}

	err = h.store.UpdateBook(ctx, bookID, &types.Book{
		JudulBuku: b.JudulBuku,
		CoverBuku: fileName,
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

	utils.WriteJSON(w, cok, utils.JsonData{
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

	b, err := h.store.GetBookByID(ctx, bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// file cover book
	fileImg := dirCoverBookPath + b.CoverBuku
	infoImg, errImg := os.Stat(fileImg)

	if errImg == nil {
		if !infoImg.IsDir() {
			os.Remove(fileImg)
		}
	}

	// file pdf book
	filePDF := dirPDFBookPath + b.BukuPDF
	infoPDF, errPDF := os.Stat(filePDF)

	if errPDF == nil {
		if !infoPDF.IsDir() {
			os.Remove(filePDF)
		}
	}

	if err := h.store.DeleteBook(ctx, bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, cok, utils.JsonData{
		Code:    cok,
		Message: "Book Deleted!",
		Status:  http.StatusText(cok),
	})
}
