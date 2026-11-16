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
EOF
echo done