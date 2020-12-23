package limiter

import (
	"os"
	"strconv"

	"go.uber.org/ratelimit"

	"github.com/sirupsen/logrus"
)

// RateLimiter
type RateLimiter struct {
	Limiter ratelimit.Limiter
}

// RateLimiter - Servizio per limitare l'invio di messaggi tramite api telegram
func (rateLimiter *RateLimiter) Init() {
	envRateLimit, _ := strconv.Atoi(os.Getenv("TELEGRAM_RATELIMIT"))
	if envRateLimit == 0 {
		envRateLimit = 30
	}

	rateLimiter.Limiter = ratelimit.New(envRateLimit)
	logrus.WithField("limit", envRateLimit).Info("[*] Telegram Rate Limiter: OK!")
	return
}
