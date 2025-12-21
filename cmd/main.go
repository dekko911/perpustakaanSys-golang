package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/go-sql-driver/mysql"
	"github.com/perpus_backend/cmd/api"
	"github.com/perpus_backend/config"
	"github.com/perpus_backend/db"

	"github.com/redis/go-redis/v9"
)

// package variables (var) -> func init() -> func main()
var (
	mysqlDB *sql.DB
	redisDB *redis.Client
)

func init() {
	log.Println("Setup databases connection...")

	mysqlDB = db.NewMySQLStorage(&mysql.Config{
		User:                 config.Env.DBUser,
		Passwd:               config.Env.DBPassword,
		Addr:                 config.Env.DBAddress,
		DBName:               config.Env.DBName,
		Loc:                  config.Env.DBLoc,
		Net:                  "tcp",
		ParseTime:            true,
		AllowNativePasswords: true,
	})

	redisDB = redis.NewClient(&redis.Options{
		Addr:       config.Env.RedisAddress,
		ClientName: config.Env.RedisClient,
		Password:   config.Env.RedisPassword,
		DB:         0,
	})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	defer mysqlDB.Close() // <- just set mysqlDB close func for safety reason.

	defer redisDB.Close() // <- just set redisDB close func for safety reason.

	switch config.Env.AppENV {
	case "production":
		debug.SetGCPercent(200)
	case "debug":
		debug.SetGCPercent(50)
	default:
		log.Fatalf("invalid app env: %s", config.Env.AppENV)
	}

	ctx := context.Background()

	pingMysqlDB(ctx, mysqlDB)

	pingRedisDB(ctx, redisDB)

	s := api.NewAPIServer(fmt.Sprintf(":%s", config.Env.Port), mysqlDB, redisDB)
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

func pingMysqlDB(ctx context.Context, db *sql.DB) {
	err := db.PingContext(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mysql Connected!")
}

func pingRedisDB(ctx context.Context, rdb *redis.Client) {
	ping, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Redis Connected: %s", ping)
}
