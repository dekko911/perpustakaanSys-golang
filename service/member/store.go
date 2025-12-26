package member

import (
	"context"
	"database/sql"
	"fmt"
	"math"
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

func (s *Store) GetMembersWithPagination(ctx context.Context, page int) ([]*types.Member, int64, error) {
	if page < 1 {
		page = 1 // set default page
	}

	sortByColumn := "id_anggota"
	sortOrder := "DESC"

	// tanda ! == data yang false akan menjadi true
	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, 0, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, 0, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	limitPage := 10 // set the limit perPage

	query := fmt.Sprintf("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at, COUNT(*) OVER() AS num_rows FROM members m GROUP BY m.id ORDER BY %s %s LIMIT %d OFFSET %d", sortByColumn, sortOrder, limitPage, (page-1)*limitPage)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, 0, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	members := make([]*types.Member, 0)

	// init lastPage
	var lastPage int64

	for rows.Next() {
		m, count, err := helper.ScanAndCountRowsMember(rows)
		if err != nil {
			return nil, 0, err
		}

		lastPage = int64(math.Ceil(float64(count) / float64(limitPage)))

		members = append(members, m)
	}

	return members, lastPage, nil
}

func (s *Store) GetMembersForSearch(ctx context.Context) []*types.Member {
	query := "SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m"

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil
	}

	defer rows.Close()

	members := make([]*types.Member, 0)

	for rows.Next() {
		m, err := helper.ScanRowsMember(rows)
		if err != nil {
			return nil
		}

		members = append(members, m)
	}

	return members
}

func (s *Store) GetMemberByID(ctx context.Context, id string) (*types.Member, error) {
	memberKey, err := utils.Redis2Key("member", id)
	if err != nil {
		return nil, err
	}

	res, err := s.rdb.Get(ctx, memberKey).Result()
	if err == nil {
		member := new(types.Member)

		if err := sonic.Unmarshal([]byte(res), member); err == nil {
			return member, nil
		}

		s.rdb.Del(ctx, memberKey)
	} else if err != redis.Nil {
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.id = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(m); err == nil {
		_ = s.rdb.SetEx(ctx, memberKey, data, 5*time.Minute)
	}

	return m, nil
}

func (s *Store) GetMemberByNama(ctx context.Context, nama string) (*types.Member, error) {
	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.nama = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(ctx, stmt, nama)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Store) GetMemberByNoTelepon(ctx context.Context, no_phone string) (*types.Member, error) {
	stmt, err := s.db.Prepare("SELECT m.id, m.id_anggota, m.nama, m.jenis_kelamin, m.kelas, m.no_telepon, m.profil_anggota, m.created_at, m.updated_at FROM members m WHERE m.no_telepon = ?")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	m, err := helper.ScanAndRetRowMember(ctx, stmt, no_phone)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Store) CreateMember(ctx context.Context, m *types.Member) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
	SELECT CAST(SUBSTRING(id_anggota, 3) AS UNSIGNED) AS last_num
	FROM members
	ORDER BY last_num DESC
	LIMIT 1
	FOR UPDATE`

	stmtSelect, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmtSelect.Close()

	var lastNum int // initial first the last number in query row members

	if err := stmtSelect.QueryRowContext(ctx).Scan(&lastNum); err == sql.ErrNoRows {
		lastNum = 0
	} else if err != nil {
		return err
	}

	// init prefix ID001 member
	var IDMember string

	if lastNum > 999 {
		IDMember, err = utils.GenerateSpecificID("ID", lastNum, 4)
		if err != nil {
			return err
		}

	} else {
		IDMember, err = utils.GenerateSpecificID("ID", lastNum, 3)
		if err != nil {
			return err
		}

	}

	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	if m.IdAnggota == "" {
		m.IdAnggota = IDMember
	}

	stmtInsert, err := tx.Prepare("INSERT INTO members (id, id_anggota, nama, jenis_kelamin, kelas, no_telepon, profil_anggota) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmtInsert.Close()

	_, err = stmtInsert.ExecContext(ctx, m.ID, m.IdAnggota, m.Nama, m.JenisKelamin, m.Kelas, m.NoTelepon, m.ProfilAnggota)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateMember(ctx context.Context, id string, m *types.Member) error {
	memberKey, err := utils.Redis2Key("member", id)
	if err != nil {
		return err
	}

	stmt, err := s.db.Prepare("UPDATE members SET nama = ?, jenis_kelamin = ?, kelas = ?, no_telepon = ?, profil_anggota = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	s.rdb.Del(ctx, memberKey)
	_, err = stmt.ExecContext(ctx, m.Nama, m.JenisKelamin, m.Kelas, m.NoTelepon, m.ProfilAnggota, id)
	return err
}

func (s *Store) DeleteMember(ctx context.Context, id string) error {
	memberKey, err := utils.Redis2Key("member", id)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM members WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	s.rdb.Del(ctx, memberKey)
	return nil
}
