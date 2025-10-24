package types

import (
	"time"
)

type Member struct {
	ID            string    `json:"id"`
	IdAnggota     string    `json:"id_anggota"` // slug type, not relation
	Nama          string    `json:"nama"`
	JenisKelamin  string    `json:"jenis_kelamin"` // enum type
	Kelas         string    `json:"kelas"`
	NoTelepon     string    `json:"no_telepon"`
	AvatarAnggota string    `json:"avatar_anggota"` // image type
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MemberStore interface {
	GetMembers() ([]*Member, error)
	GetMemberByID(id string) (*Member, error)
	CreateMember(*Member) error
	UpdateMember(*Member) error
	DeleteMember(id string) error
}

type PayloadMember struct {
	IdAnggota    string `form:"id_anggota" validate:"required,min=4"`
	Nama         string `form:"nama" validate:"required"`
	JenisKelamin string `form:"jenis_kelamin" validate:"required"`
	Kelas        string `form:"kelas" validate:"required"`
	NoTelepon    string `form:"no_telepon" validate:"required,min=6"`
}
