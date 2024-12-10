package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func geoPosition(location string, locCh chan []Location, wg *sync.WaitGroup) {
	url := fmt.Sprintf("https://graphhopper.com/api/1/geocode?q=%s&key=38a390bd-3f03-4d7a-869e-13b49c9b80e8", location)
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("ошибка создания запроса: %w", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(fmt.Errorf("ошибка создания запроса: %w", err))
		}
	}(resp.Body)

	var data struct {
		Hits []struct {
			Point struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"point"`
			Name string `json:"name"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	var locations []Location
	for _, hit := range data.Hits {
		locations = append(locations, Location{
			Name:      hit.Name,
			Latitude:  hit.Point.Lat,
			Longitude: hit.Point.Lng,
		})
	}

	locCh <- locations
	defer wg.Done()
}
