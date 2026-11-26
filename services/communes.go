package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Commune struct {
	Code   string `json:"code"`
	Nom    string `json:"nom"`
	Centre struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"centre"`
}

func ChercherCommunes(ctx context.Context, codePostal string) ([]Commune, error) {
	params := url.Values{
		"codePostal": {codePostal},
		"fields":     {"code,nom,centre"},
	}
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
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("réponse API : %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}
