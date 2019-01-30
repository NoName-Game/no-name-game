package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	port string
)

func init() {
	port = os.Getenv("SERVE_PORT")
	if port == "" {
		panic("$SERVE_PORT must be set")
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
