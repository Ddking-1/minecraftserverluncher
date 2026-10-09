package api

import (
	"net/http"
	"os"

	"minecraftserverluncher/server/pterodactyl"
)

// PterodactylWebhook wraps the pterodactyl webhook handler with the secret from
// the environment so it can be plugged into the API mux.
type PterodactylWebhook struct {
	// Handler is the inner event handler called for verified payloads.
	Handler pterodactyl.WebhookHandleFunc
}

// Handle returns an http.HandlerFunc bound to the configured secret.
func (w *PterodactylWebhook) Handle() http.HandlerFunc {
	secret := os.Getenv("PTERODACTYL_WEBHOOK_SECRET")
	if secret == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "missing webhook secret", http.StatusUnauthorized)
		})
	}
	cfg := pterodactyl.DefaultWebhookConfig(secret, w.Handler)
	return pterodactyl.HandleWebhook(cfg)
}
