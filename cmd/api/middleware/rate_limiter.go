package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor tracks the rate limiter for a single client and when it was last seen
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter tracks a separate token-bucket rate limiter per client IP
type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter creates a limiter allowing `r` requests/sec per IP (sustained), with bursts up to `b` requests.
// It also starts a background goroutine that evicts IPs which haven't been seen in a while, so the visitors
// map doesn't grow forever.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	r1 := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}
	go r1.cleanupStaleVisitors()
	return r1
}

// getLimiter returns the rate limiter for the given IP, creating one on first sight.
func (r1 *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	r1.mu.Lock()
	defer r1.mu.Unlock()

	v, exists := r1.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(r1.rate, r1.burst)
		r1.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupStaleVisitors runs forever, periodically removing IPs that haven't made a request recently.
func (r1 *IPRateLimiter) cleanupStaleVisitors() {
	for {
		time.Sleep(time.Minute)

		r1.mu.Lock()
		for ip, v := range r1.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(r1.visitors, ip)
			}
		}
		r1.mu.Unlock()
	}
}

// Middleware wraps an http.Handler and returns 429 once a client IP exceeds its rate limit.
func (r1 *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := r1.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded, slow down", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if cfIP := r.Header.Get("Cf-Connecting-Ip"); cfIP != "" {
		return cfIP
	}

	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := indexComma(fwd); i != -1 {
			return fwd[:i]
		}
		return fwd
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func indexComma(s string) int {
	for i, c := range s {
		if c == ',' {
			return i
		}
	}
	return -1
}
