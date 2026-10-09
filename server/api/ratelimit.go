package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Config a rate limiter viselkedésének beállítására szolgál.
type Config struct {
	// RequestsPerMinute megadja, hogy egy IP-cím per perc 최대 몇 kérést tehet.
	RequestsPerMinute int

	// CleanupInterval a háttér goroutines időközét adja meg másodpercben,
	// ami után a nem aktív IP-címek törlődnek.
	CleanupInterval time.Duration

	// CleanupTolerance azt az időt adja meg, ami alatt egy IP-címnek
	// nem kell újabb kérést küldenie ahhoz, hogy ne torölődjön.
	CleanupTolerance time.Duration
}

// defaultConfig az ip-szerű értékekkel tér vissza.
func defaultConfig() Config {
	return Config{
		RequestsPerMinute: 30,
		CleanupInterval:  3 * time.Minute,
		CleanupTolerance: 5 * time.Minute,
	}
}

// visitor egyetlen IP-cím állapotát tárolja.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter a konkrét rate limiter objektum, amely szálbiztos módon
// kezeli a látogatókat.
type RateLimiter struct {
	visitors    map[string]*visitor
	mu          sync.RWMutex
	cfg         Config
	cleanupDone chan struct{}
}

// NewRateLimiter létrehoz egy új RateLimiter példányt az alapértelmezett,
// vagy a megadott konfigurációval.
func NewRateLimiter(cfg ...Config) *RateLimiter {
	c := defaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	rl := &RateLimiter{
		visitors:    make(map[string]*visitor),
		cfg:         c,
		cleanupDone: make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// getVisitor visszaadja a látogatót az IP-címhez, vagy létrehoz egyet,
// ha nem áll rendelkezésre.
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		// Átszámítjuk a percenkénti kérést másodpercenkéntire
		limit := rate.Limit(float64(rl.cfg.RequestsPerMinute) / 60.0)

		v = &visitor{
			limiter: rate.NewLimiter(
				limit,
				rl.cfg.RequestsPerMinute, // Ez a "burst" (hirtelen beeső) kérések maximális száma
			),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v
}

// Allow IP-cím alapján dönti el, hogy a kérést fogadja-e a limiter,
// vagy elutasítja a korlátozás miatt.
func (rl *RateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)
	return v.limiter.Allow()
}

// cleanupCycle végrehajt egyetlen takarítási ciklust:
// kitörli azokat az IP-címeket, amelyek utolsó látogatás után
// eltelt CleaningTolerance időszak alatt nem küldtek kérést.
func (rl *RateLimiter) cleanupCycle() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-rl.cfg.CleanupTolerance)

	for ip, v := range rl.visitors {
		if v.lastSeen.Before(threshold) {
			delete(rl.visitors, ip)
		}
	}
}

// cleanupLoop a háttér goroutines, amely rendszeres időközönként
// meghívja a cleanupCycle-et a memória terhelés csökkentése érdekében.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupCycle()
		case <-rl.cleanupDone:
			return
		}
	}
}

// Stop leállítja a takarítási goroutines-et, és biztosítja az
// erőforrások felszabadítást.
func (rl *RateLimiter) Stop() {
	close(rl.cleanupDone)
}

// HandlerMiddleware HTTP middleware, amely a kérés IP-címét kéri le,
// és ha a limiter elutasítja, a 429 Too Many Requests válaszkal tér vissza
// a megfelelő header-kal együtt.
func (rl *RateLimiter) HandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Ha több IP van, az első az eredeti kliens
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		} else {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
		}

		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Minute).UTC().Format(time.RFC1123))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 Too Many Requests"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
