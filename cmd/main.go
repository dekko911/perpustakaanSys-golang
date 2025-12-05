package main

import (
	"database/sql"
	"fmt"
	"log"
	"perpus_backend/cmd/api"
	"perpus_backend/config"
	"perpus_backend/db"
	"runtime"
	"runtime/debug"

	"github.com/go-sql-driver/mysql"
)

func initDBStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database Connected!")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	switch config.Env.AppENV {
	case "production":
		debug.SetGCPercent(200)
	case "debug":
		debug.SetGCPercent(50)
	default:
		log.Fatalf("invalid app env: %s", config.Env.AppENV)
	}

	db, err := db.NewMySQLStorage(&mysql.Config{
		User:                 config.Env.DBUser,
		Passwd:               config.Env.DBPassword,
		Addr:                 config.Env.DBAddress,
		DBName:               config.Env.DBName,
		Loc:                  config.Env.DBLoc,
		Net:                  "tcp",
		ParseTime:            true,
		AllowNativePasswords: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	initDBStorage(db)

	s := api.NewAPIServer(fmt.Sprintf(":%s", config.Env.Port), db)
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
