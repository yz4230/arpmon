package db

import (
	"os"
	"strconv"

	"github.com/lib/pq"
)

func LoadConfigFromEnv() (pq.Config, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return pq.Config{}, err
	}

	return pq.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     uint16(port),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		SSLMode:  pq.SSLMode(os.Getenv("DB_SSL_MODE")),
	}, nil
}
