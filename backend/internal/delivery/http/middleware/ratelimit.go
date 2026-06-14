package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor — стан token-bucket для однієї IP-адреси.
type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// rateLimiter — обмежувач частоти запитів за IP (token bucket).
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // токенів за секунду
	burst    float64 // максимальний запас токенів
}

// RateLimit — middleware обмеження частоти: rate запитів/с зі сплеском burst на IP.
func RateLimit(rate float64, burst float64) gin.HandlerFunc {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}
	go rl.cleanup()

	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "забагато запитів, спробуйте трохи згодом"},
			})
			return
		}
		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	// Поповнюємо токени пропорційно часу, що минув.
	v.tokens += now.Sub(v.lastSeen).Seconds() * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// cleanup — періодично прибирає неактивні IP, щоб мапа не росла безмежно.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
