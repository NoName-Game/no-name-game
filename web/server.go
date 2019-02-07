package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/no-name-game/no-name/config"
)

var (
	port string
)

func init() {
	port = os.Getenv("SERVE_PORT")
	if port == "" {
		config.ErrorHandler("$SERVE_PORT must be set", errors.New("$SERVE_PORT not setted"))
	}
}

// Run - listen and serve
func Run() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Pong")
	})

	http.ListenAndServe(":"+port, nil)

	log.Println("Http service run on:" + port)
}
