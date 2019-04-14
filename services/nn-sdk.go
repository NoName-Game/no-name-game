package services

import (
	"log"
	"os"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
)

var (
	// NnSDK - NoName SDK
	NnSDK *nnsdk.NoNameSDK
)

// NnSDKUp
func NnSDKUp() {
	token := "YOUR-TOKEN"

	var err error
	NnSDK, err = nnsdk.NewNoNameSDK(token)
	if err != nil {
		ErrorHandler("Error connecting NoName WS", err)
	}

	// Set Host and API version
	NnSDK.SetWSHost("http://localhost:8080")
	NnSDK.SetAPIVersion("v1")

	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		// NnSDK.Debug = true
	}

	log.Println("************************************************")
	log.Println("NoName WS connection: OK!")
	log.Println("************************************************")
}
