package roleuser

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/perpus_backend/helper"
	"github.com/perpus_backend/types"
	"github.com/perpus_backend/utils"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewStore(db *sql.DB, rdb *redis.Client) *Store {
	return &Store{db: db, rdb: rdb}
}

// for relations many to many.
func (s *Store) GetUserWithRoleByUserID(ctx context.Context, userID string) (*types.User, error) {
	userKey, err := utils.Redis2Key("user", userID)
	if err != nil {
		return nil, err
	}

	res, err := s.rdb.Get(ctx, userKey).Result()
	if err == nil {
		user := new(types.User)

		if err := sonic.Unmarshal([]byte(res), user); err == nil {
			return user, nil
		}

		s.rdb.Del(ctx, userKey)
	} else if err != redis.Nil {
		return nil, err
	}

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

	u, err := helper.ScanAndRetRowUserAndRole(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(u); err == nil {
		_ = s.rdb.SetEx(ctx, userKey, data, 5*time.Minute).Err()
	}

	return u, nil
}

func (s *Store) AssignRoleIntoUser(ctx context.Context, userID, roleID string) error {
	stmt, err := s.db.Prepare("INSERT INTO role_user (user_id, role_id) VALUES (?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userID, roleID)
	return err
}

func (s *Store) DeleteRoleFromUser(ctx context.Context, userID, roleID string) error {
	userKey, err := utils.Redis2Key("user", userID)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM role_user WHERE user_id = ? AND role_id = ?", userID, roleID)
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

	s.rdb.Del(ctx, userKey)
	return nil
}
