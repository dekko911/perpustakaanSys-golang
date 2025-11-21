package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppENV, AppURL, ClientPort, Port, CookieName, CookieValue, DBUser, DBPassword, DBName, DBAddress, JWTSecret, SessionDomain string
	DBLoc                                                                                                                      *time.Location
}

var Env = initConfig()

func initConfig() *Config {
	_ = godotenv.Load()

	loc, err := time.LoadLocation(os.Getenv("DB_LOC"))
	if err != nil {
		log.Fatal(err)
	}

	return &Config{
		AppENV:        getAppENV(),
		AppURL:        os.Getenv("APP_URL"),
		ClientPort:    os.Getenv("CLIENT_PORT"),
		Port:          os.Getenv("PORT"),
		CookieName:    os.Getenv("COOKIE_NAME"),
		CookieValue:   os.Getenv("COOKIE_VALUE"),
		DBUser:        os.Getenv("DB_USERNAME"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_DATABASE"),
		DBAddress:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		DBLoc:         loc,
		JWTSecret:     os.Getenv("JWT_SECRET"),
		SessionDomain: os.Getenv("SESSION_DOMAIN"),
	}
}

func getAppENV() string {
	if v, ok := os.LookupEnv("APP_ENV"); ok {
		return v
	}

	return "debug" // set to "debug" if os.LookupEnv() doesn't read the actual value
}
