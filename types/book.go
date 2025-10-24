package types

import (
	"time"
)

type Book struct {
	ID        string    `json:"id"`
	IdBuku    string    `json:"id_buku"` // slug type, not relation
	JudulBuku string    `json:"judul_buku"`
	CoverBuku string    `json:"cover_buku"` // image
	Penulis   string    `json:"penulis"`
	Pengarang string    `json:"pengarang"`
	Tahun     time.Time `json:"tahun"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BookStore interface {
	GetBooks() ([]*Book, error)
	GetBookByID(id string) (*Book, error)
	CreateBook(*Book) error
	UpdateBook(id string, b *Book) error
	DeleteBook(id string) error
}

type PayloadBook struct {
	IdBuku    string    `form:"id_buku" validate:"required,min=4"`
	JudulBuku string    `form:"judul_buku" validate:"required,min=3"`
	Penulis   string    `form:"penulis" validate:"required"`
	Pengarang string    `form:"pengarang" validate:"required"`
	Tahun     time.Time `form:"tahun" validate:"required,min=2"`
}
