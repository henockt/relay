package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henockt/relay/internal/api"
	"github.com/henockt/relay/internal/config"
	"github.com/henockt/relay/internal/email"
	"github.com/henockt/relay/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}
	cfg := config.Load()

	// create db and stores
	db, err := store.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	userStore := store.NewUserStore(db)
	aliasStore := store.NewAliasStore(db)
	sender := email.NewSender(cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: api.NewServer(cfg, db, userStore, aliasStore, sender),
	}

	go func() {
		log.Printf("Server listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
