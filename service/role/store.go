package role

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

func (s *Store) GetRoles() ([]*types.Role, error) {
	rows, err := s.db.Query("SELECT * FROM roles")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	r := make([]*types.Role, 0)
	for rows.Next() {
		roles, err := helper.ScanEachRowIntoRole(rows)
		if err != nil {
			return nil, err
		}

		r = append(r, roles)
	}

	return r, nil
}

func (s *Store) GetRoleByID(id string) (*types.Role, error) {
	rows, err := s.db.Query("SELECT * FROM roles WHERE id = ?", id)
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
	}

	return r, nil
}

func (s *Store) CreateRole(r *types.Role) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}

	stmt, err := s.db.Prepare("INSERT INTO roles (id, name) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(r.ID, r.Name)
	return err
}

func (s *Store) UpdateRole(id string, r *types.Role) error {
	stmt, err := s.db.Prepare("UPDATE roles SET name = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(r.Name, id)
	return err
}

func (s *Store) DeleteRole(id string) error {
	res, err := s.db.Exec("DELETE FROM roles WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("role with id:%s not found", id)
	}

	return nil
}
