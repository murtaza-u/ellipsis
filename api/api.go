package api

import "github.com/labstack/echo/v4"

type Server struct {
	app *echo.Echo
}

func New() *Server {
	return &Server{
		app: echo.New(),
	}
}

func (s Server) Start() error {
	s.app.Static("/static", "static")
	s.app.GET("/", s.index)
	return s.app.Start(":3000")
}
