package auth

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	maxAttempts = 5
	window      = 1 * time.Minute
)

type attempt struct {
	count     int
	windowEnd time.Time
}

// RateLimiter limita intentos de login por IP con una ventana fija -- suficiente para
// un solo usuario, sin necesidad de un store externo (Redis, etc).
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attempt
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{attempts: make(map[string]*attempt)}
	go rl.janitor()
	return rl
}

func (rl *RateLimiter) janitor() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for ip, a := range rl.attempts {
			if now.After(a.windowEnd) {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow devuelve false si la IP ya agoto su presupuesto de intentos en la ventana actual.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	a, ok := rl.attempts[ip]
	if !ok || now.After(a.windowEnd) {
		rl.attempts[ip] = &attempt{count: 1, windowEnd: now.Add(window)}
		return true
	}
	if a.count >= maxAttempts {
		return false
	}
	a.count++
	return true
}

func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
