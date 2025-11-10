package main

import (
	"database/sql"
	"fmt"
	"log"
	"perpus_backend/cmd/api"
	"perpus_backend/config"
	"perpus_backend/db"

	"github.com/go-sql-driver/mysql"
)

func initDBStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Println("Database Connected!")
}

func main() {
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
