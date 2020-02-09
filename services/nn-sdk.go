package services

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var (
	// NnSDK - NoName SDK
	NnSDK *nnsdk.NoNameSDK
)

// NnSDKUp - Metodo di verfica e connessione al WS principale di NoName
func NnSDKUp() (err error) {
	// Up del servizio per la comunicazione
	NnSDK, err = nnsdk.NewNoNameSDK()
	if err != nil {
		return err
	}

	// Imposto Host e versionamento del sistema API
	NnSDK.SetWS(
		nnsdk.WebService{
			Host:       fmt.Sprintf("%s:%s", os.Getenv("WS_HOST"), os.Getenv("WS_PORT")),
			APIVersion: os.Getenv("WS_API_VERSION"),
		},
	)

	// Imposto username e password per l'autenticazione
	NnSDK.SetAuth(
		nnsdk.Authentication{
			Username: os.Getenv("WS_USER"),
			Password: os.Getenv("WS_PASSWORD"),
		},
	)

	// Avvio autenticazione
	_, err = NnSDK.Authenticate()
	if err != nil {
		return err
	}

	// Riporto a video stato servizio
	log.Println("************************************************")
	log.Println("NoName WS connection: OK!")
	log.Println("************************************************")

	return
}
