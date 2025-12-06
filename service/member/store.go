package member

import (
	"database/sql"
	"fmt"
	"perpus_backend/helper"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetMembers() ([]*types.Member, error) {
	sortByColumn := "id_anggota"
	sortOrder := "DESC"

	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	query := fmt.Sprintf("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m ORDER BY %s %s", sortByColumn, sortOrder)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]*types.Member, 0)

	for rows.Next() {
		m, err := helper.ScanEachRowIntoMember(rows)
		if err != nil {
			return nil, err
		}

		members = append(members, m)
	}

	return members, nil
}

func (s *Store) GetMemberByID(id string) (*types.Member, error) {
	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(stmt, id)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Store) GetMemberByNama(nama string) (*types.Member, error) {
	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.nama = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(stmt, nama)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Store) GetMemberByNoTelepon(no_phone string) (*types.Member, error) {
	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.no_telepon = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(stmt, no_phone)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Store) CreateMember(m *types.Member) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
	SELECT CAST(SUBSTRING(id_anggota, 3) AS UNSIGNED) AS last_num
	FROM members
	ORDER BY last_num DESC
	LIMIT 1
	FOR UPDATE`

	var lastNum int // initial first the last number in query row members

	if err := tx.QueryRow(query).Scan(&lastNum); err == sql.ErrNoRows {
		lastNum = 0
	} else if err != nil {
		return err
	}

	IDMember := new(string)

	if lastNum > 999 {
		*IDMember = utils.GenerateSpecificID("ID", lastNum, 4)
	} else {
		*IDMember = utils.GenerateSpecificID("ID", lastNum, 3)
	}

	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	if m.IdAnggota == "" {
		m.IdAnggota = *IDMember
	}

	_, err = tx.Exec("INSERT INTO members (id, id_anggota, nama, jenis_kelamin, kelas, no_telepon, profil_anggota) VALUES (?,?,?,?,?,?,?)", m.ID, m.IdAnggota, m.Nama, m.JenisKelamin, m.Kelas, m.NoTelepon, m.ProfilAnggota)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateMember(id string, m *types.Member) error {
	stmt, err := s.db.Prepare("UPDATE members SET nama = ?, jenis_kelamin = ?, kelas = ?, no_telepon = ?, profil_anggota = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(m.Nama, m.JenisKelamin, m.Kelas, m.NoTelepon, m.ProfilAnggota, id)
	return err
}

func (s *Store) DeleteMember(id string) error {
	res, err := s.db.Exec("DELETE FROM members WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}
