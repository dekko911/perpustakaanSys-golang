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
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()
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
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(id)
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
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.name = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(name)
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

	stmt, err := s.db.Prepare("INSERT INTO roles (id, name) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(r.ID, r.Name)
	return err
}

func (s *Store) UpdateRole(id string, r types.Role) error {
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
		return fmt.Errorf("role not found")
	}

	return nil
}
