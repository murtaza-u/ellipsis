package api

import (
	"fmt"

	"github.com/murtaza-u/account/api/console"
	"github.com/murtaza-u/account/api/oidc"
	"github.com/murtaza-u/account/db"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/labstack/echo/v4"
)

type Server struct {
	Config
	app     *echo.Echo
	queries *sqlc.Queries
}

func New(c Config) (*Server, error) {
	err := c.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	db, err := db.New(db.Config{
		User:     c.DatabaseUser,
		Pass:     c.DatabasePassword,
		Database: c.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Server{
		Config:  c,
		app:     echo.New(),
		queries: sqlc.New(db),
	}, nil
}

func (s Server) Start() error {
	s.app.Static("/static", "static")
	s.app.GET("/", s.indexPage)
	console.New(s.queries).Register(s.app)
	oidc.New(s.queries).Register(s.app)
	return s.app.Start(fmt.Sprintf(":%d", s.Port))
}
