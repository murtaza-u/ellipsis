package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func NewMySQL(c MySQLConfig) (*sql.DB, error) {
	err := c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate mysql config: %w", err)
	}
	url := fmt.Sprintf("%s:%s@/%s?parseTime=true", c.User, c.Pass, c.Database)
	return sql.Open("mysql", url)
}

func NewSqlite(c SqliteConfig) (*sql.DB, error) {
	err := c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate sqlite config: %w", err)
	}
	url := fmt.Sprintf("file:%s", c.Path)
	return sql.Open("sqlite3", url)
}
