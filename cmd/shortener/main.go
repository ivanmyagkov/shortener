package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

var flags struct {
	a string
	b string
	f string
}

var envVar struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func init() {
	err := env.Parse(&envVar)
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&flags.a, "a", envVar.ServerAddress, "server address")
	flag.StringVar(&flags.b, "b", envVar.BaseURL, "base url")
	flag.StringVar(&flags.f, "f", envVar.FileStoragePath, "file storage path")
}

func main() {
	var db interfaces.Storage
	var err error

	flag.Parse()
	cfg := config.NewConfig(flags.a, flags.b, flags.f)

	if cfg.FilePath() != "" {
		db, err = storage.NewInFile(cfg.FilePath())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db = storage.NewDBConn()
	}
	defer db.Close()

	srv := handlers.New(db, cfg)

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Decompress())

	e.GET("/:id", handlers.GetURL(srv))
	e.POST("/", handlers.PostURL(srv))
	e.POST("/api/shorten", handlers.PostJSON(srv))

	if err := e.Start(cfg.SrvAddr()); err != nil {
		e.Logger.Fatal(err)
	}

}
