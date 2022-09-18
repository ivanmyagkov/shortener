// Package main is the launch point of the link shortening application.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/caarlos0/env/v6"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/ivanmyagkov/shortener.git/internal/config"
	grpc2 "github.com/ivanmyagkov/shortener.git/internal/grpc"
	handlers2 "github.com/ivanmyagkov/shortener.git/internal/grpc/handlers"
	"github.com/ivanmyagkov/shortener.git/internal/grpc/interceptors"
	pb "github.com/ivanmyagkov/shortener.git/internal/grpc/proto"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
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
	mw := interceptors.New(usr)
	handler := handlers2.NewGRPCHandler(db, cfg, usr, inWorker)
	server, err := grpc2.NewGRPCServer(handler)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(mw.UserIDInterceptor))
	// register a service
	pb.RegisterShortenerServer(s, server)
	listen, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Server start")
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
	// set a listener for os.Signal
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-signalChan

		log.Println("Shutting down...")
		s.GracefulStop()
		if err = db.Close(); err != nil {
			log.Fatal(err)
		}

		cancel()
		close(recordCh)
		close(doneCh)
		err = g.Wait()
		if err != nil {
			log.Println(err)
		}
	}()

	log.Print("Server shutdown")

}
