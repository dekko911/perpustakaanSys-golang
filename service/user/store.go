package user

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

func (s *Store) GetUsers() ([]*types.User, error) {
	query := `SELECT 
	u.id AS user_id, 
	u.name AS user_name, 
	u.email AS user_email, 
	u.password AS user_password,
	u.avatar AS user_avatar,
	u.token_version AS user_token_version,
	u.created_at,
	u.updated_at,
	r.id AS role_id,
	r.name AS role_name
	FROM users u
	LEFT JOIN role_user ru ON u.id = ru.user_id
	LEFT JOIN roles r ON ru.role_id = r.id
	ORDER BY u.created_at DESC`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	usersMap := make(map[string]*types.User)
	for rows.Next() {
		user, role, err := helper.ScanEachRowUserAndRoleIntoRoleUser(rows)
		if err != nil {
			return nil, err
		}

		u, exists := usersMap[user.ID]
		if !exists {
			u = user
			usersMap[user.ID] = u
		}

		if role != nil {
			u.Roles = append(u.Roles, *role)
		}
	}

	users := make([]*types.User, 0, len(usersMap))
	for _, u := range usersMap {
		users = append(users, u)
	}

	return users, nil
}

func (s *Store) GetUserWithRolesByID(id string) (*types.User, error) {
	query := `SELECT
	u.id AS user_id, 
	u.name AS user_name, 
	u.email AS user_email, 
	u.password AS user_password,
	u.avatar AS user_avatar,
	u.token_version AS user_token_version,
	u.created_at,
	u.updated_at,
	r.id AS role_id,
	r.name AS role_name
	FROM users u
	LEFT JOIN role_user ru ON u.id = ru.user_id 
	LEFT JOIN roles r ON ru.role_id = r.id
	WHERE u.id = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var (
		u types.User

		roleID, roleName sql.NullString
	)

	r := new(types.Role)

	err = stmt.QueryRow(id).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Avatar, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt, &roleID, &roleName)
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

func (s *Store) GetUserWithRolesByEmail(email string) (*types.User, error) {
	query := `SELECT
	u.id AS user_id, 
	u.name AS user_name, 
	u.email AS user_email, 
	u.password AS user_password,
	u.avatar AS user_avatar,
	u.token_version AS user_token_version,
	u.created_at,
	u.updated_at,
	r.id AS role_id,
	r.name AS role_name
	FROM users u
	LEFT JOIN role_user ru ON u.id = ru.user_id 
	LEFT JOIN roles r ON ru.role_id = r.id
	WHERE u.email = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var (
		u types.User

		roleID, roleName sql.NullString
	)

	r := new(types.Role)

	err = stmt.QueryRow(email).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Avatar, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt, &roleID, &roleName)
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

	if u.ID == "" {
		return nil, fmt.Errorf("user not found")
	}

	return &u, nil
}

func (s *Store) CreateUser(u *types.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}

	stmt, err := s.db.Prepare("INSERT INTO users (id, name, email, password, avatar) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(u.ID, u.Name, u.Email, u.Password, u.Avatar)
	return err
}

func (s *Store) UpdateUser(id string, u *types.User) error {
	stmt, err := s.db.Prepare("UPDATE users SET name = ?, email = ?, password = ?, avatar = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(u.Name, u.Email, u.Password, u.Avatar, id)
	return err
}

func (s *Store) DeleteUser(id string) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (s *Store) IncrementTokenVersion(id string) error {
	// use s.db.Prepare(query) and stmt(variable).Exec(...args) <- when used at SPAM MOTHERFUCKER
	// stmt, err := s.db.Prepare("UPDATE users SET token_version = token_version + 1 WHERE id = ?")
	// if err != nil {
	// 	return err
	// }

	// defer stmt.Close()

	// _, err = stmt.Exec(id)

	// use s.db.Exec(query, ...args) <- when used it once go execution.
	_, err := s.db.Exec("UPDATE users SET token_version = token_version + 1 WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
