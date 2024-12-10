package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type Info struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

func getWeather(location Location, weather *[]Info, wg *sync.WaitGroup) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=756791a68a4fefa625d09a8ee8ed99c3", location.Latitude, location.Longitude)

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
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
		} `json:"weather"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	for _, hit := range data.Weather {
		*weather = append(*weather, Info{
			Main:        hit.Main,
			Description: hit.Description,
		})
	}
	defer wg.Done()
}
