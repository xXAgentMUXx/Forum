package security

import (
	
	"net/http"
	"sync"
	"time"
)

// RateLimiter struct
type RateLimiter struct {
	visits map[string][]time.Time
	mu     sync.Mutex
	limit  int           
	window time.Duration 
}

// LoginLimiter struct
type LoginLimiter struct {
	attempts map[string]int       
	lockout  map[string]time.Time 
	mu       sync.Mutex
}

// Function creates a new RateLimiter with the specified request limit
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visits: make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
	// Start a background goroutine
	go func() {
		for {
			time.Sleep(window)
			rl.cleanup() // Clean up old timestamps
		}
	}()
	return rl
}
// Limit is a middleware for the rate limit on incoming requests
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr 
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now() // Get the current time
		validTimes := []time.Time{}

		// Remove visit timestamps that are outside the rate limit window
		for _, t := range rl.visits[ip] {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		// Update the visits map with the valid timestamps
		rl.visits[ip] = validTimes

		// If the number of valid visits exceeds the limit, return a message
		if len(validTimes) >= rl.limit {
			http.Error(w, "Trop de requêtes. Attendez un moment.", http.StatusTooManyRequests)
			return
		}
		// Otherwise, add the current timestamp
		rl.visits[ip] = append(rl.visits[ip], now)
		next.ServeHTTP(w, r)
	})
}

// Function to cleanup removes timestamps that are outside the time window
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Iterate through each IP's visit timestamps
	for ip, timestamps := range rl.visits {
		var validTimes []time.Time
		// Keep only the timestamps within the time window
		for _, t := range timestamps {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}
		// If no valid timestamps remain for the IP, delete it
		if len(validTimes) == 0 {
			delete(rl.visits, ip) 
		} else {
			// Otherwise, update the visits with the valid timestamps
			rl.visits[ip] = validTimes
		}
	}
}

// Function to creates a new LoginLimiter to manage failed login attempts
func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{
		attempts: make(map[string]int), // Initialize map to track failed attempts
		lockout:  make(map[string]time.Time), // Initialize map to track lockout times
	}
}

// Check if checks if an IP is currently locked out and returns the remaining lockout time
func (ll *LoginLimiter) CheckLock(ip string) (bool, time.Duration) {
	ll.mu.Lock()  // Lock the map for safe access
	defer ll.mu.Unlock()

	if lockUntil, found := ll.lockout[ip]; found {
		// If the IP is locked, check if the lockout time has passed
		if time.Now().Before(lockUntil) {
			return true, time.Until(lockUntil) // Return lockout status and remaining time
		}
		// If lockout period has passed, remove the IP from lockout
		delete(ll.lockout, ip) 
	}
	return false, 0
}

// Function to increments the failed attempt count for an IP and applies lockout
func (ll *LoginLimiter) FailedAttempt(ip string) time.Duration {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	ll.attempts[ip]++ // Increment the failed attempt count for the IP

	// If the number of failed attempts exceeds 5, initiate lockout with increasing timeout
	if ll.attempts[ip] > 5 { 
		timeout := time.Duration(30*1<<int(ll.attempts[ip]-5)) * time.Second
		ll.lockout[ip] = time.Now().Add(timeout)
		return timeout
	}
	// No lockout applied, return 0
	return 0
}

// Function to resets the failed attempt count and lockout time
func (ll *LoginLimiter) Reset(ip string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	delete(ll.attempts, ip) // Remove the failed attempts for the IP
	delete(ll.lockout, ip) // Remove the lockout time for the IP
}