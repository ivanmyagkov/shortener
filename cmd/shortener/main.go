package main

import (
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	db := storage.NewDBConn()
	srv := handlers.New(db)
	e := echo.New()
	e.GET("/:id", handlers.GetURL(srv))
	e.POST("/", handlers.PostURL(srv))
	e.POST("/api/shorten", handlers.PostJSON(srv))
	e.Logger.Fatal(e.Start(":8080"))

}
