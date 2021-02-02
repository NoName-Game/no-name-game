package antiflood

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Antiflood
type Antiflood struct {
	TTL int
}

// Antiflood - Servizio per limitare l'invio di messaggi tramite api telegram
func (antiflood *Antiflood) Init() {
	envAntifloodTTL, _ := strconv.Atoi(os.Getenv("ANTIFLOOD_TTL"))
	if envAntifloodTTL == 0 {
		envAntifloodTTL = 5
	}

	antiflood.TTL = envAntifloodTTL
	logrus.WithField("ttl", envAntifloodTTL).Info("[*] Telegram Antiflood: OK!")
	return
}
