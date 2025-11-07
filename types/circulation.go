package types

import (
	"time"
)

type Circulation struct {
	ID            string    `json:"id"`
	BookId        string    `json:"buku_id"` // relation
	IdSKL         string    `json:"id_skl"`  // slug type
	Peminjam      string    `json:"peminjam"`
	TanggalPinjam time.Time `json:"tanggal_pinjam"`
	JatuhTempo    string    `json:"jatuh_tempo"`
	Denda         string    `json:"denda"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CirculationStore interface {
	GetCirculations() ([]*Circulation, error)
	GetCirculationByID(id string) (*Circulation, error)
	CreateCirculation(*Circulation) error
	UpdateCirculation(id string, c *Circulation) error
	DeleteCirculation(id string) error
}

type PayloadCirculation struct {
	BookId        string    `form:"book_id" validate:"required"`
	IdSKL         string    `form:"id_skl" validate:"required,min=4"`
	Peminjam      string    `form:"peminjam" validate:"required"`
	TanggalPinjam time.Time `form:"tanggal_pinjam" validate:"required"`
	JatuhTempo    string    `form:"jatuh_tempo" validate:"required"`
	Denda         string    `form:"denda" validate:"required"`
}
