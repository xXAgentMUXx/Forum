package security

import (
	
	"net/http"
	"sync"
	"time"
)


type RateLimiter struct {
	visits map[string][]time.Time
	mu     sync.Mutex
	limit  int           
	window time.Duration 
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visits: make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
	go func() {
		for {
			time.Sleep(window)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr 
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		validTimes := []time.Time{}

		for _, t := range rl.visits[ip] {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		rl.visits[ip] = validTimes
		if len(validTimes) >= rl.limit {
			http.Error(w, "Trop de requêtes. Attendez un moment.", http.StatusTooManyRequests)
			return
		}
		rl.visits[ip] = append(rl.visits[ip], now)
		next.ServeHTTP(w, r)
	})
}
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, timestamps := range rl.visits {
		var validTimes []time.Time
		for _, t := range timestamps {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		if len(validTimes) == 0 {
			delete(rl.visits, ip) 
		} else {
			rl.visits[ip] = validTimes
		}
	}
}