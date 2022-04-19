package main

import (
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	db := storage.NewDBConn()
	e := echo.New()
	e.GET("/:id", handlers.GetURL(db))
	e.POST("/", handlers.PostURL(db))
	e.Logger.Fatal(e.Start(":8080"))

}
