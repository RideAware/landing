package main

import (
  "context"
  "log"
  "os"
  "os/signal"
  "syscall"

  "landing/internal/config"
  "landing/internal/database"
  "landing/internal/handlers"
)

func main() {
  // Load configuration
  cfg, err := config.LoadConfig()
  if err != nil {
    log.Fatalf("failed to load config: %v", err)
  }

  // Initialize database
  db, err := database.New(cfg)
  if err != nil {
    log.Fatalf("failed to connect to database: %v", err)
  }
  defer db.Close(context.Background())

  // Initialize database schema
  if err := db.InitDB(context.Background()); err != nil {
    log.Fatalf("failed to initialize database: %v", err)
  }

  // Create handler with dependencies
  h := handlers.New(db, cfg)

  // Start HTTP server
  go func() {
    log.Printf("starting server on %s:%s", cfg.Host, cfg.Port)
    if err := h.Start(cfg.Host, cfg.Port); err != nil {
      log.Printf("server error: %v", err)
    }
  }()

  // Graceful shutdown
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
  <-sigChan

  log.Println("shutting down server")
}