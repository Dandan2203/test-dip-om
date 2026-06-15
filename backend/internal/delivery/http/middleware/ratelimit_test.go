package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newLimiter(rate, burst float64) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}
}

func TestRateLimiter_AllowsUpToBurst(t *testing.T) {
	rl := newLimiter(1, 5)
	// Перші 5 запитів (burst) проходять.
	for i := 0; i < 5; i++ {
		assert.True(t, rl.allow("1.1.1.1"), "запит %d у межах сплеску має пройти", i+1)
	}
	// Шостий — вичерпано, токенів < 1.
	assert.False(t, rl.allow("1.1.1.1"), "запит понад сплеск має бути відхилений")
}

func TestRateLimiter_IndependentPerIP(t *testing.T) {
	rl := newLimiter(1, 2)
	assert.True(t, rl.allow("1.1.1.1"))
	assert.True(t, rl.allow("1.1.1.1"))
	assert.False(t, rl.allow("1.1.1.1"), "перша IP вичерпала ліміт")

	// Інша IP має власний кошик.
	assert.True(t, rl.allow("2.2.2.2"))
	assert.True(t, rl.allow("2.2.2.2"))
	assert.False(t, rl.allow("2.2.2.2"))
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := newLimiter(20, 1) // 20 токенів/с, сплеск 1
	assert.True(t, rl.allow("3.3.3.3"))
	assert.False(t, rl.allow("3.3.3.3"), "одразу після першого — порожньо")

	// За 100 мс набіжить 20*0.1 = 2 токени (обмежиться сплеском 1) — знову можна.
	time.Sleep(120 * time.Millisecond)
	assert.True(t, rl.allow("3.3.3.3"), "після паузи токен поповнився")
}

func TestRateLimit_Middleware429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(1, 3))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	do := func() int {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	// Сплеск 3 → три 200, четвертий 429.
	assert.Equal(t, http.StatusOK, do())
	assert.Equal(t, http.StatusOK, do())
	assert.Equal(t, http.StatusOK, do())
	assert.Equal(t, http.StatusTooManyRequests, do())
}
