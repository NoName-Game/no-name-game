package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/no-name-game/no-name/app/commands/godeffect"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
)

var (
	port string
)

func init() {
	port = os.Getenv("SERVE_PORT")
	if port == "" {
		services.ErrorHandler("$SERVE_PORT must be set", errors.New("$SERVE_PORT not setted"))
	}

	log.Println("Http service run on:" + port)
}

// Run - listen and serve
func Run() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Pong")
	})

	http.HandleFunc("/omg", func(w http.ResponseWriter, r *http.Request) {
		godeffect.OMG()
		fmt.Fprintf(w, "Created")
	})

	http.HandleFunc("/galaxy", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)

		results := models.Stars{}
		services.Database.Find(&results)

		response, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		services.ErrorHandler("Error start ListenAndServe", err)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
