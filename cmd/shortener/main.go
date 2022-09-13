// Package main is the launch point of the link shortening application.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	"github.com/ivanmyagkov/shortener.git/internal/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"github.com/ivanmyagkov/shortener.git/internal/middleware"
	"github.com/ivanmyagkov/shortener.git/internal/storage"
	"github.com/ivanmyagkov/shortener.git/internal/workerpool"
)

//	Structure of flags.
var flags struct {
	A string `json:"server_address"`
	B string `json:"base_url"`
	F string `json:"file_storage_path"`
	D string `json:"database_dns"`
	S bool   `json:"enable_https"`
	C string `json:"-"`
	T string `json:"trusted_subnet"`
}

//	envVar structure is struct of env variables.
var envVar struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Database        string `env:"DATABASE_DSN" json:"database_dns"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	Config          string `env:"CONFIG" json:"-"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"  envDefault:"192.168.1.0/24" json:"trusted_subnet"`
}

//build and compile flags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func buildParams() {
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
}

//	init Initializing startup parameters.
func init() {
	buildParams()
	err := env.Parse(&envVar)
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&flags.A, "a", envVar.ServerAddress, "server address")
	flag.StringVar(&flags.B, "b", envVar.BaseURL, "base url")
	flag.StringVar(&flags.F, "f", envVar.FileStoragePath, "file storage path")
	flag.StringVar(&flags.D, "d", envVar.Database, "database path")
	flag.BoolVar(&flags.S, "s", envVar.EnableHTTPS, "enable ssl")
	flag.StringVar(&flags.C, "config", envVar.Config, "config file")
	flag.StringVar(&flags.C, "c", envVar.Config, "config file")
	flag.StringVar(&flags.T, "t", envVar.TrustedSubnet, "config file")
	flag.Parse()
	config.ParseConfig(flags.C, &flags)
}

//	main is entry point
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	var db interfaces.Storage

	cfg := config.NewConfig(flags.A, flags.B, flags.F, flags.D, flags.S, flags.T)
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
	e.GET("/api/internal/stats", srv.GetStats)
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("cache-dir"),
		HostPolicy: autocert.HostWhitelist("mysite.ru"),
	}
	s := http.Server{
		Addr:      cfg.SrvAddr(),
		Handler:   e, // set Echo as handler
		TLSConfig: m.TLSConfig(),
	}
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {

		<-signalChan

		log.Println("Shutting down...")

		if cfg.EnableHTTPS {
			if err = s.Shutdown(ctx); err != nil && err != ctx.Err() {
				s.ErrorLog.Fatal(ctx)
			}
		} else {
			if err = e.Shutdown(ctx); err != nil && err != ctx.Err() {
				e.Logger.Fatal(err)
			}
		}

		if err = db.Close(); err != nil {
			log.Fatal(err)
		}

		err = g.Wait()
		if err != nil {
			log.Println(err)
		}
		cancel()
		close(recordCh)
		close(doneCh)

	}()

	if !cfg.EnableHTTPS {
		if err = e.Start(cfg.SrvAddr()); err != nil {
			e.Logger.Fatal(err)
		}
	} else {
		if err = s.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			s.ErrorLog.Fatal(err)
		}
	}

}
