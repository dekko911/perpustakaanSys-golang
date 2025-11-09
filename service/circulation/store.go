package circulation

import (
	"database/sql"
	"errors"
	"fmt"
	"perpus_backend/helper"
	"perpus_backend/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetCirculations() ([]*types.Circulation, error) {
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
	LEFT JOIN books b ON c.buku_id = b.id`

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

		if book != nil {
			circulation.Book = book
		}

		c = append(c, circulation)
	}

	return c, nil
}

func (s *Store) GetCirculationByID(id string) (*types.Circulation, error) {
	var (
		c types.Circulation
		b types.Book
	)

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

	err = stmt.QueryRow(id).Scan(&c.ID, &c.BukuID, &c.IdSKL, &c.Peminjam, &c.TanggalPinjam, &c.JatuhTempo, &c.Denda, &c.CreatedAt, &c.UpdatedAt, &b.ID, &b.JudulBuku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("circulation not found")
		}

		return nil, err
	}

	c.Book = &b

	return &c, nil
}

func (s *Store) CreateCirculation(*types.Circulation) error {
	return nil
}

func (s *Store) UpdateCirculation(id string, c *types.Circulation) error {
	return nil
}

func (s *Store) DeleteCirculation(id string) error {
	return nil
}
