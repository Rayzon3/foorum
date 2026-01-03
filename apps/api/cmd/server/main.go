package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"jabber_v3/apps/api/internal/auth"
	"jabber_v3/apps/api/internal/config"
	"jabber_v3/apps/api/internal/http/routes"
	"jabber_v3/apps/api/internal/store"
)

func main() {
  cfg := config.Load()

  db, err := sql.Open("pgx", cfg.DatabaseURL)
  if err != nil {
    log.Fatalf("db open failed: %v", err)
  }
  defer db.Close()

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  if err := db.PingContext(ctx); err != nil {
    log.Fatalf("db ping failed: %v", err)
  }

  appStore := store.New(db)
  jwtManager := auth.JWTManager{Secret: []byte(cfg.JWTSecret), TTL: 24 * time.Hour}

  httpServer := &http.Server{
    Addr:         ":" + cfg.Port,
    Handler:      routes.NewRouter(appStore, jwtManager, cfg.CorsOrigins),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
  }

  go func() {
    log.Printf("** server listening on %s **", httpServer.Addr)
    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
      log.Fatalf("server error: %v", err)
    }
  }()

  quit := make(chan os.Signal, 1)
  signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
  <-quit

  shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer shutdownCancel()
  if err := httpServer.Shutdown(shutdownCtx); err != nil {
    log.Printf("shutdown error: %v", err)
  }
}
