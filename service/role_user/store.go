package roleuser

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

		_ = s.rdb.Del(ctx, userKey).Err()
	} else if err != redis.Nil {

		_ = s.rdb.Del(ctx, userKey).Err()
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

func (s *Store) GetUserAndRoleNames(ctx context.Context, userID string) (*types.User, map[string][]string, error) {
	userKey, err := utils.Redis2Key("user", userID)
	if err != nil {
		return nil, nil, err
	}

	// initial empty roles
	roles := make(map[string][]string)

	res, err := s.rdb.Get(ctx, userKey).Result()
	if err == nil {
		user := new(types.User)

		var (
			redisRoleIDs   []string
			redisRoleNames []string
		)

		if err := sonic.Unmarshal([]byte(res), user); err == nil {
			for _, r := range user.Roles {
				redisRoleIDs = strings.Split(r.ID, ", ")
				redisRoleNames = strings.Split(r.Name, ", ")
			}

			for _, id := range redisRoleIDs {
				roles["id"] = append(roles["id"], id)
			}

			for _, name := range redisRoleNames {
				roles["name"] = append(roles["name"], name)
			}

			return user, roles, nil
		}

		_ = s.rdb.Del(ctx, userKey).Err()
	} else if err != redis.Nil {

		_ = s.rdb.Del(ctx, userKey).Err()
		return nil, nil, err
	}

	query := "SELECT u.id AS user_id, u.name AS user_name, GROUP_CONCAT(r.id SEPARATOR ', ') AS role_id, GROUP_CONCAT(r.name SEPARATOR ', ') AS role_name FROM users u LEFT JOIN role_user ru ON u.id = ru.user_id LEFT JOIN roles r ON ru.role_id = r.id WHERE ru.user_id = ? GROUP BY ru.user_id"

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, nil, err
	}

	defer stmt.Close()

	u, err := helper.ScanAndRetRowUserAndRoleNames(ctx, stmt, userID)
	if err != nil {
		return nil, nil, err
	}

	var (
		roleIDs   []string
		roleNames []string
	)

	for _, r := range u.Roles {
		roleIDs = strings.Split(r.ID, ", ")     // jika id nya ada lebih, maka pecahkan menjadi subbagian?
		roleNames = strings.Split(r.Name, ", ") // jika nama nya ada lebih, maka pecahkan menjadi subbagian?
	}

	for _, id := range roleIDs {
		roles["id"] = append(roles["id"], id)
	}

	for _, name := range roleNames {
		roles["name"] = append(roles["name"], name)
	}

	if data, err := sonic.Marshal(u); err == nil {
		_ = s.rdb.SetEx(ctx, userKey, data, 5*time.Minute).Err()
	}

	return u, roles, nil
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
	userKey, errUser := utils.Redis2Key("user", userID)
	roleKey, errRole := utils.Redis2Key("role", roleID)

	if errUser != nil {
		_ = s.rdb.Del(ctx, userKey, roleKey).Err()
		return errUser
	}

	if errRole != nil {
		_ = s.rdb.Del(ctx, userKey, roleKey).Err()
		return errRole
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM role_user WHERE user_id = ? AND role_id = ?", userID, roleID)
	if err != nil {
		_ = s.rdb.Del(ctx, userKey, roleKey).Err()
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		_ = s.rdb.Del(ctx, userKey, roleKey).Err()
		return err
	}

	if rows == 0 {
		_ = s.rdb.Del(ctx, userKey, roleKey).Err()
		return fmt.Errorf("user or role not found")
	}

	_ = s.rdb.Del(ctx, userKey, roleKey).Err()
	return nil
}
