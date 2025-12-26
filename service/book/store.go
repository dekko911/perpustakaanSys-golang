package book

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/perpus_backend/helper"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewStore(db *sql.DB, rdb *redis.Client) *Store {
	return &Store{db: db, rdb: rdb}
}

func (s *Store) GetBooksWithPagination(ctx context.Context, page int) ([]*types.Book, int64, error) {
	if page < 1 {
		page = 1
	}

	sortByColumn := "id_buku"
	sortOrder := "DESC"

	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, 0, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, 0, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	limit := 10

	query := fmt.Sprintf("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at, COUNT(*) OVER() AS num_rows FROM books b GROUP BY b.id ORDER BY %s %s LIMIT %d OFFSET %d", sortByColumn, sortOrder, limit, (page-1)*limit)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, 0, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	books := make([]*types.Book, 0)

	var lastPage int64

	for rows.Next() {
		b, total, err := helper.ScanAndCountRowsBook(rows)
		if err != nil {
			return nil, 0, err
		}

		lastPage = int64(math.Ceil(float64(total) / float64(limit)))

		books = append(books, b)
	}

	return books, lastPage, nil
}

func (s *Store) GetBooksForSearch(ctx context.Context) []*types.Book {
	query := "SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b"

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil
	}

	defer rows.Close()

	books := make([]*types.Book, 0)

	for rows.Next() {
		b, err := helper.ScanRowsBook(rows)
		if err != nil {
			return nil
		}

		books = append(books, b)
	}

	return books
}

func (s *Store) GetBookByID(ctx context.Context, id string) (*types.Book, error) {
	bookKey, err := utils.Redis2Key("book", id)
	if err != nil {
		return nil, err
	}

	res, err := s.rdb.Get(ctx, bookKey).Result()
	if err == nil {
		book := new(types.Book)

		if err := sonic.Unmarshal([]byte(res), book); err == nil {
			return book, nil
		}

		s.rdb.Del(ctx, bookKey)
	} else if err != redis.Nil {
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b WHERE b.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	b, err := helper.ScanAndRetRowBook(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(b); err == nil {
		_ = s.rdb.SetEx(ctx, bookKey, data, 5*time.Minute)
	}

	return b, nil
}

func (s *Store) GetBookByJudulBuku(ctx context.Context, judulBuku string) (*types.Book, error) {
	stmt, err := s.db.Prepare("SELECT b.id, b.id_buku, b.judul_buku, b.cover_buku, b.buku_pdf, b.penulis, b.pengarang, b.tahun, b.created_at, b.updated_at FROM books b WHERE b.judul_buku = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	b, err := helper.ScanAndRetRowBook(ctx, stmt, judulBuku)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *Store) CreateBook(ctx context.Context, b *types.Book) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
	SELECT CAST(SUBSTRING(id_buku, 3) AS UNSIGNED) as last_num
	FROM books
	ORDER BY last_num DESC
	LIMIT 1
	FOR UPDATE
	`

	var lastNum int

	stmtQuery, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmtQuery.Close()

	if err := stmtQuery.QueryRowContext(ctx).Scan(&lastNum); err == sql.ErrNoRows {
		lastNum = 0
	} else if err != nil {
		return err
	}

	var IDBook string

	if lastNum > 999 {
		IDBook, err = utils.GenerateSpecificID("BK", lastNum, 4)
		if err != nil {
			return err
		}

	} else {
		IDBook, err = utils.GenerateSpecificID("BK", lastNum, 3)
		if err != nil {
			return err
		}

	}

	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	if b.IdBuku == "" {
		b.IdBuku = IDBook
	}

	stmtInsert, err := tx.Prepare("INSERT INTO books (id, id_buku, judul_buku, cover_buku, buku_pdf, penulis, pengarang, tahun) VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmtInsert.Close()

	_, err = stmtInsert.ExecContext(ctx, b.ID, b.IdBuku, b.JudulBuku, b.CoverBuku, b.BukuPDF, b.Penulis, b.Pengarang, b.Tahun)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateBook(ctx context.Context, id string, b *types.Book) error {
	bookKey, err := utils.Redis2Key("book", id)
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare("UPDATE books SET judul_buku = ?, cover_buku = ?, buku_pdf = ?, penulis = ?, pengarang = ?, tahun = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	s.rdb.Del(ctx, bookKey)
	_, err = stmt.ExecContext(ctx, b.JudulBuku, b.CoverBuku, b.BukuPDF, b.Penulis, b.Pengarang, b.Tahun, id)
	return err
}

func (s *Store) DeleteBook(ctx context.Context, id string) error {
	bookKey, err := utils.Redis2Key("book", id)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM books WHERE id = ?", id)
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

	s.rdb.Del(ctx, bookKey)
	return nil
}
