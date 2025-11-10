package book

import (
	"database/sql"
	"errors"
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
	stmt, err := s.db.Prepare("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b ORDER BY b.id_buku DESC")
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
	stmt, err := s.db.Prepare("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b WHERE b.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var b types.Book

	err = stmt.QueryRow(id).Scan(&b.ID, &b.IdBuku, &b.JudulBuku, &b.CoverBuku, &b.BukuPDF, &b.Penulis, &b.Pengarang, &b.Tahun, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("book not found")
		}

		return nil, err
	}

	return &b, nil
}

func (s *Store) GetBookByJudulBuku(judulBuku string) (*types.Book, error) {
	stmt, err := s.db.Prepare("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b WHERE b.judul_buku = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var b types.Book

	err = stmt.QueryRow(judulBuku).Scan(&b.ID, &b.IdBuku, &b.JudulBuku, &b.CoverBuku, &b.BukuPDF, &b.Penulis, &b.Pengarang, &b.Tahun, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("book not found")
		}

		return nil, err
	}

	return &b, nil
}

func (s *Store) CreateBook(b *types.Book) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	if b.IdBuku == "" {
		b.IdBuku = helper.GenerateNextIDBuku(s.db)
	}

	stmt, err := s.db.Prepare("INSERT INTO books (id, id_buku, judul_buku, cover_buku, buku_pdf, penulis, pengarang, tahun) VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(b.ID, b.IdBuku, b.JudulBuku, b.CoverBuku, b.BukuPDF, b.Penulis, b.Pengarang, b.Tahun)
	return err
}

func (s *Store) UpdateBook(id string, b *types.Book) error {
	stmt, err := s.db.Prepare("UPDATE books SET judul_buku = ?, cover_buku = ?, buku_pdf = ?, penulis = ?, pengarang = ?, tahun = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(b.JudulBuku, b.CoverBuku, b.BukuPDF, b.Penulis, b.Pengarang, b.Tahun, id)
	return err
}

func (s *Store) DeleteBook(id string) error {
	res, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return fmt.Errorf("book not found")
	}

	return nil
}
