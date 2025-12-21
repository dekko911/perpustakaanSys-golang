package user

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

func (s *Store) GetUsers(ctx context.Context) ([]*types.User, error) {
	sortByColumn := "created_at"
	sortOrder := "DESC"

	// prevent SQL INJECTION
	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	// prevent SQL INJECTION
	if !utils.IsValidSortOrder(sortOrder) {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	query := fmt.Sprintf(`SELECT 
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
	ORDER BY %s %s`, sortByColumn, sortOrder)

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

	usersMap := make(map[string]*types.User)

	for rows.Next() { // <- like while
		user, role, err := helper.ScanEachRowUserAndRoleIntoUser(rows)
		if err != nil {
			return nil, err
		}

		u, exists := usersMap[user.ID]
		if !exists {
			u = user              // assign user scan to usersMap
			usersMap[user.ID] = u // and last, assign final user scan to usersMap
		}

		if role != nil {
			u.Roles = append(u.Roles, *role) // add roles data to usersMap
		}
	}

	users := make([]*types.User, 0, len(usersMap))

	for _, u := range usersMap {
		users = append(users, u)
	}

	return users, nil
}

func (s *Store) GetUserWithRolesByID(ctx context.Context, id string) (*types.User, error) {
	// init redis db
	userKey := utils.Redis2Key("user", id)

	// get from cache
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
		WHERE u.id = ?
		GROUP BY u.id`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	u, err := helper.ScanAndRetRowUserAndRole(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	// set cache user
	if data, err := sonic.Marshal(u); err == nil {
		_ = s.rdb.SetEx(ctx, userKey, data, 5*time.Minute).Err()
	}

	return u, nil
}

func (s *Store) GetUserWithRolesByEmail(ctx context.Context, email string) (*types.User, error) {
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
	WHERE u.email = ?
	GROUP BY u.id`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	u, err := helper.ScanAndRetRowUserAndRole(ctx, stmt, email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) CreateUser(ctx context.Context, u *types.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}

	stmt, err := s.db.Prepare("INSERT INTO users (id, name, email, password, avatar) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, u.ID, u.Name, u.Email, u.Password, u.Avatar)
	return err
}

func (s *Store) UpdateUser(ctx context.Context, id string, u *types.User) error {
	userKey := utils.Redis2Key("user", id)

	stmt, err := s.db.Prepare("UPDATE users SET name = ?, email = ?, password = ?, avatar = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, u.Name, u.Email, u.Password, u.Avatar, id)

	s.rdb.Del(ctx, userKey)
	return err
}

func (s *Store) DeleteUser(ctx context.Context, id string) error {
	userKey := utils.Redis2Key("user", id)

	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
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

	s.rdb.Del(ctx, userKey)
	return nil
}

func (s *Store) IncrementTokenVersion(ctx context.Context, id, token string) error {
	userKey := utils.Redis2Key("user", id)

	if err := uuid.Validate(id); err != nil {
		return fmt.Errorf("invalid uuid format")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE users SET token_version = token_version + 1 WHERE id = ?", id)
	if err != nil {
		return err
	}

	s.rdb.Del(ctx, userKey, token)
	return nil
}
