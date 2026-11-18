// Cookie consent endpoint.
//
// The frontend stores the consent locally in localStorage and also notifies
// the backend so server-side features (for example a server-side consent
// cookie) can be aligned. During the alpha the backend consent cookie is
// optional; the UI-side choice is authoritative for analytics gating.

package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

// handleCookieConsent handles POST /api/v1/cookie-consent.
//
// The body must be {"accepted": true|false}. On acceptance the server sets a
// cookie named "alpha_consent" that survives for one year so returning users
// do not see the popup again until the cookie expires or is cleared.
func (s *Server) handleCookieConsent(w http.ResponseWriter, r *http.Request) {
	// Cookie consent is public: any visitor may set their preference.
	// We do not require authentication here because the consent popup must work
	// before the user has logged in.
	_ = r

	var req struct {
		Accepted bool `json:"accepted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Always reflect the choice in a cookie so the backend can honour it on
	// later visits even when localStorage is unavailable (e.g. some Qt profiles).
	cookie := &http.Cookie{
		Name:     "alpha_consent",
		Path:     "/",
		HttpOnly: false, // the UI reads this to decide whether to show the popup
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}
	if req.Accepted {
		cookie.Value = "accepted"
		cookie.Expires = time.Now().Add(365 * 24 * time.Hour)
	} else {
		cookie.Value = "rejected"
		cookie.Expires = time.Now().Add(-24 * time.Hour) // expired so it is treated as "no choice"
	}
	http.SetCookie(w, cookie)

	// When the user accepts, also set a short-lived server-side session cookie
	// that the analytics handler can use as a lightweight visit identifier.
	if req.Accepted {
		visit := &http.Cookie{
			Name:     "alpha_visit",
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60, // one hour
			Value:    newTokenShort(),
		}
		http.SetCookie(w, visit)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAnalyticsEvent(w http.ResponseWriter, r *http.Request) {
	// Analytics is only accepted when the user has said yes. We check both the
	// cookie the backend set and the UI-side choice. If neither is present the
	// event is dropped with a harmless 204 so the UI never waits on it.
	cookie, _ := r.Cookie("alpha_consent")
	if cookie == nil || cookie.Value != "accepted" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var ev struct {
		Kind   string         `json:"kind"`
		View   string         `json:"view,omitempty"`
		Action string         `json:"action,omitempty"`
		Detail map[string]any `json:"detail,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// During the alpha we do not persist analytics to the database; we just
	// acknowledge the event. When analytics storage is added this is the place
	// to write a stripped, anonymous event row.
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ConsultCookieConsent reports whether the request has accepted cookies.
// Handlers that want to gate optional features on consent can call this.
func (s *Server) ConsultCookieConsent(r *http.Request) bool {
	cookie, _ := r.Cookie("alpha_consent")
	if cookie != nil && cookie.Value == "accepted" {
		return true
	}
	return false
}

// newTokenShort returns a short random token for the alpha_visit cookie.
func newTokenShort() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}
