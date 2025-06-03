package main

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vpainless/api"
	"vpainless/pkg/middleware"
)

func main() {
	r := http.NewServeMux()

	apiServer := &MockServer{}
	handler := api.HandlerWithOptions(apiServer, api.StdHTTPServerOptions{
		BaseURL:    "/api",
		BaseRouter: r,
	})

	addr := "0.0.0.0:8080"
	s := &http.Server{
		Handler: middleware.CORSMiddleware(handler),
		Addr:    addr,
	}

	go func() {
		slog.Info("starting server ", "addr", addr)
		err := s.ListenAndServe()
		if err != http.ErrServerClosed {
			slog.Error("server listen error", "err", err)
			os.Exit(1)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		slog.Warn("error shutting down server", "error", err)
	}
}
