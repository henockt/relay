package main

import (
	"log"
	"net/http"

	"github.com/henockt/relay/internal/api"
	"github.com/henockt/relay/internal/config"
	"github.com/henockt/relay/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("no .env file found")
	}
	cfg := config.Load()

	// create db and stores
	db, err := store.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	userStore := store.NewUserStore(db)
	aliasStore := store.NewAliasStore(db)

	// create server
	srv := api.NewServer(cfg, userStore, aliasStore)
	addr := ":" + cfg.Port
	log.Printf("Server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
