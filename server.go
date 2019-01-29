package main

import (
	"fmt"
	"net/http"
	"os"
)

func server() {
	port := os.Getenv("SERVE_PORT")

	if port == "" {
		panic("$SERVE_PORT must be set")
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Pong")
	})

	http.ListenAndServe(":"+port, nil)
}
