// Command server runs the auth backend HTTP server.
//
// Configuration is read from environment variables (see config package).
// No secrets are hardcoded.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"minecraftserverluncher/server/api"
	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/skins"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Skin handling wire-up.
	skinStorage := skins.NewThreadSafeSkinHistory()
	rateLimiter := api.NewRateLimiter()
	skinHandler := &api.SkinHandler{
		SkinStorageDir: "./storage/skins",
		History:        skinStorage,
		RateLimiter:    rateLimiter,
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           api.NewServer(db, cfg, skinHandler).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("auth server listening on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
