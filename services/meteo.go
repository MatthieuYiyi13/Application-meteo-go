package services

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PrevisionJour struct {
	Date      string
	Min       string
	Max       string
	Icone     string
	Condition string
}

func Meteo(ctx context.Context, latitude, longitude float64) ([]PrevisionJour, error) {
	const nombreJours = 7
	params := url.Values{
		"latitude":      {strconv.FormatFloat(latitude, 'f', -1, 64)},
		"longitude":     {strconv.FormatFloat(longitude, 'f', -1, 64)},
		"daily":         {"weather_code,temperature_2m_min,temperature_2m_max"},
		"timezone":      {"auto"},
		"forecast_days": {strconv.Itoa(nombreJours)},
	}
	var data struct {
		Daily struct {
			Time []string   `json:"time"`
			Min  []*float64 `json:"temperature_2m_min"`
			Max  []*float64 `json:"temperature_2m_max"`
			Code []*int     `json:"weather_code"`
		} `json:"daily"`
	}
	if err := getJSON(ctx, "https://api.open-meteo.com/v1/forecast?"+params.Encode(), &data); err != nil {
		return nil, err
	}
	if len(data.Daily.Time) != nombreJours || len(data.Daily.Min) != nombreJours || len(data.Daily.Max) != nombreJours {
		return nil, errors.New("prévisions quotidiennes incomplètes")
	}

	previsions := make([]PrevisionJour, 0, nombreJours)
	for i, date := range data.Daily.Time {
		jour, err := time.Parse(time.DateOnly, date)
		if err != nil {
			return nil, err
		}
		icone, condition := conditionMeteo(nil)
		if i < len(data.Daily.Code) {
			icone, condition = conditionMeteo(data.Daily.Code[i])
		}
		previsions = append(previsions, PrevisionJour{
			Date:      formaterDate(jour),
			Min:       formaterTemperature(data.Daily.Min[i]),
			Max:       formaterTemperature(data.Daily.Max[i]),
			Icone:     icone,
			Condition: condition,
		})
	}
	return previsions, nil
}

func conditionMeteo(code *int) (icone, condition string) {
	if code != nil {
		switch *code {
		case 0:
			return "☀️", "Ensoleillé"
		case 1, 2:
			return "⛅", "Éclaircies"
		case 3:
			return "☁️", "Couvert"
		case 45, 48:
			return "🌫️", "Brouillard"
		case 51, 53, 55, 56, 57:
			return "🌦️", "Bruine"
		case 61, 63, 65, 66, 67, 80, 81, 82:
			return "🌧️", "Pluie"
		case 71, 73, 75, 77, 85, 86:
			return "🌨️", "Neige"
		case 95, 96, 99:
			return "⛈️", "Orage"
		}
	}
	return "—", "Indisponible"
}

func formaterDate(date time.Time) string {
	jours := [...]string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"}
	mois := [...]string{"janv", "févr", "mars", "avr", "mai", "juin", "juil", "août", "sept", "oct", "nov", "déc"}
	return jours[date.Weekday()] + " " + strconv.Itoa(date.Day()) + " " + mois[date.Month()-1]
}

func formaterTemperature(valeur *float64) string {
	if valeur == nil {
		return "Indisponible"
	}
	temperature := strconv.FormatFloat(*valeur, 'f', 1, 64)
	return strings.ReplaceAll(temperature, ".", ",") + "°C"
}
