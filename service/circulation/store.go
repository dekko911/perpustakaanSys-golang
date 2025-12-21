package circulation

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

func (s *Store) GetCirculations(ctx context.Context) ([]*types.Circulation, error) {
	sortByColumn := "id_skl"
	sortOrder := "DESC"

	if !utils.IsValidSortColumn(sortByColumn) {
		return nil, fmt.Errorf("invalid sort column: %s", sortByColumn)
	}

	if !utils.IsValidSortOrder(sortOrder) {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	query := fmt.Sprintf(`SELECT c.id, c.buku_id, c.id_skl, c.peminjam, c.tanggal_pinjam, c.jatuh_tempo, c.denda, c.created_at, c.updated_at, b.id, b.judul_buku FROM circulations c INNER JOIN books b ON c.buku_id = b.id ORDER BY %s %s`, sortByColumn, sortOrder)

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

	c := make([]*types.Circulation, 0)

	for rows.Next() {
		circulation, book, err := helper.ScanEachRowIntoCirculation(rows)
		if err != nil {
			return nil, err
		}

		circulation.Book = book
		c = append(c, circulation)
	}

	return c, nil
}

func (s *Store) GetCirculationByID(ctx context.Context, id string) (*types.Circulation, error) {
	circKey := utils.Redis2Key("circulation", id)

	res, err := s.rdb.Get(ctx, circKey).Result()
	if err == nil {
		circ := new(types.Circulation)

		if err := sonic.Unmarshal([]byte(res), circ); err == nil {
			return circ, nil
		}

		s.rdb.Del(ctx, circKey)
	} else if err != redis.Nil {
		return nil, err
	}

	query := `SELECT
	c.id,
	c.buku_id,
	c.id_skl,
	c.peminjam,
	c.tanggal_pinjam,
	c.jatuh_tempo,
	c.denda,
	c.created_at,
	c.updated_at,
	b.id,
	b.judul_buku
	FROM circulations c
	INNER JOIN books b ON c.buku_id = b.id
	WHERE c.id = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	c, err := helper.ScanAndRetRowCirculation(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(c); err == nil {
		_ = s.rdb.SetEx(ctx, circKey, data, 5*time.Minute).Err()
	}

	return c, nil
}

func (s *Store) GetCirculationByPeminjam(ctx context.Context, borrowerName string) (*types.Circulation, error) {
	query := `SELECT
	c.id,
	c.buku_id,
	c.id_skl,
	c.peminjam,
	c.tanggal_pinjam,
	c.jatuh_tempo,
	c.denda,
	c.created_at,
	c.updated_at,
	b.id,
	b.judul_buku
	FROM circulations c
	INNER JOIN books b ON c.buku_id = b.id
	WHERE c.peminjam = ?`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	c, err := helper.ScanAndRetRowCirculation(ctx, stmt, borrowerName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Store) CreateCirculation(ctx context.Context, c *types.Circulation) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
	SELECT CAST(SUBSTRING(id_skl, 3) AS UNSIGNED) AS last_num
	FROM circulations
	ORDER BY last_num DESC
	LIMIT 1
	FOR UPDATE
	`

	var lastNum int

	stmtQuery, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmtQuery.Close()

	if err := stmtQuery.QueryRowContext(ctx).Scan(&lastNum); err == sql.ErrNoRows {
		lastNum = 0
	} else if err != nil {
		return err
	}

	IDSKL := new(string)

	if lastNum > 999 {
		*IDSKL = utils.GenerateSpecificID("SKL", lastNum, 4)
	} else {
		*IDSKL = utils.GenerateSpecificID("SKL", lastNum, 3)
	}

	if c.ID == "" {
		c.ID = uuid.NewString()
	}

	if c.IdSKL == "" {
		c.IdSKL = *IDSKL
	}

	stmtInsert, err := tx.Prepare("INSERT INTO circulations (id, buku_id, id_skl, peminjam, tanggal_pinjam, jatuh_tempo, denda) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmtInsert.Close()

	_, err = stmtInsert.ExecContext(ctx, c.ID, c.BukuID, c.IdSKL, c.Peminjam, c.TanggalPinjam, c.JatuhTempo, c.Denda)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateCirculation(ctx context.Context, id string, c *types.Circulation) error {
	circKey := utils.Redis2Key("circulation", id)

	stmt, err := s.db.Prepare("UPDATE circulations SET buku_id = ?, peminjam = ?, tanggal_pinjam = ?, jatuh_tempo = ?, denda = ? WHERE id = ?")
	if err != nil {
		return err
	}

	s.rdb.Del(ctx, circKey)
	_, err = stmt.ExecContext(ctx, c.BukuID, c.Peminjam, c.TanggalPinjam, c.JatuhTempo, c.Denda, id)
	return err
}

func (s *Store) DeleteCirculation(ctx context.Context, id string) error {
	circKey := utils.Redis2Key("circulation", id)

	res, err := s.db.ExecContext(ctx, "DELETE FROM circulations WHERE id = ?", id)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return fmt.Errorf("circulation not found")
	}

	s.rdb.Del(ctx, circKey)
	return nil
}
