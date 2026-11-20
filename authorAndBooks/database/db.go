package database

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/go-sql-driver/mysql"
)

// Config holds the MySQL connection settings.
type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

// ConfigFromEnv reads the connection settings from DB_USER, DB_PASSWORD, DB_HOST, DB_PORT and DB_NAME.
// Everything except the password has a local-development default.
func ConfigFromEnv() Config {
	return Config{
		User:     getEnv("DB_USER", "root"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		Name:     getEnv("DB_NAME", "test"),
	}
}

// Connect opens a connection pool and verifies the database is reachable.
func Connect(cfg Config) (*sql.DB, error) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	mysqlCfg.DBName = cfg.Name

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
