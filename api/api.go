package api

import (
	"fmt"
	"time"

	"github.com/murtaza-u/account/api/console"
	"github.com/murtaza-u/account/api/me"
	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/oidc"
	"github.com/murtaza-u/account/db"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/murtaza-u/dream"
)

type Server struct {
	Config
	app     *echo.Echo
	queries *sqlc.Queries
	cache   *dream.Store
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
	app.Pre(echoMiddleware.RemoveTrailingSlash())

	return &Server{
		Config:  c,
		app:     app,
		queries: sqlc.New(db),
		cache:   dream.New(dream.WithCleanUp(time.Minute)),
	}, nil
}

func (s Server) Start() error {
	s.app.Static("/static", "static")
	s.app.GET("/", s.indexPage)

	auth := middleware.NewAuthMiddleware(s.queries)

	s.app.GET("/signup", s.SignUpPage, auth.AlreadyAuthenticated)
	s.app.POST("/signup", s.SignUp, auth.AlreadyAuthenticated)
	s.app.GET("/login", s.LoginPage, auth.AlreadyAuthenticated)
	s.app.POST("/login", s.Login, auth.AlreadyAuthenticated)
	s.app.GET("/logout", s.Logout)

	// console
	console.New(s.queries).Register(s.app)

	// oidc
	oidcAPI, err := oidc.New(s.queries, s.cache, s.KeyStore)
	if err != nil {
		return fmt.Errorf("failed to setup OIDC APIs: %w", err)
	}
	oidcAPI.Register(s.app)

	me.New(s.queries).Register(s.app)

	return s.app.Start(fmt.Sprintf(":%d", s.Port))
}
