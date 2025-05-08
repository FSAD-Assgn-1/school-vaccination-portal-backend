package server

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Start() {
	router := newRouter()

	if router == nil {
		log.Println("Router Not Initialized")
		return
	}
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		resp := c.Response()
		router.ServeHTTP(resp, req)
		return
	})
	e.Logger.Fatal(e.Start(":8080"))
}
