package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppENV, AppURL, ClientPort, CookieName, CookieValue, DBUser, DBPassword, DBName, DBAddress, LocalAddress, MeilisearchURL, MSApiKey, Port, RedisAddress, RedisClient, RedisPassword, JWTSecret, SessionDomain string

	DBLoc *time.Location
}

var Env = initConfig()

func initConfig() *Config {
	_ = godotenv.Load()

	loc, err := time.LoadLocation(getENVConfigValue("DB_LOC"))
	if err != nil {
		log.Fatal(err)
	}

	return &Config{
		AppENV:         getENVConfigValue("APP_ENV"),
		AppURL:         getENVConfigValue("APP_URL"),
		ClientPort:     getENVConfigValue("CLIENT_PORT"),
		CookieName:     getENVConfigValue("COOKIE_NAME"),
		CookieValue:    getENVConfigValue("COOKIE_VALUE"),
		DBUser:         getENVConfigValue("DB_USERNAME"),
		DBPassword:     getENVConfigValue("DB_PASSWORD"),
		DBName:         getENVConfigValue("DB_DATABASE"),
		DBAddress:      fmt.Sprintf("%s:%s", getENVConfigValue("DB_HOST"), getENVConfigValue("DB_PORT")),
		DBLoc:          loc,
		LocalAddress:   fmt.Sprintf("%s:%s", getENVConfigValue("APP_URL"), getENVConfigValue("CLIENT_PORT")),
		MeilisearchURL: getENVConfigValue("MEILISEARCH_URL"),
		MSApiKey:       getENVConfigValue("MS_API_KEY"),
		Port:           getENVConfigValue("PORT"),
		RedisAddress:   fmt.Sprintf("%s:%s", getENVConfigValue("REDIS_HOST"), getENVConfigValue("REDIS_PORT")),
		RedisClient:    getENVConfigValue("REDIS_CLIENT"),
		RedisPassword:  getENVConfigValue("REDIS_PASSWORD"),
		JWTSecret:      getENVConfigValue("JWT_SECRET"),
		SessionDomain:  getENVConfigValue("SESSION_DOMAIN"),
	}
}

// get value on file env, and check one by one variable at param to get the value.
func getENVConfigValue(variable string) string {
	v, ok := os.LookupEnv(variable)
	if !ok {
		return ""
	}

	return v
}
