package api

import (
	"fmt"
	"time"

	"github.com/murtaza-u/account/api/console"
	"github.com/murtaza-u/account/api/oidc"
	"github.com/murtaza-u/account/db"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/murtaza-u/dream"
)

type Server struct {
	Config
	app     *echo.Echo
	queries *sqlc.Queries
	dream   *dream.Store
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

	app := echo.New()
	app.Use(middleware.RemoveTrailingSlash())

	return &Server{
		Config:  c,
		app:     app,
		queries: sqlc.New(db),
		dream:   dream.New(dream.WithCleanUp(time.Minute)),
	}, nil
}

func (s Server) Start() error {
	s.app.Static("/static", "static")
	s.app.GET("/", s.indexPage)
	s.app.GET("/signup", s.SignUpPage)
	s.app.POST("/signup", s.SignUp)
	s.app.GET("/login", s.LoginPage)
	s.app.POST("/login", s.Login)

	// console
	console.New(s.queries).Register(s.app)

	// oidc
	oidc.New(s.queries, s.dream).Register(s.app)

	return s.app.Start(fmt.Sprintf(":%d", s.Port))
}
