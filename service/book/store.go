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

func (s *Store) GetBookByIDBuku(idBuku string) (*types.Book, error) {
	rows, err := s.db.Query("SELECT * FROM book WHERE id_buku = ?", idBuku)
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

func (s *Store) GetBookByJudulBuku(judulBuku string) (*types.Book, error) {
	rows, err := s.db.Query("SELECT * FROM book WHERE judul_buku = ?", judulBuku)
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

	_, err := s.db.Exec("INSERT INTO books (id, id_buku, judul_buku, cover_buku, penulis, pengarang, tahun) VALUES (?,?,?,?,?,?,?)", b.ID, b.IdBuku, b.JudulBuku, b.CoverBuku, b.Penulis, b.Pengarang, b.Tahun)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateBook(id string, b *types.Book) error {
	_, err := s.db.Exec("UPDATE books SET id_buku = ?, judul_buku = ?, cover_buku = ?, penulis = ?, pengarang = ?, tahun = ? WHERE id = ?", b.IdBuku, b.JudulBuku, b.CoverBuku, b.Penulis, b.Pengarang, b.Tahun, id)
	if err != nil {
		return err
	}

	return nil
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
