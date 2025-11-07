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
	ProfilAnggota string    `json:"profil_anggota"` // image type
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MemberStore interface {
	GetMembers() ([]*Member, error)
	GetMemberByID(id string) (*Member, error)
	GetMemberByNama(nama string) (*Member, error)
	GetMemberByNoTelepon(no_phone string) (*Member, error)
	CreateMember(*Member) error
	UpdateMember(id string, m *Member) error
	DeleteMember(id string) error
}

type PayloadMember struct {
	Nama         string `form:"nama" validate:"required"`
	JenisKelamin string `form:"jenis_kelamin" validate:"required"`
	Kelas        string `form:"kelas" validate:"required"`
	NoTelepon    string `form:"no_telepon" validate:"required,min=6"`
}

type PayloadUpdateMember struct {
	Nama         string `form:"nama" validate:"omitempty,required"`
	JenisKelamin string `form:"jenis_kelamin" validate:"omitempty,required"`
	Kelas        string `form:"kelas" validate:"omitempty,required"`
	NoTelepon    string `form:"no_telepon" validate:"omitempty,required,min=6"`
}
