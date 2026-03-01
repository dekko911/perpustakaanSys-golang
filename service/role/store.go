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

	roles := make([]*types.Role, 0)

	for rows.Next() {
		role, err := helper.ScanEachRowIntoRole(rows)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (s *Store) GetRoleByID(ctx context.Context, id string) (*types.Role, error) {
	roleKey, err := utils.SetRedisKey("role", id)
	if err != nil {
		return nil, err
	}

	res, err := s.rdb.Get(ctx, roleKey).Result()
	if err == nil {
		role := new(types.Role)

		if err := sonic.Unmarshal([]byte(res), role); err == nil {
			return role, nil
		}

		s.rdb.Del(ctx, roleKey)
	} else if err != redis.Nil {

		s.rdb.Del(ctx, roleKey)
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	role, err := helper.ScanAndRetRowRole(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(role); err == nil {
		s.rdb.SetEx(ctx, roleKey, data, time.Duration(5)*time.Minute)
	}

	return role, nil
}

func (s *Store) GetRoleByName(ctx context.Context, name string) (*types.Role, error) {
	stmt, err := s.db.Prepare("SELECT r.id, r.name, r.created_at, r.updated_at FROM roles r WHERE r.name = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	role, err := helper.ScanAndRetRowRole(ctx, stmt, name)
	if err != nil {
		return nil, err
	}

	if role.ID == "" {
		return nil, fmt.Errorf("role not found")
	}

	return role, nil
}

func (s *Store) CreateRole(ctx context.Context, r *types.Role) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}

	stmt, err := s.db.Prepare("INSERT INTO roles (id, name) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

	if err := utils.InvalidateIndexMeili(ctx, "roles"); err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, r.ID, r.Name)
	return err
}

func (s *Store) UpdateRole(ctx context.Context, id string, r *types.Role) error {
	stmt, err := s.db.Prepare("UPDATE roles SET name = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

	if err := utils.InvalidateIndexMeili(ctx, "roles"); err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, r.Name, id)
	return err
}

func (s *Store) DeleteRole(ctx context.Context, id string) error {
	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

	if err := utils.InvalidateIndexMeili(ctx, "roles"); err != nil {
		return err
	}

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

	return nil
}
