package services

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var (
	// NnSDK - NoName SDK
	NnSDK *nnsdk.NoNameSDK
)

// NnSDKUp
func NnSDKUp() {
	var err error
	NnSDK, err = nnsdk.NewNoNameSDK()
	if err != nil {
		ErrorHandler("Error connecting NoName WS", err)
	}

	// Set Host and API version
	NnSDK.SetWSHost(fmt.Sprintf("%s:%s", os.Getenv("WS_HOST"), os.Getenv("WS_PORT")))
	NnSDK.SetAPIVersion(os.Getenv("WS_API_VERSION"))
	NnSDK.SetAuth(
		nnsdk.NoNameSDKAuthentication{
			Username: os.Getenv("WS_USER"),
			Password: os.Getenv("WS_PASSWORD"),
		},
	)

	// Autenticate to WS
	_, err = NnSDK.Authenticate()
	if err != nil {
		log.Panicln("NNSDK - Authentication error", err)
	}

	// if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
	// 	NnSDK.Debug = true
	// }

	log.Println("************************************************")
	log.Println("NoName WS connection: OK!")
	log.Println("************************************************")
}
