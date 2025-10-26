package book

import (
	"database/sql"
	"perpus_backend/helper"
	"perpus_backend/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}
func (s *Store) GetBooks() ([]*types.Book, error) {
	rows, err := s.db.Query("SELECT * FROM books")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	books := make([]*types.Book, 0)

	for rows.Next() {
		b, err := helper.ScanEachRowIntoBook(rows)
		if err != nil {
			return nil, err
		}

		books = append(books, b)
	}

	return books, nil
}

func (s *Store) GetBookByID(id string) (*types.Book, error) {
	rows, err := s.db.Query("SELECT * FROM book WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	b := new(types.Book)
	for rows.Next() {
		b, err = helper.ScanEachRowIntoBook(rows)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}
