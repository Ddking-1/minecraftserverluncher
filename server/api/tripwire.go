package api

import (
	"net"
	"net/http"
	"strings"
)

// recordTripwire persists a tamper/anomaly event and logs it. It never fails
// the request; tripwire recording is best-effort.
func (s *Server) recordTripwire(r *http.Request, kind, detail string) {
	if s == nil || s.db == nil {
		return
	}
	ip := clientIPForTripwire(r)
	accountID := RequestedAccountID(r)
	if _, err := s.db.RecordTripwire(kind, detail, ip, accountID); err != nil {
		// Best effort: log and continue.
		_ = err
	}
}

// clientIP returns the address used for security decisions (login IP history,
// IP bans). Forwarding headers are attacker-controlled unless a trusted proxy
// sets them, so they are only honoured when AUTH_TRUST_PROXY is enabled, using
// the same right-most X-Forwarded-For rule as the rate limiter. Without this,
// a backend behind a reverse proxy records the proxy address for every login,
// which would make an IP ban silently hit every account.
func (s *Server) clientIP(r *http.Request) string {
	if s != nil && s.cfg != nil && s.cfg.TrustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				if ip := strings.TrimSpace(parts[i]); ip != "" {
					return ip
				}
			}
		}
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
	}
	return clientIPForTripwire(r)
}

func clientIPForTripwire(r *http.Request) string {
	if r == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		host = r.RemoteAddr
	}
	return strings.TrimSpace(host)
}
