package main

import (
	"flag"
	"log"

	_ "github.com/lib/pq"

	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/middleware"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
)

var flags struct {
	a string
	b string
	f string
	d string
}

var envVar struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN" envDefault:"postgres://ivanmyagkov@localhost:5432/postgres?sslmode=disable"`
}

func init() {
	err := env.Parse(&envVar)
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&flags.a, "a", envVar.ServerAddress, "server address")
	flag.StringVar(&flags.b, "b", envVar.BaseURL, "base url")
	flag.StringVar(&flags.f, "f", envVar.FileStoragePath, "file storage path")
	flag.StringVar(&flags.d, "d", envVar.Database, "database path")
	flag.Parse()
}

func main() {
	var db interfaces.Storage

	cfg := config.NewConfig(flags.a, flags.b, flags.f, flags.d)

	if cfg.FilePath() != "" {
		var err error
		if db, err = storage.NewInFile(cfg.FilePath()); err != nil {
			log.Fatal(err)
		}
	} else if cfg.Database() != "" {
		db = storage.NewDb(cfg.Database())
	} else {
		db = storage.NewDBConn()
	}
	defer db.Close()
	usr := storage.New()
	mw := middleware.New(usr)
	srv := handlers.New(db, cfg, usr)

	e := echo.New()
	e.Use(middleware.CompressHandle)
	e.Use(middleware.Decompress)
	e.Use(mw.SessionWithCookies)

	e.GET("/:id", srv.GetURL)
	e.GET("/api/user/urls", srv.GetURLsByUserID)
	e.GET("/ping", srv.GetPing)
	e.POST("/", srv.PostURL)
	e.POST("/api/shorten", srv.PostJSON)

	if err := e.Start(cfg.SrvAddr()); err != nil {
		e.Logger.Fatal(err)
	}

}
