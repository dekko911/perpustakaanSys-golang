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

	if r.ID != id {
		return nil, fmt.Errorf("role not found")
	}

	return r, nil
}

func (s *Store) GetRoleByName(name string) (*types.Role, error) {
	rows, err := s.db.Query("SELECT * FROM roles WHERE name = ?", name)
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

	if r.ID == "" {
		return nil, fmt.Errorf("role not found")
	}

	return r, nil
}

func (s *Store) CreateRole(r *types.Role) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}

	_, err := s.db.Exec("INSERT INTO roles (id, name) VALUES (?,?)", r.ID, r.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateRole(id string, r *types.Role) error {
	_, err := s.db.Exec("UPDATE roles SET name = ? WHERE id = ?", r.Name, id)
	if err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf("role not found")
	}

	return nil
}
