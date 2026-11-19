package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count int
	reset time.Time
}

type IPRateLimiter struct {
	mu          sync.Mutex
	entries     map[string]rateLimitEntry
	limit       int
	window      time.Duration
	lastCleanup time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		entries:     make(map[string]rateLimitEntry),
		limit:       limit,
		window:      window,
		lastCleanup: time.Now(),
	}
}

func (l *IPRateLimiter) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next(w, r)
			return
		}
		now := time.Now()
		key := clientIP(r)

		l.mu.Lock()
		if now.Sub(l.lastCleanup) >= l.window {
			for ip, entry := range l.entries {
				if !now.Before(entry.reset) {
					delete(l.entries, ip)
				}
			}
			l.lastCleanup = now
		}
		entry := l.entries[key]
		if entry.reset.IsZero() || !now.Before(entry.reset) {
			entry = rateLimitEntry{reset: now.Add(l.window)}
		}
		if entry.count >= l.limit {
			retryAfter := max(1, int(time.Until(entry.reset).Seconds()))
			l.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests"})
			return
		}
		entry.count++
		l.entries[key] = entry
		l.mu.Unlock()

		next(w, r)
	}
}

func clientIP(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Real-IP")); value != "" {
		return value
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
