package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	var driverName string
	var connStr string

	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		driverName = "postgres"
		connStr = databaseURL
	} else if strings.HasPrefix(databaseURL, "sqlite://") {
		driverName = "sqlite3"
		connStr = strings.TrimPrefix(databaseURL, "sqlite://")
	} else {
		// Default to sqlite3 for backward compatibility
		driverName = "sqlite3"
		connStr = databaseURL
	}

	sqlDB, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if driverName == "postgres" {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
	} else {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{sqlDB}, nil
}
