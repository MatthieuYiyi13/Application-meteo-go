package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Commune struct {
	Code        string `json:"code"`
	Nom         string `json:"nom"`
	Departement string `json:"codeDepartement"`
	Centre      struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"centre"`
}

func ChercherCommunes(ctx context.Context, codePostal string) ([]Commune, error) {
	return chercherCommunes(ctx, url.Values{"codePostal": {codePostal}})
}

func ChercherCommunesParNom(ctx context.Context, nom string) ([]Commune, error) {
	return chercherCommunes(ctx, url.Values{"nom": {nom}, "boost": {"population"}})
}

func chercherCommunes(ctx context.Context, params url.Values) ([]Commune, error) {
	params.Set("fields", "code,nom,codeDepartement,centre")
	var communes []Commune
	err := getJSON(ctx, "https://geo.api.gouv.fr/communes?"+params.Encode(), &communes)
	return communes, err
}

var client = &http.Client{Timeout: 8 * time.Second}

func getJSON(ctx context.Context, endpoint string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	const maxTentatives = 2
	for tentative := 1; ; tentative++ {
		resp, err := client.Do(req)
		if err != nil {
			var errReseau net.Error
			if tentative < maxTentatives && ctx.Err() == nil && errors.As(err, &errReseau) && errReseau.Timeout() {
				continue
			}
			return err
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			return json.NewDecoder(resp.Body).Decode(result)
		}
		resp.Body.Close()
		if tentative == maxTentatives || ctx.Err() != nil ||
			(resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusGatewayTimeout) {
			return fmt.Errorf("réponse API : %s", resp.Status)
		}
	}
}
