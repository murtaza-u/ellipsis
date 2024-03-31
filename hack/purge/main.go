package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/murtaza-u/ellipsis/db"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
)

const defaultConfPath = "/etc/ellipsis/config.yaml"

var usage string

func init() {
	log.SetFlags(0)
	usage = fmt.Sprintf("%s [sessions|codes]", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("invalid number of arguments. Usage: %s", usage)
	}

	path := os.Getenv("ELLIPSIS_CONFIG")
	if path == "" {
		path = defaultConfPath
	}

	c, err := conf.New(path)
	if err != nil {
		log.Fatal(err)
	}

	var conn *sql.DB

	if c.DB.Mysql.Enable {
		conn, err = db.NewMySQL(db.MySQLConfig{
			User:     c.DB.Mysql.User,
			Pass:     c.DB.Mysql.Password,
			Database: c.DB.Mysql.Database,
		})
		if err != nil {
			log.Fatalf("failed to instantiate to mysql database: %s", err.Error())
		}
	}

	if c.DB.Sqlite.Enable {
		conn, err = db.NewSqlite(db.SqliteConfig{
			Path: c.DB.Sqlite.Path,
		})
		if err != nil {
			log.Fatalf("failed to instantiate to sqlite database: %s", err.Error())
		}
	}

	if c.DB.Turso.Enable {
		conn, err = db.NewTurso(db.TursoConfig{
			Database:  c.DB.Turso.Database,
			AuthToken: c.DB.Turso.AuthToken,
		})
		if err != nil {
			log.Fatalf("failed to instantiate to turso database: %s", err.Error())
		}
	}

	defer conn.Close()
	q := sqlc.New(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	switch os.Args[1] {
	case "sessions":
		err := q.DeleteExpiredSessions(ctx)
		if err != nil {
			log.Fatalf("failed to delete expired sessions: %s", err.Error())
		}
	case "codes":
		err := q.DeleteExpiredAuthzCode(ctx)
		if err != nil {
			log.Fatalf("failed to delete expired auth codes: %s", err.Error())
		}
	default:
		log.Fatalf("unknown argument. Usage: %s", usage)
	}
}
