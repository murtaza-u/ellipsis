package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/console"
	"github.com/murtaza-u/ellipsis/api/me"
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/oidc"
	"github.com/murtaza-u/ellipsis/db"
	"github.com/murtaza-u/ellipsis/fs"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	conf.C
	app     *echo.Echo
	queries *sqlc.Queries
	fs      fs.Storage
}

func New(c conf.C) (*Server, error) {
	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	var conn *sql.DB

	if c.DB.Mysql.Enable {
		conn, err = db.NewMySQL(db.MySQLConfig{
			User:     c.DB.Mysql.User,
			Pass:     c.DB.Mysql.Password,
			Database: c.DB.Mysql.Database,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"failed to instantiate to mysql database: %w", err)
		}
	}

	if c.DB.Sqlite.Enable {
		conn, err = db.NewSqlite(db.SqliteConfig{
			Path: c.DB.Sqlite.Path,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"failed to instantiate to sqlite database: %w", err)
		}
	}

	if c.DB.Turso.Enable {
		conn, err = db.NewTurso(db.TursoConfig{
			Database:  c.DB.Turso.Database,
			AuthToken: c.DB.Turso.AuthToken,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"failed to instantiate to turso connection: %w", err)
		}
	}

	app := echo.New()
	app.Pre(echoMiddleware.RemoveTrailingSlash())
	app.Use(session.Middleware(sessions.NewCookieStore([]byte(c.SessionEncryptionKey))))

	// set custom error handler
	app.HTTPErrorHandler = apierr.Handler

	s3, err := fs.NewS3Store(c.S3.Region, c.S3.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3: %w", err)
	}

	opt := &slog.HandlerOptions{Level: slog.LevelInfo}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opt))
	if c.JsonLogger {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opt))
	}
	slog.SetDefault(logger)

	return &Server{
		C:       c,
		app:     app,
		queries: sqlc.New(conn),
		fs:      s3,
	}, nil
}

func (s Server) Start() error {
	err := Static(s.app)
	if err != nil {
		return fmt.Errorf("failed to serve static files: %w", err)
	}

	auth := middleware.NewAuthMiddleware(s.queries)

	s.app.GET("/signup", s.SignUpPage, auth.AlreadyAuthenticated)
	s.app.POST("/signup", s.SignUp, auth.AlreadyAuthenticated)
	s.app.GET("/login", s.LoginPage, auth.AlreadyAuthenticated)
	s.app.POST("/login", s.Login, auth.AlreadyAuthenticated)
	s.app.GET("/logout", s.Logout)

	// console
	console.New(s.queries).Register(s.app)

	// my account
	me.New(s.queries, s.Key, s.BaseURL, s.fs).Register(s.app)

	// oidc
	oidcAPI, err := oidc.New(oidc.Config{
		DB:        s.queries,
		Key:       s.Key,
		Providers: s.Providers,
		BaseURL:   s.BaseURL,
		FS:        s.fs,
	})
	if err != nil {
		return fmt.Errorf("failed to setup OIDC APIs: %w", err)
	}
	err = oidcAPI.Register(s.app)
	if err != nil {
		return fmt.Errorf("failed to register OIDC APIs: %w", err)
	}

	return s.app.Start(fmt.Sprintf(":%d", s.Port))
}
