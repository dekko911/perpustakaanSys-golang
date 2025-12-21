package types

import (
	"context"
	"time"
)

type Member struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`

	ID            string `json:"id"`
	IdAnggota     string `json:"id_anggota"` // slug type, not relation
	Nama          string `json:"nama"`
	JenisKelamin  string `json:"jenis_kelamin"` // enum type
	Kelas         string `json:"kelas"`
	NoTelepon     string `json:"no_telepon"`
	ProfilAnggota string `json:"profil_anggota"` // image type
}

type MemberStore interface {
	GetMembers(ctx context.Context) ([]*Member, error)
	GetMemberByID(ctx context.Context, id string) (*Member, error)
	GetMemberByNama(ctx context.Context, nama string) (*Member, error)
	GetMemberByNoTelepon(ctx context.Context, no_phone string) (*Member, error)
	CreateMember(ctx context.Context, m *Member) error
	UpdateMember(ctx context.Context, id string, m *Member) error
	DeleteMember(ctx context.Context, id string) error
}

type SetPayloadMember struct {
	Nama         string `form:"nama" validate:"required"`
	JenisKelamin string `form:"jenis_kelamin" validate:"required"`
	Kelas        string `form:"kelas" validate:"required"`
	NoTelepon    string `form:"no_telepon" validate:"required,min=6"`
}

type SetPayloadUpdateMember struct {
	Nama         string `form:"nama" validate:"omitempty,required"`
	JenisKelamin string `form:"jenis_kelamin" validate:"omitempty,required"`
	Kelas        string `form:"kelas" validate:"omitempty,required"`
	NoTelepon    string `form:"no_telepon" validate:"omitempty,required,min=6"`
}
