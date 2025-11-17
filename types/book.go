package types

import (
	"time"
)

type Book struct {
	ID        string    `json:"id"`
	IdBuku    string    `json:"id_buku,omitempty"` // slug type, not relation
	JudulBuku string    `json:"judul_buku"`
	CoverBuku string    `json:"cover_buku,omitempty"` // image
	BukuPDF   string    `json:"buku_pdf,omitempty"`   // pdf
	Penulis   string    `json:"penulis,omitempty"`
	Pengarang string    `json:"pengarang,omitempty"`
	Tahun     int       `json:"tahun,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

type BookStore interface {
	GetBooks() ([]*Book, error)
	GetBookByID(id string) (*Book, error)
	GetBookByJudulBuku(judulBuku string) (*Book, error)
	CreateBook(*Book) error
	UpdateBook(id string, b *Book) error
	DeleteBook(id string) error
}

type SetPayloadBook struct {
	JudulBuku string `form:"judul_buku" validate:"required,min=3"`
	Penulis   string `form:"penulis" validate:"required"`
	Pengarang string `form:"pengarang" validate:"required"`
	Tahun     string `form:"tahun" validate:"required,min=2"`
}

type SetPayloadUpdateBook struct {
	JudulBuku string `form:"judul_buku" validate:"omitempty,required,min=3"`
	Penulis   string `form:"penulis" validate:"omitempty,required"`
	Pengarang string `form:"pengarang" validate:"omitempty,required"`
	Tahun     string `form:"tahun" validate:"omitempty,required,min=2"`
}
