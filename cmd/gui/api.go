package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

const baseURL = "http://localhost:8080"

type (
	showPage = internal.Paginated[internal.ShowMeta]
	songPage = internal.Paginated[internal.SongsTimesPlayed]
)

var (
	httpClient = &http.Client{Timeout: 10 * time.Second}

	errNotFound = errors.New("not found")

	validSetNames = []string{"set_1", "set_2", "set_3", "encore", "acoustic", "electric", "acoustic_1", "acoustic_2"}
)

// getJSON GETs u and decodes the JSON body into a T
func getJSON[T any](u string) (T, error) {
	var out T
	resp, err := httpClient.Get(u)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close() //nolint:errcheck

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return out, errNotFound
	case resp.StatusCode != http.StatusOK:
		return out, fmt.Errorf("api returned %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

// fetchPage loads any paginated URL, including the next/previous links
// the API hands back.
func fetchPage[T any](u string) (*T, error) {
	p, err := getJSON[T](u)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func searchSongs(song string) (*showPage, error) {
	if song == "" {
		return nil, errors.New("enter a song name")
	}
	return fetchPage[showPage](baseURL + "/shows?song=" + url.QueryEscape(song))
}

func searchVenues(venue string) (*showPage, error) {
	if venue == "" {
		return nil, errors.New("enter a venue name")
	}
	return fetchPage[showPage](baseURL + "/shows?venue=" + url.QueryEscape(venue))
}

func searchCities(city string) (*showPage, error) {
	if city == "" {
		return nil, errors.New("enter a city name")
	}
	return fetchPage[showPage](baseURL + "/shows?city=" + url.QueryEscape(city))
}

func searchLocations(location string) (*showPage, error) {
	if location == "" {
		return nil, errors.New("enter a location")
	}
	return fetchPage[showPage](baseURL + "/shows?location=" + url.QueryEscape(location))
}

func searchSetNames(setName string) (*showPage, error) {
	return fetchPage[showPage](baseURL + "/shows?set_name=" + url.QueryEscape(setName))
}

func searchNotes(option string) (*showPage, error) {
	return fetchPage[showPage](baseURL + "/shows?has_notes=" + url.QueryEscape(option))
}

func getMostPlayedSongs() (*songPage, error) {
	return fetchPage[songPage](baseURL + "/songs?sort=most_played")
}

func fetchShowByID(id string) (internal.ShowResponse, error) {
	show, err := getJSON[internal.ShowResponse](baseURL + "/shows/" + url.PathEscape(id))
	if errors.Is(err, errNotFound) {
		return show, fmt.Errorf("no show with ID %s", id)
	}
	return show, err
}

// fetchShowsByDate returns every show on a YYY-MM-DD date (some dates have 2 shows)
func fetchShowsByDate(date string) ([]internal.ShowResponse, error) {
	shows, err := getJSON[[]internal.ShowResponse](baseURL + "/shows/" + url.PathEscape(date))
	if errors.Is(err, errNotFound) || (err == nil && len(shows) == 0) {
		return nil, fmt.Errorf("no show found on %s", date)
	}
	if err != nil {
		return nil, err
	}
	return shows, nil
}
