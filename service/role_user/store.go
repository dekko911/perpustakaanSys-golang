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
	query := `SELECT
	u.id AS user_id, 
	u.name AS user_name, 
	u.email AS user_email, 
	u.password AS user_password,
	u.avatar AS user_avatar,
	u.token_version AS user_token_version,
	u.created_at,
	u.updated_at,
	GROUP_CONCAT(r.id SEPARATOR ', ') AS role_id,
	GROUP_CONCAT(r.name SEPARATOR ', ') AS role_name
	FROM users u
	LEFT JOIN role_user ru ON u.id = ru.user_id 
	LEFT JOIN roles r ON ru.role_id = r.id
	WHERE ru.user_id = ?
	GROUP BY ru.user_id`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	u, err := helper.ScanAndRetRowUserAndRole(stmt, userID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) AssignRoleIntoUser(userID, roleID string) error {
	stmt, err := s.db.Prepare("INSERT INTO role_user (user_id, role_id) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(userID, roleID)
	return err
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
		return fmt.Errorf("user or role not found")
	}

	return nil
}
