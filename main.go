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
	mux.HandleFunc("GET /previsions", routes.Home)
	mux.HandleFunc("GET /carte", routes.Carte)
	mux.HandleFunc("GET /assets/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/style.css")
	})
	log.Println("Application météo sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
