package types

import (
	"context"
	"time"
)

type Book struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`

	ID        string `json:"id"`
	IdBuku    string `json:"id_buku,omitempty"` // slug type, not relation
	JudulBuku string `json:"judul_buku"`
	CoverBuku string `json:"cover_buku,omitempty"` // image
	BukuPDF   string `json:"buku_pdf,omitempty"`   // pdf
	Penulis   string `json:"penulis,omitempty"`
	Pengarang string `json:"pengarang,omitempty"`

	Tahun int `json:"tahun,omitempty"`
}

type BookStore interface {
	GetBooksWithPagination(ctx context.Context, page int) ([]*Book, int64, error)
	GetBooksForSearch(ctx context.Context) []*Book

	GetBookByID(ctx context.Context, id string) (*Book, error)
	GetBookByJudulBuku(ctx context.Context, judulBuku string) (*Book, error)

	CreateBook(ctx context.Context, b *Book) error
	UpdateBook(ctx context.Context, id string, b *Book) error
	DeleteBook(ctx context.Context, id string) error
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
