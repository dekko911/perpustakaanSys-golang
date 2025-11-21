package db

import (
	"database/sql"
	"log"
	"perpus_backend/config"

	"github.com/go-sql-driver/mysql"
)

func NewMySQLStorage(cfg *mysql.Config) (*sql.DB, error) {
	pool, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// set limit request, direct to db.
	switch config.Env.AppENV {
	case "production":
		pool.SetMaxOpenConns(60)
		pool.SetMaxIdleConns(10)
	case "debug":
		pool.SetMaxOpenConns(5)
		pool.SetMaxIdleConns(2)
	default:
		log.Fatalf("invalid value env: %s", config.Env.AppENV)
	}

	return pool, nil
}
