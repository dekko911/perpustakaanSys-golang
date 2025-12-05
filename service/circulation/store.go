package circulation

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

func (s *Store) GetCirculations() ([]*types.Circulation, error) {
	sortByColumn := "id_skl"
	sortOrder := "DESC"

	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	query := fmt.Sprintf(`SELECT
	c.id,
	c.buku_id,
	c.id_skl,
	c.peminjam,
	c.tanggal_pinjam,
	c.jatuh_tempo,
	c.denda,
	c.created_at,
	c.updated_at,
	b.id,
	b.judul_buku
	FROM circulations c
	INNER JOIN books b ON c.buku_id = b.id
	ORDER BY %s %s`, sortByColumn, sortOrder)

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

	c := make([]*types.Circulation, 0)

	for rows.Next() {
		circulation, book, err := helper.ScanEachRowIntoCirculation(rows)
		if err != nil {
			return nil, err
		}

		circulation.Book = book
		c = append(c, circulation)
	}

	return c, nil
}

func (s *Store) GetCirculationByID(id string) (*types.Circulation, error) {
	query := `SELECT
	c.id,
	c.buku_id,
	c.id_skl,
	c.peminjam,
	c.tanggal_pinjam,
	c.jatuh_tempo,
	c.denda,
	c.created_at,
	c.updated_at,
	b.id,
	b.judul_buku
	FROM circulations c
	INNER JOIN books b ON c.buku_id = b.id
	WHERE c.id = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	c, err := helper.ScanAndRetRowCirculation(stmt, id)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Store) GetCirculationByPeminjam(borrowerName string) (*types.Circulation, error) {
	query := `SELECT
	c.id,
	c.buku_id,
	c.id_skl,
	c.peminjam,
	c.tanggal_pinjam,
	c.jatuh_tempo,
	c.denda,
	c.created_at,
	c.updated_at,
	b.id,
	b.judul_buku
	FROM circulations c
	INNER JOIN books b ON c.buku_id = b.id
	WHERE c.peminjam = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	c, err := helper.ScanAndRetRowCirculation(stmt, borrowerName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Store) CreateCirculation(c *types.Circulation) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}

	if c.IdSKL == "" {
		c.IdSKL = helper.GenerateNextIDSKL(s.db)
	}

	stmt, err := s.db.Prepare("INSERT INTO circulations (id, buku_id, id_skl, peminjam, tanggal_pinjam, jatuh_tempo, denda) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(c.ID, c.BukuID, c.IdSKL, c.Peminjam, c.TanggalPinjam, c.JatuhTempo, c.Denda)
	return err
}

func (s *Store) UpdateCirculation(id string, c *types.Circulation) error {
	stmt, err := s.db.Prepare("UPDATE circulations SET buku_id = ?, peminjam = ?, tanggal_pinjam = ?, jatuh_tempo = ?, denda = ? WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(c.BukuID, c.Peminjam, c.TanggalPinjam, c.JatuhTempo, c.Denda, id)
	return err
}

func (s *Store) DeleteCirculation(id string) error {
	res, err := s.db.Exec("DELETE FROM circulations WHERE id = ?", id)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return fmt.Errorf("circulation not found")
	}

	return nil
}
