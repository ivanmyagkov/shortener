// Package main is the launch point of the link shortening application.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/middleware"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/workerpool"
)

//build and compile flags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

//	Structure of flags.
var flags struct {
	a string
	b string
	f string
	d string
}

//	envVar structure is struct of env variables.
var envVar struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN"`
}

//	init Initializing startup parameters.
func init() {
	// Build parameters
	switch buildVersion {
	case "":
		fmt.Printf("Build version: %s\n", "N/A")
	default:
		fmt.Printf("Build version: %s\n", buildVersion)
	}
	switch buildDate {
	case "":
		fmt.Printf("Build date: %s\n", "N/A")
	default:
		fmt.Printf("Build date: %s\n", buildDate)
	}
	switch buildCommit {
	case "":
		fmt.Printf("Build commit: %s\n", "N/A")
	default:
		fmt.Printf("Build commit: %s\n", buildCommit)
	}

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

//	main is entry point
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	var db interfaces.Storage

	cfg := config.NewConfig(flags.a, flags.b, flags.f, flags.d)
	var err error
	if cfg.FilePath() != "" {
		if db, err = storage.NewInFile(cfg.FilePath()); err != nil {
			log.Fatal(err)
		}
	} else if cfg.Database() != "" {
		db, err = storage.NewDB(cfg.Database())
		if err != nil {
			log.Fatalf("Failed to create db %e", err)
		}
	} else {
		db = storage.NewDBConn()
	}
	defer db.Close()

	//	Init Workers
	g, _ := errgroup.WithContext(ctx)
	recordCh := make(chan interfaces.Task, 50)
	doneCh := make(chan struct{})

	inWorker := workerpool.NewInputWorker(recordCh, doneCh, ctx)
	for i := 1; i <= runtime.NumCPU(); i++ {
		outWorker := workerpool.NewOutputWorker(i, recordCh, doneCh, ctx, db)
		g.Go(outWorker.Do)
	}

	g.Go(inWorker.Loop)

	usr := storage.New()
	mw := middleware.New(usr)
	srv := handlers.New(db, cfg, usr, inWorker)

	e := echo.New()
	pprof.Register(e)
	e.Use(middleware.CompressHandle)
	e.Use(middleware.Decompress)
	e.Use(mw.SessionWithCookies)
	e.GET("/:id", srv.GetURL)
	e.GET("/api/user/urls", srv.GetURLsByUserID)
	e.GET("/ping", srv.GetPing)
	e.POST("/", srv.PostURL)
	e.POST("/api/shorten", srv.PostJSON)
	e.POST("/api/shorten/batch", srv.PostBatch)
	e.DELETE("/api/user/urls", srv.DelURLsBATCH)

	go func() {

		<-signalChan

		log.Println("Shutting down...")

		cancel()
		if err = e.Shutdown(ctx); err != nil && err != ctx.Err() {
			e.Logger.Fatal(err)
		}

		if err = db.Close(); err != nil {
			log.Fatal(err)
		}

		close(recordCh)
		close(doneCh)
		err = g.Wait()
		if err != nil {
			log.Println(err)
		}

	}()

	if err := e.Start(cfg.SrvAddr()); err != nil {
		e.Logger.Fatal(err)
	}

}
