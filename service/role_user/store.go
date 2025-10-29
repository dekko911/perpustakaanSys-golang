package roleuser

import (
	"database/sql"
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

// for relations many to many.
func (s *Store) GetUserWithRoleByUserID(userID string) (*types.User, error) {
	u, err := helper.GetUserByID(userID, s.db)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r INNER JOIN role_user ru ON r.id = ru.role_id WHERE ru.user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	r := new(types.Role)
	for rows.Next() {
		r, err = helper.ScanEachRowIntoRole(rows)
		if err != nil {
			return nil, err
		}

		if r != nil {
			u.Roles = append(u.Roles, *r)
		}
	}

	if u.ID == "" {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) AssignRoleIntoUser(userID, roleID string) error {
	_, err := s.db.Exec("INSERT INTO role_user (user_id, role_id) VALUES (?,?)", userID, roleID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteRoleFromUser(userID, roleID string) error {
	res, err := s.db.Exec("DELETE FROM role_user WHERE user_id = ? AND role_id = ?", userID, roleID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user with id:%s and role with id:%s not found", userID, roleID)
	}

	return nil
}
