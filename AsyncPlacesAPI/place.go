package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Place struct {
	Xid         string `json:"xid"`
	Name        string `json:"name"`
	Description string
}

func getPlaces(location Location, prop *[]Place, wg *sync.WaitGroup) {
	url := fmt.Sprintf("https://api.opentripmap.com/0.1/en/places/radius?radius=1000&lon=%f&lat=%f&apikey=5ae2e3f221c38a28845f05b616852a75d7d1101b4c7e9f71737b2594", location.Longitude, location.Latitude)
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("ошибка создания запроса: %w", err))
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var data struct {
		Features []struct {
			Properties struct {
				Xid  string `json:"xid"`
				Name string `json:"name"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	for _, hit := range data.Features {
		*prop = append(*prop, Place{
			Xid:         hit.Properties.Xid,
			Name:        hit.Properties.Name,
			Description: getDescription(hit.Properties.Xid),
		})
	}
	defer wg.Done()
}

func getDescription(xid string) string {
	url := fmt.Sprintf("https://api.opentripmap.com/0.1/en/places/xid/%s?apikey=5ae2e3f221c38a28845f05b616852a75d7d1101b4c7e9f71737b2594", xid)
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("ошибка создания запроса: %w", err))
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var data struct {
		Wiki struct {
			Description string `json:"text"`
		} `json:"wikipedia_extracts"`

		Info struct {
			Description string `json:"descr"`
		} `json:"info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	if data.Info.Description != "" {
		//fmt.Println(data.Info.Description)
		return data.Info.Description
	} else {
		//fmt.Println(data.Wiki.Description)
		return data.Wiki.Description
	}

}
