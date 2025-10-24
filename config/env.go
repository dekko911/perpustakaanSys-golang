package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppURL, Port, CookieName, CookieValue, DBUser, DBPassword, DBName, DBAddress, JWTSecret string
	DBLoc                                                                                   *time.Location
}

var Env = initConfig()

func initConfig() Config {
	godotenv.Load()

	loc, err := time.LoadLocation(os.Getenv("DB_LOC"))
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		AppURL:      os.Getenv("APP_URL"),
		Port:        os.Getenv("PORT"),
		CookieName:  os.Getenv("COOKIE_NAME"),
		CookieValue: os.Getenv("COOKIE_VALUE"),
		DBUser:      os.Getenv("DB_USERNAME"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_DATABASE"),
		DBAddress:   fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		DBLoc:       loc,
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}
