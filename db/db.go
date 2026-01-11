package db

import (
	"database/sql"
	"log"

	"github.com/perpus_backend/config"

	"github.com/go-sql-driver/mysql"
)

type MySQLStatus struct {
	Idle  int `json:"conn_idle"`
	InUse int `json:"conn_in_use"`

	OpenConnections    int `json:"current_open_conn"`
	MaxOpenConnections int `json:"max_open_conn"`

	WaitingConnection int64 `json:"current_wait_conn"`
}

func NewMySQLStorage(cfg *mysql.Config) *sql.DB {
	pool, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	switch config.Env.AppENV {
	case "production":
		pool.SetMaxOpenConns(60)
		pool.SetMaxIdleConns(10)
	case "debug":
		pool.SetMaxOpenConns(5)
		pool.SetMaxIdleConns(2)
	default:
		log.Fatalf("invalid value app_env: %s", config.Env.AppENV)
	}

	return pool
}

func HealthDBMySQL(db *sql.DB) MySQLStatus {
	return MySQLStatus{
		Idle:               db.Stats().Idle,
		InUse:              db.Stats().InUse,
		OpenConnections:    db.Stats().OpenConnections,
		WaitingConnection:  db.Stats().WaitCount,
		MaxOpenConnections: db.Stats().MaxOpenConnections,
	}
}
