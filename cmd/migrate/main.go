package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"perpus_backend/config"
	"perpus_backend/db"

	mysqlCfg "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := db.NewMySQLStorage(&mysqlCfg.Config{
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

	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer m.Close()

	command := os.Args[(len(os.Args) - 1)]
	if command == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}

		log.Println("Migration success created!")
	}

	if command == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}

		dirPath := "./assets"
		err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return os.Remove(path)
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		log.Println("Migration success deleted!")
	}
}
