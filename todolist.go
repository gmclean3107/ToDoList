package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func healthZ(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Cotnent-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	log.Info("Starting To-Do List API server")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", healthZ).Methods("GET")
	http.ListenAndServe(":8000", router)
}
