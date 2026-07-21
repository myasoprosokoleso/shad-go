package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type athleteResponse struct {
	orderedRecord
	MedalsByYear map[string]medals `json:"medals_by_year"`
}

type medals struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

func (srv *olympicsServer) handleAthleteInfo(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	name := params.Get("name")
	if name == "" {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}

	targetAthletes, medalsCounter := filterAndCount(srv.data, func(rec record) bool {
		return rec.Athlete == name
	})
	if len(targetAthletes) == 0 {
		http.Error(w, fmt.Sprintf("athlete '%s' not found", name), http.StatusNotFound)
		return
	}

	medalsByYear := make(map[string]medals)
	for _, a := range targetAthletes {
		if _, ok := medalsByYear[a.Year.String()]; ok {
			continue
		}

		_, medalsByYear[a.Year.String()] = filterAndCount(targetAthletes, func(rec record) bool {
			return rec.Year == a.Year
		})
	}

	resp := athleteResponse{
		orderedRecord: orderedRecord{
			Athlete: name,
			Country: targetAthletes[0].Country,
			Medals:  medalsCounter,
		},
		MedalsByYear: medalsByYear,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (srv *olympicsServer) handleTopAthletes(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	sport, k := params.Get("sport"), getLimit(params)
	if sport == "" || k < 0 {
		var param string
		if sport == "" {
			param = "sport"
		} else {
			param = "limit"
		}
		http.Error(w, fmt.Sprintf("invalid %s", param), http.StatusBadRequest)
		return
	}

	targetAthletes, _ := filterAndCount(srv.data, func(rec record) bool {
		return rec.Sport == sport
	})
	if len(targetAthletes) == 0 {
		http.Error(w, fmt.Sprintf("sport '%s' not found", sport), http.StatusNotFound)
		return
	}

	topAthletes := make(map[string]*orderedRecord)
	for _, a := range targetAthletes {
		if _, ok := topAthletes[a.Athlete]; ok {
			continue
		}

		_, medalsCounter := filterAndCount(targetAthletes, func(rec record) bool {
			return rec.Athlete == a.Athlete
		})
		topAthletes[a.Athlete] = &orderedRecord{
			Athlete: a.Athlete,
			Country: a.Country,
			Medals:  medalsCounter,
		}
	}

	topKAthletes := getTopK(getMapValues(topAthletes), k)

	resp := make([]athleteResponse, len(topKAthletes))
	for i, topAthlete := range topKAthletes {
		resp[i] = athleteResponse{orderedRecord: *topAthlete}
		medalsByYear := make(map[string]medals)
		for _, a := range targetAthletes {
			if a.Athlete != topAthlete.Athlete {
				continue
			}
			if _, ok := medalsByYear[a.Year.String()]; ok {
				continue
			}

			_, medalsByYear[a.Year.String()] = filterAndCount(targetAthletes, func(rec record) bool {
				return rec.Athlete == a.Athlete && rec.Year == a.Year
			})

		}
		resp[i].MedalsByYear = medalsByYear
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (srv *olympicsServer) handleTopCountries(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	year, k := params.Get("year"), getLimit(params)
	if year == "" || k < 0 {
		var param string
		if year == "" {
			param = "year"
		} else {
			param = "limit"
		}
		http.Error(w, fmt.Sprintf("invalid %s", param), http.StatusBadRequest)
		return
	}

	targetCountries, _ := filterAndCount(srv.data, func(rec record) bool {
		return rec.Year.String() == year
	})
	if len(targetCountries) == 0 {
		http.Error(w, fmt.Sprintf("year '%s' not found", year), http.StatusNotFound)
		return
	}

	topCountries := make(map[string]*orderedRecord)
	for _, c := range targetCountries {
		if _, ok := topCountries[c.Country]; ok {
			continue
		}

		_, medalsCounter := filterAndCount(targetCountries, func(rec record) bool {
			return rec.Country == c.Country
		})
		topCountries[c.Country] = &orderedRecord{
			Country: c.Country,
			Medals:  medalsCounter,
		}
	}

	topKCountries := getTopK(getMapValues(topCountries), k)

	type countryResponse struct {
		Country string `json:"country"`
		medals
	}
	resp := make([]countryResponse, len(topKCountries))
	for i, c := range topKCountries {
		resp[i] = countryResponse{Country: c.Country, medals: c.Medals}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func filterAndCount(records []record, filter func(record) bool) ([]record, medals) {
	var (
		targetRecords []record
		medalsCounter medals
	)
	for _, rec := range records {
		if filter(rec) {
			targetRecords = append(targetRecords, rec)
			medalsCounter.Gold += rec.Gold
			medalsCounter.Silver += rec.Silver
			medalsCounter.Bronze += rec.Bronze
			medalsCounter.Total += rec.Total
		}
	}
	return targetRecords, medalsCounter
}

func getLimit(params url.Values) int {
	limit := params.Get("limit")
	if limit == "" {
		return 3
	}

	if k, err := strconv.Atoi(limit); err == nil && k >= 0 {
		return k
	}
	return -1
}
