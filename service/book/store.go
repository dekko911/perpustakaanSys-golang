package book

import (
	"database/sql"
	"fmt"
	"perpus_backend/helper"
	"perpus_backend/types"

	"github.com/google/uuid"
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
	stmt, err := s.db.Prepare("SELECT * FROM books")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
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
	stmt, err := s.db.Prepare("SELECT * FROM books WHERE id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(id)
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

	if b.ID != id {
		return nil, fmt.Errorf("book not found")
	}

	return b, nil
}

func (s *Store) GetBookByJudulBuku(judulBuku string) (*types.Book, error) {
	stmt, err := s.db.Prepare("SELECT * FROM books WHERE judul_buku = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(judulBuku)
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

	if b.ID == "" {
		return nil, fmt.Errorf("book not found")
	}

	return b, nil
}

func (s *Store) CreateBook(b *types.Book) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	if b.IdBuku == "" {
		b.IdBuku = helper.GenerateNextIDBuku(s.db)
	}

	stmt, err := s.db.Prepare("INSERT INTO books (id, id_buku, judul_buku, cover_buku, penulis, pengarang, tahun) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(b.ID, b.IdBuku, b.JudulBuku, b.CoverBuku, b.Penulis, b.Pengarang, b.Tahun)
	return err
}

func (s *Store) UpdateBook(id string, b *types.Book) error {
	stmt, err := s.db.Prepare("UPDATE books SET judul_buku = ?, cover_buku = ?, penulis = ?, pengarang = ?, tahun = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(b.JudulBuku, b.CoverBuku, b.Penulis, b.Pengarang, b.Tahun, id)
	return err
}

func (s *Store) DeleteBook(id string) error {
	res, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("book with id:%s not found", id)
	}

	return nil
}
