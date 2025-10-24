package helper

import (
	"database/sql"
	"fmt"
	"perpus_backend/types"
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
	var roleID sql.NullString
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
		&r.Name,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	if roleID.Valid {
		r.ID = roleID.String
		return u, r, nil
	}

	return u, nil, nil
}

func GetUserByID(id string, db *sql.DB) (*types.User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
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

func GetUserByEmail(email string, db *sql.DB) (*types.User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
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
