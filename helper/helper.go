package helper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/perpus_backend/types"
)

type stringAndNumberOnly interface {
	~string | ~int | ~int64 | ~float64
}

func ScanAndCountRowsUserAndRole(rows *sql.Rows) (*types.User, *types.Role, int64, error) {
	u := new(types.User)
	r := new(types.Role)

	var (
		count int64

		roleID, roleName sql.NullString
	)

	err := rows.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Avatar,
		&u.TokenVersion,
		&u.CreatedAt,
		&u.UpdatedAt,
		&roleID,
		&roleName,
		&count,
	)
	if err != nil {
		return nil, nil, 0, err
	}

	if roleID.Valid && roleName.Valid {
		r.ID = roleID.String
		r.Name = roleName.String
		return u, r, count, nil
	}

	return u, nil, count, nil
}

func ScanRowsUserAndRole(rows *sql.Rows) (*types.User, *types.Role, error) {
	u := new(types.User)
	r := new(types.Role)

	var roleID, roleName sql.NullString

	err := rows.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Avatar,
		&u.TokenVersion,
		&u.CreatedAt,
		&u.UpdatedAt,
		&roleID,
		&roleName,
	)
	if err != nil {
		return nil, nil, err
	}

	if roleID.Valid && roleName.Valid {
		r.ID = roleID.String
		r.Name = roleName.String
		return u, r, nil
	}

	return u, nil, nil
}

func ScanEachRowIntoRole(rows *sql.Rows) (*types.Role, error) {
	r := new(types.Role)

	err := rows.Scan(
		&r.ID,
		&r.Name,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return r, nil
}

func ScanAndCountRowsBook(rows *sql.Rows) (*types.Book, int64, error) {
	b := new(types.Book)

	var count int64

	err := rows.Scan(
		&b.ID,
		&b.IdBuku,
		&b.JudulBuku,
		&b.CoverBuku,
		&b.BukuPDF,
		&b.Penulis,
		&b.Pengarang,
		&b.Tahun,
		&b.CreatedAt,
		&b.UpdatedAt,
		&count,
	)
	if err != nil {
		return nil, 0, err
	}

	return b, count, nil
}

func ScanRowsBook(rows *sql.Rows) (*types.Book, error) {
	b := new(types.Book)

	err := rows.Scan(
		&b.ID,
		&b.IdBuku,
		&b.JudulBuku,
		&b.CoverBuku,
		&b.BukuPDF,
		&b.Penulis,
		&b.Pengarang,
		&b.Tahun,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func ScanAndCountRowsMember(rows *sql.Rows) (*types.Member, int64, error) {
	m := new(types.Member)

	var count int64

	err := rows.Scan(
		&m.ID,
		&m.IdAnggota,
		&m.Nama,
		&m.JenisKelamin,
		&m.Kelas,
		&m.NoTelepon,
		&m.ProfilAnggota,
		&m.CreatedAt,
		&m.UpdatedAt,
		&count,
	)
	if err != nil {
		return nil, 0, err
	}

	return m, count, nil
}

func ScanRowsMember(rows *sql.Rows) (*types.Member, error) {
	m := new(types.Member)

	err := rows.Scan(
		&m.ID,
		&m.IdAnggota,
		&m.Nama,
		&m.JenisKelamin,
		&m.Kelas,
		&m.NoTelepon,
		&m.ProfilAnggota,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func ScanAndCountRowsCirculation(rows *sql.Rows) (*types.Circulation, *types.Book, int64, error) {
	c := new(types.Circulation)
	b := new(types.Book)

	var count int64

	err := rows.Scan(
		&c.ID,
		&c.BukuID,
		&c.IdSKL,
		&c.Peminjam,
		&c.TanggalPinjam,
		&c.JatuhTempo,
		&c.Denda,
		&c.CreatedAt,
		&c.UpdatedAt,
		&b.ID,
		&b.JudulBuku,
		&count,
	)
	if err != nil {
		return nil, nil, 0, err
	}

	return c, b, count, nil
}

func ScanRowsCirculation(rows *sql.Rows) (*types.Circulation, *types.Book, error) {
	c := new(types.Circulation)
	b := new(types.Book)

	err := rows.Scan(
		&c.ID,
		&c.BukuID,
		&c.IdSKL,
		&c.Peminjam,
		&c.TanggalPinjam,
		&c.JatuhTempo,
		&c.Denda,
		&c.CreatedAt,
		&c.UpdatedAt,
		&b.ID,
		&b.JudulBuku,
	)
	if err != nil {
		return nil, nil, err
	}

	return c, b, nil
}

// scan and return user row query has given before.
func ScanAndRetRowUserAndRole[T stringAndNumberOnly](ctx context.Context, stmt *sql.Stmt, param T) (*types.User, error) {
	var u types.User
	r := new(types.Role)

	var roleID, roleName sql.NullString

	err := stmt.QueryRowContext(ctx, param).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Avatar, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt, &roleID, &roleName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}

		return nil, err
	}

	if roleID.Valid && roleName.Valid {
		r.ID = roleID.String
		r.Name = roleName.String

		u.Roles = append(u.Roles, *r)
	}

	return &u, nil
}

// scan and return role row query has given before.
func ScanAndRetRowRole[T stringAndNumberOnly](ctx context.Context, stmt *sql.Stmt, param T) (*types.Role, error) {
	var r types.Role

	err := stmt.QueryRowContext(ctx, param).Scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found")
		}

		return nil, err
	}

	return &r, nil
}

// scan and return member row query has given before.
func ScanAndRetRowMember[T stringAndNumberOnly](ctx context.Context, stmt *sql.Stmt, param T) (*types.Member, error) {
	var m types.Member

	err := stmt.QueryRowContext(ctx, param).Scan(&m.ID, &m.IdAnggota, &m.Nama, &m.JenisKelamin, &m.Kelas, &m.NoTelepon, &m.ProfilAnggota, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("member not found")
		}

		return nil, err
	}

	return &m, nil
}

// scan and return book row query has given before.
func ScanAndRetRowBook[T stringAndNumberOnly](ctx context.Context, stmt *sql.Stmt, param T) (*types.Book, error) {
	var b types.Book

	err := stmt.QueryRowContext(ctx, param).Scan(&b.ID, &b.IdBuku, &b.JudulBuku, &b.CoverBuku, &b.BukuPDF, &b.Penulis, &b.Pengarang, &b.Tahun, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("book not found")
		}

		return nil, err
	}

	return &b, nil
}

// scan and return circulation row query has given before.
func ScanAndRetRowCirculation[T stringAndNumberOnly](ctx context.Context, stmt *sql.Stmt, param T) (*types.Circulation, error) {
	var c types.Circulation
	var b types.Book

	err := stmt.QueryRowContext(ctx, param).Scan(&c.ID, &c.BukuID, &c.IdSKL, &c.Peminjam, &c.TanggalPinjam, &c.JatuhTempo, &c.Denda, &c.CreatedAt, &c.UpdatedAt, &b.ID, &b.JudulBuku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("circulation not found")
		}

		return nil, err
	}

	c.Book = &b

	return &c, nil
}

// method AddDocuments custom meili for all tables.
func AddDocumentsWithWait(client meilisearch.ServiceManager, index string, primaryKey string, dbTable any) error {
	res, err := client.Index(index).AddDocuments(dbTable, &primaryKey)
	if err != nil {
		return err
	}

	task, err := client.WaitForTask(res.TaskUID, 3*time.Second)
	if err != nil {
		return err
	}

	if task.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf("error indexing task: %v", task.Error)
	}

	return nil
}
