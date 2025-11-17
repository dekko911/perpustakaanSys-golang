package types

import (
	"time"
)

type Circulation struct {
	ID            string    `json:"id"`
	BukuID        string    `json:"buku_id"` // relation
	Book          *Book     `json:"book"`
	IdSKL         string    `json:"id_skl"` // slug type
	Peminjam      string    `json:"peminjam"`
	TanggalPinjam time.Time `json:"tanggal_pinjam"` // date type, not datetime
	JatuhTempo    time.Time `json:"jatuh_tempo"`
	Denda         float64   `json:"denda"`
	CreatedAt     time.Time `json:"created_at,omitzero"`
	UpdatedAt     time.Time `json:"updated_at,omitzero"`
}

type CirculationStore interface {
	GetCirculations() ([]*Circulation, error)
	GetCirculationByID(id string) (*Circulation, error)
	GetCirculationByPeminjam(borrowerName string) (*Circulation, error)
	CreateCirculation(*Circulation) error
	UpdateCirculation(id string, c *Circulation) error
	DeleteCirculation(id string) error
}

type SetPayloadCirculation struct {
	BukuID        string `form:"book_id" validate:"required"`
	Peminjam      string `form:"peminjam" validate:"required"`
	TanggalPinjam string `form:"tanggal_pinjam" validate:"required"`
	JatuhTempo    string `form:"jatuh_tempo" validate:"required"`
	Denda         string `form:"denda" validate:"required"`
}

type SetPayloadUpdateCirculation struct {
	BukuID        string `form:"book_id" validate:"omitempty,required"`
	Peminjam      string `form:"peminjam" validate:"omitempty,required"`
	TanggalPinjam string `form:"tanggal_pinjam" validate:"omitempty,required"`
	JatuhTempo    string `form:"jatuh_tempo" validate:"omitempty,required"`
	Denda         string `form:"denda" validate:"omitempty,required"`
}
