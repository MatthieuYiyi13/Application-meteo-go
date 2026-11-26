package services

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type TemperatureHeure struct {
	Heure       string
	Temperature string
}

func Meteo(ctx context.Context, latitude, longitude float64) ([]TemperatureHeure, error) {
	params := url.Values{
		"latitude":      {strconv.FormatFloat(latitude, 'f', -1, 64)},
		"longitude":     {strconv.FormatFloat(longitude, 'f', -1, 64)},
		"hourly":        {"temperature_2m"},
		"timezone":      {"auto"},
		"forecast_days": {"1"},
	}
	var data struct {
		Hourly struct {
			Time        []string   `json:"time"`
			Temperature []*float64 `json:"temperature_2m"`
		} `json:"hourly"`
	}
	if err := getJSON(ctx, "https://api.open-meteo.com/v1/forecast?"+params.Encode(), &data); err != nil {
		return nil, err
	}
	if len(data.Hourly.Time) == 0 || len(data.Hourly.Time) != len(data.Hourly.Temperature) {
		return nil, errors.New("prévisions horaires incomplètes")
	}
	previsions := make([]TemperatureHeure, 0, len(data.Hourly.Time))
	for i, date := range data.Hourly.Time {
		_, heure, ok := strings.Cut(date, "T")
		if !ok {
			return nil, errors.New("heure de prévision invalide")
		}
		temperature := "Indisponible"
		if valeur := data.Hourly.Temperature[i]; valeur != nil {
			temperature = strconv.FormatFloat(*valeur, 'f', 1, 64)
		}
		previsions = append(previsions, TemperatureHeure{Heure: heure, Temperature: temperature})
	}
	return previsions, nil
}
