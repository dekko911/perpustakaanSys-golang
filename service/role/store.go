package role

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/perpus_backend/helper"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewStore(db *sql.DB, rdb *redis.Client) *Store {
	return &Store{db: db, rdb: rdb}
}

func (s *Store) GetRoles(ctx context.Context) ([]*types.Role, error) {
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

	rows, err := stmt.QueryContext(ctx)
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

func (s *Store) GetRoleByID(ctx context.Context, id string) (*types.Role, error) {
	roleKey := utils.Redis2Key("role", id)

	res, err := s.rdb.Get(ctx, roleKey).Result()
	if err == nil {
		role := new(types.Role)

		if err := sonic.Unmarshal([]byte(res), role); err == nil {
			return role, nil
		}

		s.rdb.Del(ctx, roleKey)
	} else if err != redis.Nil {
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	r, err := helper.ScanAndRetRowRole(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(r); err == nil {
		_ = s.rdb.SetEx(ctx, roleKey, data, 5*time.Minute).Err()
	}

	return r, nil
}

func (s *Store) GetRoleByName(ctx context.Context, name string) (*types.Role, error) {
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.name = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	r, err := helper.ScanAndRetRowRole(ctx, stmt, name)
	if err != nil {
		return nil, err
	}

	if r.ID == "" {
		return nil, fmt.Errorf("role not found")
	}

	return r, nil
}

func (s *Store) CreateRole(ctx context.Context, r types.Role) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}

	stmt, err := s.db.Prepare("INSERT INTO roles (id, name) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, r.ID, r.Name)
	return err
}

func (s *Store) UpdateRole(ctx context.Context, id string, r types.Role) error {
	roleKey := utils.Redis2Key("role", id)

	stmt, err := s.db.Prepare("UPDATE roles SET name = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	s.rdb.Del(ctx, roleKey)
	_, err = stmt.ExecContext(ctx, r.Name, id)
	return err
}

func (s *Store) DeleteRole(ctx context.Context, id string) error {
	roleKey := utils.Redis2Key("role", id)

	res, err := s.db.ExecContext(ctx, "DELETE FROM roles WHERE id = ?", id)
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

	s.rdb.Del(ctx, roleKey)
	return nil
}
