package db

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
)

func NewMySQLStorage(cfg *mysql.Config) (*sql.DB, error) {
	pool, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// set limit request direct to db.
	pool.SetMaxOpenConns(60)
	pool.SetMaxIdleConns(10)

	return pool, nil
}
