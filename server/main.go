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
	"minecraftserverluncher/server/pterodactyl"
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

	// Processed skins (and their avatars) are persisted so /api/v1/skin-files
	// URLs stay valid across restarts.
	skinStorage, err := skins.NewDiskStorage(cfg.SkinStorageDir)
	if err != nil {
		log.Fatalf("skin storage: %v", err)
	}
	skinRateLimiter := api.NewRateLimiter()
	defer skinRateLimiter.Stop()
	skinHandler := &api.SkinHandler{
		Storage:     skinStorage,
		History:     skins.NewThreadSafeSkinHistory(),
		RateLimiter: skinRateLimiter,
	}

	apiServer := api.NewServer(db, cfg, skinHandler)
	defer apiServer.Close()

	// The webhook is signature-verified by the handler; a missing
	// PTERODACTYL_WEBHOOK_SECRET makes it answer 401 instead of failing open.
	apiServer.PterodactylWebhook = &api.PterodactylWebhook{
		Handler: func(payload []byte) {
			event, err := pterodactyl.ParseServerEvent(payload)
			if err != nil {
				log.Printf("event=pterodactyl_webhook_invalid_payload err=%v", err)
				return
			}
			log.Printf("event=%s source=pterodactyl", event.Event)
		},
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           apiServer.Handler(),
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
