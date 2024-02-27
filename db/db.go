package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func New(c Config) (*sql.DB, error) {
	err := c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	url := fmt.Sprintf("%s:%s@/%s?parseTime=true", c.User, c.Pass, c.Database)
	return sql.Open("mysql", url)
}
