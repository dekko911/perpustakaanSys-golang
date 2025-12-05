package role

import (
	"database/sql"
	"fmt"
	"perpus_backend/helper"
	"perpus_backend/types"
	"perpus_backend/utils"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetRoles() ([]*types.Role, error) {
	sortByColumn := "created_at"
	sortOrder := "DESC"

	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	query := fmt.Sprintf("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r ORDER BY %s %s", sortByColumn, sortOrder)

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

	r := make([]*types.Role, 0)

	for rows.Next() {
		role, err := helper.ScanEachRowIntoRole(rows)
		if err != nil {
			return nil, err
		}

		r = append(r, role)
	}

	return r, nil
}

func (s *Store) GetRoleByID(id string) (*types.Role, error) {
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	r, err := helper.ScanAndRetRowRole(stmt, id)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Store) GetRoleByName(name string) (*types.Role, error) {
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.name = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	r, err := helper.ScanAndRetRowRole(stmt, name)
	if err != nil {
		return nil, err
	}

	if r.ID == "" {
		return nil, fmt.Errorf("role not found")
	}

	return r, nil
}

func (s *Store) CreateRole(r types.Role) error {
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
