package main

import (
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	db := storage.NewDbConn()
	e := echo.New()
	e.GET("/:id", handlers.GetUrl(db))
	e.POST("/", handlers.PostUrl(db))
	e.Logger.Fatal(e.Start(":8080"))

}
