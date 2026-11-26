package routes

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"meteo/helper"
	"meteo/services"
)

var codePostalValide = regexp.MustCompile(`^[0-9]{5}$`)
var rechercheNumerique = regexp.MustCompile(`^[0-9]+$`)

type accueil struct {
	Recherche     string
	ParCodePostal bool
	Communes      []services.Commune
	Commune       *services.Commune
	Previsions    []services.PrevisionJour
	Message       string
	CarteURL      string
}

func Home(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	aRecherche := query.Has("q") || query.Has("codePostal")
	page := accueil{Recherche: strings.TrimSpace(query.Get("q"))}
	if !query.Has("q") && query.Has("codePostal") {
		page.Recherche = strings.TrimSpace(query.Get("codePostal"))
		page.ParCodePostal = true
	}
	page.ParCodePostal = page.ParCodePostal || rechercheNumerique.MatchString(page.Recherche)
	status := http.StatusOK
	if aRecherche {
		status = chargerMeteo(r, query.Get("commune"), &page)
	}
	if status != http.StatusOK {
		helpers.RenderError(w, status, page.Message)
		return
	}
	if len(page.Previsions) == 0 {
		helpers.Render(w, http.StatusOK, "index.html", page)
		return
	}
	helpers.Render(w, status, "previsions.html", page)
}

func chargerMeteo(r *http.Request, codeCommune string, page *accueil) int {
	if page.ParCodePostal && !codePostalValide.MatchString(page.Recherche) {
		page.Message = "Saisis un code postal à cinq chiffres."
		return http.StatusBadRequest
	}
	if !page.ParCodePostal && (page.Recherche == "" || len([]rune(page.Recherche)) > 100) {
		page.Message = "Saisis un nom de commune de 1 à 100 caractères."
		return http.StatusBadRequest
	}

	var communes []services.Commune
	var err error
	if page.ParCodePostal {
		communes, err = services.ChercherCommunes(r.Context(), page.Recherche)
	} else {
		communes, err = services.ChercherCommunesParNom(r.Context(), page.Recherche)
	}
	if err != nil {
		log.Printf("API communes : %v", err)
		page.Message = "La recherche de communes est temporairement indisponible."
		return http.StatusBadGateway
	}
	page.Communes = communes
	if len(communes) == 0 {
		page.Message = "Aucune commune trouvée pour cette recherche."
		return http.StatusNotFound
	}

	if codeCommune == "" {
		switch {
		case len(communes) == 1:
			codeCommune = communes[0].Code
		case !page.ParCodePostal:
			codeCommune = codeNomExact(communes, page.Recherche)
		}
		if codeCommune == "" {
			return http.StatusOK
		}
	}
	for i := range communes {
		if communes[i].Code == codeCommune {
			page.Commune = &communes[i]
			break
		}
	}
	if page.Commune == nil {
		page.Message = "Cette commune est introuvable dans les résultats."
		return http.StatusNotFound
	}

	coord := page.Commune.Centre.Coordinates
	if len(coord) != 2 {
		page.Message = "Les coordonnées de cette commune sont indisponibles."
		return http.StatusBadGateway
	}
	longitude, latitude := coord[0], coord[1]
	page.Previsions, err = services.Meteo(r.Context(), latitude, longitude)
	if err != nil {
		log.Printf("API météo : %v", err)
		page.Message = "Les prévisions météo sont temporairement indisponibles."
		return http.StatusBadGateway
	}
	params := url.Values{
		"lat": {fmt.Sprintf("%.6f", latitude)},
		"lon": {fmt.Sprintf("%.6f", longitude)},
	}
	page.CarteURL = "/carte?" + params.Encode()

	return http.StatusOK
}

func Carte(w http.ResponseWriter, r *http.Request) {
	helpers.Render(w, http.StatusOK, "carte.html", nil)
}

func codeNomExact(communes []services.Commune, nom string) string {
	var code string
	for _, commune := range communes {
		if strings.EqualFold(commune.Nom, nom) {
			if code != "" {
				return ""
			}
			code = commune.Code
		}
	}
	return code
}
