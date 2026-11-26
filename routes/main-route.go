package routes

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"meteo/helper"
	"meteo/services"
)

var codePostalValide = regexp.MustCompile(`^[0-9]{5}$`)

type accueil struct {
	CodePostal string
	Communes   []services.Commune
	Commune    *services.Commune
	Previsions []services.TemperatureHeure
	Message    string
}

func Home(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := accueil{CodePostal: strings.TrimSpace(query.Get("codePostal"))}
	status := http.StatusOK
	if query.Has("codePostal") {
		status = chargerMeteo(r, query.Get("commune"), &page)
	}
	helpers.Render(w, status, "index.html", page)
}

func chargerMeteo(r *http.Request, codeCommune string, page *accueil) int {
	if !codePostalValide.MatchString(page.CodePostal) {
		page.Message = "Saisis un code postal à cinq chiffres."
		return http.StatusBadRequest
	}

	communes, err := services.ChercherCommunes(r.Context(), page.CodePostal)
	if err != nil {
		log.Printf("API communes : %v", err)
		page.Message = "La recherche de communes est temporairement indisponible."
		return http.StatusBadGateway
	}
	page.Communes = communes
	if len(communes) == 0 {
		page.Message = "Aucune commune trouvée pour ce code postal."
		return http.StatusOK
	}

	if codeCommune == "" {
		if len(communes) > 1 {
			return http.StatusOK
		}
		codeCommune = communes[0].Code
	}
	for i := range communes {
		if communes[i].Code == codeCommune {
			page.Commune = &communes[i]
			break
		}
	}
	if page.Commune == nil {
		page.Message = "Choisis une commune dans la liste."
		return http.StatusBadRequest
	}

	coord := page.Commune.Centre.Coordinates
	if len(coord) != 2 {
		page.Message = "Les coordonnées de cette commune sont indisponibles."
		return http.StatusBadGateway
	}
	page.Previsions, err = services.Meteo(r.Context(), coord[1], coord[0])
	if err != nil {
		log.Printf("API météo : %v", err)
		page.Message = "Les prévisions météo sont temporairement indisponibles."
		return http.StatusBadGateway
	}
	return http.StatusOK
}
