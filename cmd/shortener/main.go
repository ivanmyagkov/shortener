package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
	envVar := config.EnvVar{}
	err := env.Parse(&envVar)
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.NewConfig(envVar.ServerAddress, envVar.BaseURL)

	db := storage.NewDBConn()
	srv := handlers.New(db, cfg)
	e := echo.New()
	e.GET("/:id", handlers.GetURL(srv))
	e.POST("/", handlers.PostURL(srv))
	e.POST("/api/shorten", handlers.PostJSON(srv))

	if err := e.Start(cfg.SrvAddr()); err != nil {
		e.Logger.Fatal(err)
	}

}
