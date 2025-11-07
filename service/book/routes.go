package book

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/pkg/jwt"
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

	dirCoverBookPath = "./assets/public/images/cover/"
	dirPDFBookPath   = "./assets/private/pdf/"

	size15MB = 15 << 20
)

func NewHandler(s types.BookStore, us types.UserStore) *Handler {
	return &Handler{
		store:     s,
		userStore: us,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/books", jwt.AuthWithJWTToken(jwt.NeededRole(h.userStore, "admin", "staff", "user")(h.handleGetBooks), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/books/{bookID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.userStore, "admin", "staff", "user")(h.handleGetBookByID), h.userStore)).Methods(http.MethodGet)

	r.HandleFunc("/books", jwt.AuthWithJWTToken(jwt.NeededRole(h.userStore, "admin", "staff")(h.handleCreateBook), h.userStore)).Methods(http.MethodPost)

	r.HandleFunc("/books/{bookID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.userStore, "admin", "staff")(h.handleUpdateBook), h.userStore)).Methods(http.MethodPut)

	r.HandleFunc("/books/{bookID}", jwt.AuthWithJWTToken(jwt.NeededRole(h.userStore, "admin", "staff")(h.handleDeleteBook), h.userStore)).Methods(http.MethodDelete)
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
		Status: http.StatusText(COK),
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
		Status: http.StatusText(COK),
	})
}

func (h *Handler) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var (
		payload types.PayloadBook

		fileName, filePDF  string
		coverPath, pdfPath string
		extCover, extPDF   string
		sizeCover, sizePDF int64
	)

	if err := r.ParseMultipartForm(size15MB); err != nil { // 20 = 2 dikalikan sebanyak 20 kali.
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload = types.PayloadBook{
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

	if _, err := h.store.GetBookByJudulBuku(payload.JudulBuku); err == nil {
		utils.WriteJSONError(w, http.StatusBadRequest, fmt.Errorf("judul_buku: %s is already exists", payload.JudulBuku))
		return
	}

	// cover books
	fileCoverBook, headerCoverBook, errCover := r.FormFile("cover_buku")

	// pdf books
	filePDFbook, headerPDF, errPDF := r.FormFile("buku_pdf")

	// fill the cover book
	if errCover == http.ErrMissingFile {
		fileName = "-"
	}

	// fill the pdf file
	if errPDF == http.ErrMissingFile {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("required pdf"))
		filePDF = "-"
		return
	}

	extCover = filepath.Ext(headerCoverBook.Filename) // get extension in file cover book
	sizeCover = headerCoverBook.Size

	extPDF = filepath.Ext(headerPDF.Filename) // get extension in file book pdf
	sizePDF = headerPDF.Size

	// doing validation
	// check if ext no same like at my below
	if extCover != ".png" && extCover != ".jpg" && extCover != ".jpeg" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only support png, jpg, and jpeg"))
		return
	}

	// check if size file cover over 15mb
	if sizeCover > size15MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file cover under 15mb"))
		return
	}

	// check this if file doesn't pdf ext
	if extPDF != ".pdf" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("convert to pdf first"))
		return
	}

	// check if size file pdf over 15mb
	if sizePDF > size15MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file pdf under 15mb"))
		return
	}

	// if it is clean, do this
	if errCover == nil {
		defer fileCoverBook.Close()

		randomString := xid.New().String()

		fileName = randomString + extCover
		coverPath = dirCoverBookPath + fileName

		dst, _ := os.Create(coverPath)
		defer dst.Close()

		io.Copy(dst, fileCoverBook)
	}

	// if it is clean, do this
	if errPDF == nil {
		defer filePDFbook.Close()

		filePDF = headerPDF.Filename
		pdfPath = dirPDFBookPath + filePDF

		dest, _ := os.Create(pdfPath)
		defer dest.Close()

		io.Copy(dest, filePDFbook)
	}

	tahun, err := strconv.Atoi(payload.Tahun)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.store.CreateBook(&types.Book{
		JudulBuku: payload.JudulBuku,
		CoverBuku: fileName,
		BukuPDF:   filePDF,
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
		Status:  http.StatusText(http.StatusCreated),
	})
}

func (h *Handler) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	var (
		payload types.PayloadUpdateBook

		fileName, filePDF  string
		coverPath, pdfPath string
		extCov, extPdf     string
		sizeCover, sizePDF int64
	)

	if r.Method != http.MethodPut {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, errors.New("method doesn't allowed"))
		return
	}

	if err := r.ParseMultipartForm(15 << 20); err != nil { // 20 = 2 dikalikan sebanyak 20 kali.
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	payload = types.PayloadUpdateBook{
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

	b, err := h.store.GetBookByID(bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	tahun, err := strconv.Atoi(payload.Tahun)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
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
		b.Tahun = tahun
	}

	// for cover books
	fileCoverBook, headerCoverB, errCoverB := r.FormFile("cover_buku")

	// for pdf books
	filePDFBook, headerPDFf, errPDFf := r.FormFile("buku_pdf")

	if errCoverB == http.ErrMissingFile {
		fileName = b.CoverBuku
	}

	if errPDFf == http.ErrMissingFile {
		filePDF = b.BukuPDF
	}

	extCov = filepath.Ext(headerCoverB.Filename)
	sizeCover = headerCoverB.Size

	extPdf = filepath.Ext(headerPDFf.Filename)
	sizePDF = headerPDFf.Size

	if extCov != ".png" && extCov != ".jpg" && extCov != ".jpeg" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only support png, jpg, and jpeg"))
		return
	}

	if sizeCover > size15MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file cover under 15mb"))
		return
	}

	if extPdf != ".pdf" {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("convert to pdf first"))
		return
	}

	if sizePDF > size15MB {
		utils.WriteJSONError(w, http.StatusUnprocessableEntity, fmt.Errorf("only serve file pdf under 15mb"))
		return
	}

	if errCoverB == nil {
		defer fileCoverBook.Close()

		fileImg := dirCoverBookPath + b.CoverBuku
		if err := os.Remove(fileImg); err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, err)
			return
		}

		randomString := xid.New().String()

		fileName = randomString + extCov

		coverPath = dirCoverBookPath + fileName

		dst, _ := os.Create(coverPath)
		defer dst.Close()

		io.Copy(dst, fileCoverBook)
	}

	if errPDFf == nil {
		defer filePDFBook.Close()

		filePDFOld := dirPDFBookPath + b.BukuPDF
		if err := os.Remove(filePDFOld); err != nil {
			utils.WriteJSONError(w, http.StatusNotFound, err)
			return
		}

		filePDF = headerPDFf.Filename

		pdfPath = dirPDFBookPath + filePDF

		dest, _ := os.Create(pdfPath)
		defer dest.Close()

		io.Copy(dest, fileCoverBook)
	}

	if err := h.store.UpdateBook(bookID, &types.Book{
		JudulBuku: b.JudulBuku,
		CoverBuku: fileName,
		BukuPDF:   filePDF,
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
		Status:  http.StatusText(COK),
	})
}

func (h *Handler) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["bookID"]

	b, err := h.store.GetBookByID(bookID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}

	// file cover book
	fileImg := dirCoverBookPath + b.CoverBuku
	infoImg, err := os.Stat(fileImg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			utils.WriteJSONError(w, http.StatusNotFound, fmt.Errorf("img not found"))
			return
		}

		utils.WriteJSONError(w, http.StatusInternalServerError, err)
		return
	}
	if !infoImg.IsDir() {
		os.Remove(fileImg)
	}

	// file pdf book
	filePDF := dirPDFBookPath + b.BukuPDF
	infoPDF, err := os.Stat(filePDF)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			utils.WriteJSONError(w, http.StatusNotFound, fmt.Errorf("pdf not found"))
			return
		}

		utils.WriteJSONError(w, http.StatusInternalServerError, err)
	}
	if !infoPDF.IsDir() {
		os.Remove(filePDF)
	}

	if err := h.store.DeleteBook(bookID); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, COK, utils.JsonData{
		Code:    COK,
		Message: "Book Deleted!",
		Status:  http.StatusText(COK),
	})
}
