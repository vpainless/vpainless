package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"vpainless/api"
	hostingAdapter "vpainless/internal/access/adapter/hosting"
	accessRest "vpainless/internal/access/adapter/rest"
	accessStorage "vpainless/internal/access/adapter/storage"
	access "vpainless/internal/access/core"
	hostingRest "vpainless/internal/hosting/adapter/rest"
	hostingStorage "vpainless/internal/hosting/adapter/storage"
	"vpainless/internal/hosting/adapter/vpsprovider"
	hosting "vpainless/internal/hosting/core"
	"vpainless/internal/pkg/authz"
	internaldb "vpainless/internal/pkg/db"
	"vpainless/internal/pkg/log"
	"vpainless/pkg/middleware"
)

//go:embed default/startup.sh
var XrayInitScript string

type Config struct {
	MigrationsPath      string
	DBDir               string
	VpainlessPrivateKey string
	VpainlessPublicKey  string
}

func loadConfig() (*Config, error) {
	dbDir := os.Getenv("DB_DIR")
	if dbDir == "" {
		return nil, fmt.Errorf("missing db directory, set DB_DIR environment variable")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		return nil, fmt.Errorf("missing migrations path, set MIGRATIONS_PATH environment variable")
	}

	privateKeyPath := os.Getenv("VPAINLESS_PRIVATE_KEY")
	if privateKeyPath == "" {
		return nil, fmt.Errorf("missing path to the private key, set the VPAINLESS_PRIVATE_KEY environment variable")
	}

	publicKeyPath := os.Getenv("VPAINLESS_PUBLIC_KEY")
	if publicKeyPath == "" {
		return nil, fmt.Errorf("missing path to the public key, set the VPAINLESS_PUBLIC_KEY environment variable")
	}

	return &Config{
		MigrationsPath:      migrationsPath,
		DBDir:               dbDir,
		VpainlessPrivateKey: privateKeyPath,
		VpainlessPublicKey:  publicKeyPath,
	}, nil
}

func main() {
	logger := slog.New(&log.CustomHandler{
		Handler: slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}),
	})
	slog.SetDefault(logger)

	config, err := loadConfig()
	if err != nil {
		slog.Error("unable to load configs", "error", err)
		os.Exit(1)
	}

	dbPath := path.Join(config.DBDir, "access.db")
	accessDB, err := internaldb.OpenDB(dbPath)
	if err != nil {
		slog.Error("unable to open db", "path", dbPath, "error", err)
		os.Exit(1)
	}
	defer accessDB.Close()

	dbPath = path.Join(config.DBDir, "hosting.db")
	hostingDB, err := internaldb.OpenDB(dbPath)
	if err != nil {
		slog.Error("unable to open db", "path", dbPath, "error", err)
		os.Exit(1)
	}
	defer hostingDB.Close()

	if err := internaldb.ApplyMigrations(accessDB, config.MigrationsPath); err != nil {
		slog.Error("unable to apply migrations", "path", config.MigrationsPath, "error", err)
		os.Exit(1)
	}

	logger.Debug("Hello")

	host := url.URL{Scheme: "https", Host: "api.vultr.com"}
	vps := vpsprovider.NewVultr(host)
	hostingRepository := hostingStorage.NewRepository(hostingDB)

	systemKey := hosting.SSHKeyPair{
		Name:       "vpainless-key",
		PrivateKey: readFile(config.VpainlessPrivateKey),
		PublicKey:  readFile(config.VpainlessPublicKey),
	}

	defaultStartupScript := hosting.StartUpScript{
		Content: XrayInitScript,
	}

	hostingService := hosting.NewService(hostingRepository, vps, systemKey, defaultStartupScript)
	adapter := hostingAdapter.NewAdapter(hostingService)
	hostingRestAdapter := hostingRest.NewAdapter(hostingService)

	accessRepository := accessStorage.NewRepository(accessDB)
	accessService := access.NewService(accessRepository, adapter)
	accessRestAdapter := accessRest.NewAdapter(accessService)
	apiServer := api.NewServer(accessRestAdapter, hostingRestAdapter)

	r := http.NewServeMux()

	handler := api.HandlerWithOptions(apiServer, api.StdHTTPServerOptions{
		BaseURL:    "/api",
		BaseRouter: r,
		Middlewares: []api.MiddlewareFunc{
			api.MiddlewareFunc(authz.AuthenticationMiddleware(accessService, []middleware.Exclusion{
				{PathPrefix: "/api/users", Method: "POST"},
			})),
			api.MiddlewareFunc(middleware.BasicAuthMiddleware([]middleware.Exclusion{
				{PathPrefix: "/api/users", Method: "POST"},
			})),
			api.MiddlewareFunc(middleware.RequestIDMiddleware),
		},
	})

	startServer(context.Background(), logger, handler)
	logger.Info("Good Bye!")
}

func startServer(ctx context.Context, log *slog.Logger, handler http.Handler) {
	addr := "0.0.0.0:8080"
	s := &http.Server{
		Handler: middleware.CORSMiddleware(handler),
		Addr:    addr,
	}

	go func() {
		slog.Info("starting server ", "addr", addr)
		err := s.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Error("server listen error", "err", err)
			os.Exit(1)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Warn("error shutting down server", "error", err)
	}
}

func readFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		slog.Error("error opening file...", "path", path, "error", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		slog.Error("error reading file...", "path", path, "error", err)
		os.Exit(1)
	}

	return b
}
