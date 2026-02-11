package circulation

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

func (s *Store) GetCirculationsWithPagination(ctx context.Context, page int) ([]*types.Circulation, int64, error) {
	if page < 1 {
		page = 1
	}

	sortByColumn := "id_skl"
	sortOrder := "DESC"

	limit := 10 // set the limit perPage

	circulationsKey, err := utils.SetRedisKeyForPagination("circulations", page, limit)
	if err != nil {
		return nil, 0, err
	}

	res, err := s.rdb.Get(ctx, circulationsKey).Bytes()
	if err == nil {
		data := &types.CirculationsCachePage{}

		if err := sonic.Unmarshal(res, data); err != nil {
			return data.Circulations, data.LastPage, nil
		}

		s.rdb.Del(ctx, circulationsKey)
	} else if err != redis.Nil {

		s.rdb.Del(ctx, circulationsKey)
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT c.id, c.buku_id, c.id_skl, c.peminjam, c.tanggal_pinjam, c.jatuh_tempo, c.denda, c.created_at, c.updated_at, b.id, b.judul_buku, COUNT(*) OVER() AS num_rows FROM circulations c INNER JOIN books b ON c.buku_id = b.id GROUP BY c.id, b.id ORDER BY %s %s LIMIT %d OFFSET %d`, sortByColumn, sortOrder, limit, (page-1)*limit)

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

	circulations := make([]*types.Circulation, 0)

	var lastPage int64

	for rows.Next() {
		circulation, book, total, err := helper.ScanAndCountRowsCirculation(rows)
		if err != nil {
			return nil, 0, err
		}

		lastPage = int64(math.Ceil(float64(total) / float64(limit)))

		circulation.Book = book

		circulations = append(circulations, circulation)
	}

	payloadCirculations := types.CirculationsCachePage{
		Circulations: circulations,
		LastPage:     lastPage,
	}

	if data, err := sonic.Marshal(payloadCirculations); err == nil {
		s.rdb.SetEx(ctx, circulationsKey, data, time.Duration(2)*time.Minute)
	}

	return circulations, lastPage, nil
}

func (s *Store) GetCirculationsForSearch(ctx context.Context) []*types.Circulation {
	query := "SELECT c.id, c.buku_id, c.id_skl, c.peminjam, c.tanggal_pinjam, c.jatuh_tempo, c.denda, c.created_at, c.updated_at, b.id, b.judul_buku FROM circulations c INNER JOIN books b ON c.buku_id = b.id"

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

	circulations := make([]*types.Circulation, 0)

	for rows.Next() {
		circulation, book, err := helper.ScanRowsCirculation(rows)
		if err != nil {
			return nil
		}

		circulation.Book = book

		circulations = append(circulations, circulation)
	}

	return circulations
}

func (s *Store) GetCirculationByID(ctx context.Context, id string) (*types.Circulation, error) {
	circKey, err := utils.SetRedisKey("circulation", id)
	if err != nil {
		return nil, err
	}

	res, err := s.rdb.Get(ctx, circKey).Bytes()
	if err == nil {
		circ := new(types.Circulation)

		if err := sonic.Unmarshal(res, circ); err == nil {
			return circ, nil
		}

		s.rdb.Del(ctx, circKey)
	} else if err != redis.Nil {

		s.rdb.Del(ctx, circKey)
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

	circulation, err := helper.ScanAndRetRowCirculation(ctx, stmt, id)
	if err != nil {
		return nil, err
	}

	if data, err := sonic.Marshal(circulation); err == nil {
		s.rdb.SetEx(ctx, circKey, data, time.Duration(5)*time.Minute)
	}

	return circulation, nil
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

	circulation, err := helper.ScanAndRetRowCirculation(ctx, stmt, borrowerName)
	if err != nil {
		return nil, err
	}

	return circulation, nil
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

	// init prefix SKL001 circulation
	var IDSKL string

	if lastNum > 999 {
		IDSKL, err = utils.GenerateSpecificID("SKL", lastNum, 4)
		if err != nil {
			return err
		}

	} else {
		IDSKL, err = utils.GenerateSpecificID("SKL", lastNum, 3)
		if err != nil {
			return err
		}

	}

	if c.ID == "" {
		c.ID = uuid.NewString()
	}

	if c.IdSKL == "" {
		c.IdSKL = IDSKL
	}

	stmtInsert, err := tx.Prepare("INSERT INTO circulations (id, buku_id, id_skl, peminjam, tanggal_pinjam, jatuh_tempo, denda) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmtInsert.Close()

	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

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
	stmt, err := s.db.Prepare("UPDATE circulations SET buku_id = ?, peminjam = ?, tanggal_pinjam = ?, jatuh_tempo = ?, denda = ? WHERE id = ?")
	if err != nil {
		return err
	}

	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, c.BukuID, c.Peminjam, c.TanggalPinjam, c.JatuhTempo, c.Denda, id)
	return err
}

func (s *Store) DeleteCirculation(ctx context.Context, id string) error {
	if err := utils.InvalidateAllKeysInCache(s.rdb, ctx); err != nil {
		return err
	}

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

	return nil
}
