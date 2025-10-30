package helper

import (
	"database/sql"
	"fmt"
	"perpus_backend/types"
	"strconv"
	"strings"
)

func ScanEachRowIntoUser(rows *sql.Rows) (*types.User, error) {
	u := new(types.User)

	err := rows.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Avatar,
		&u.TokenVersion,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
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

func ScanEachRowUserAndRoleIntoRoleUser(rows *sql.Rows) (*types.User, *types.Role, error) {
	var (
		roleID   sql.NullString
		roleName sql.NullString

		roleCreatedAt sql.NullTime
		roleUpdatedAt sql.NullTime
	)

	u := new(types.User)
	r := new(types.Role)

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
		&roleCreatedAt,
		&roleUpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	if roleID.Valid || roleName.Valid || roleCreatedAt.Valid || roleUpdatedAt.Valid {
		r.ID = roleID.String
		r.Name = roleName.String
		r.CreatedAt = roleCreatedAt.Time
		r.UpdatedAt = roleUpdatedAt.Time
		return u, r, nil
	}

	return u, nil, nil
}

func ScanEachRowIntoBook(rows *sql.Rows) (*types.Book, error) {
	b := new(types.Book)

	err := rows.Scan(
		&b.ID,
		&b.IdBuku,
		&b.JudulBuku,
		&b.CoverBuku,
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

func GetUserByID(id string, db *sql.DB) (*types.User, error) {
	stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	u := new(types.User) // database
	for rows.Next() {
		u, err = ScanEachRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func GetUserByEmail(email string, db *sql.DB) (*types.User, error) {
	stmt, err := db.Prepare("SELECT * FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(email)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	u := new(types.User)
	for rows.Next() {
		u, err = ScanEachRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == "" {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func GenerateNextIDBuku(db *sql.DB) string {
	var lastID string
	stmt, err := db.Prepare("SELECT id_buku FROM books ORDER BY id_buku DESC LIMIT 1")
	if err != nil {
		return err.Error()
	}

	defer stmt.Close()

	if err := stmt.QueryRow().Scan(&lastID); err != nil {
		if err == sql.ErrNoRows {
			return "BK001"
		}

		return err.Error()
	}

	idStr := strings.TrimPrefix(lastID, "BK")
	num, err := strconv.Atoi(idStr)
	if err != nil {
		return err.Error()
	}

	nextID := fmt.Sprintf("BK%03d", num+1)
	return nextID
}
