package api

import (
	"fmt"
	"time"

	"github.com/murtaza-u/account/api/console"
	"github.com/murtaza-u/account/api/me"
	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/oidc"
	"github.com/murtaza-u/account/db"
	"github.com/murtaza-u/account/internal/conf"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/murtaza-u/dream"
)

type Server struct {
	conf.C
	app     *echo.Echo
	queries *sqlc.Queries
	cache   *dream.Store
}

func New(c conf.C) (*Server, error) {
	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	db, err := db.New(db.Config{
		User:     c.Mysql.User,
		Pass:     c.Mysql.Password,
		Database: c.Mysql.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app := echo.New()
	app.Pre(echoMiddleware.RemoveTrailingSlash())
	app.Use(session.Middleware(sessions.NewCookieStore([]byte(c.SessionEncryptionKey))))

	return &Server{
		C:       c,
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
	oidcAPI, err := oidc.New(oidc.Config{
		DB:        s.queries,
		Cache:     s.cache,
		KeyStore:  s.KeyStore,
		Providers: s.Providers,
	})
	if err != nil {
		return fmt.Errorf("failed to setup OIDC APIs: %w", err)
	}
	err = oidcAPI.Register(s.app)
	if err != nil {
		return fmt.Errorf("failed to register OIDC APIs: %w", err)
	}

	// my account
	me.New(s.queries).Register(s.app)

	return s.app.Start(fmt.Sprintf(":%d", s.Port))
}
