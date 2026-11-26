package main

import (
	"log"
	"net/http"

	"meteo/helper"
	"meteo/routes"
)

func main() {
	helpers.LoadTemplates()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", routes.Home)
	log.Println("Application météo sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
