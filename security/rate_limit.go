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
type LoginLimiter struct {
	attempts map[string]int       
	lockout  map[string]time.Time 
	mu       sync.Mutex
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

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{
		attempts: make(map[string]int),
		lockout:  make(map[string]time.Time),
	}
}

func (ll *LoginLimiter) CheckLock(ip string) (bool, time.Duration) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if lockUntil, found := ll.lockout[ip]; found {
		if time.Now().Before(lockUntil) {
			return true, time.Until(lockUntil)
		}
		delete(ll.lockout, ip) 
	}
	return false, 0
}
func (ll *LoginLimiter) FailedAttempt(ip string) time.Duration {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	ll.attempts[ip]++

	if ll.attempts[ip] > 5 { 
		timeout := time.Duration(30*1<<int(ll.attempts[ip]-5)) * time.Second
		ll.lockout[ip] = time.Now().Add(timeout)
		return timeout
	}

	return 0
}
func (ll *LoginLimiter) Reset(ip string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	delete(ll.attempts, ip)
	delete(ll.lockout, ip)
}