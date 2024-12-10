package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	var weather []Info
	var place []Place
	var loc string

	locCh := make(chan []Location)
	fmt.Println("Введите локацию")
	_, err := fmt.Scanln(&loc)

	if err != nil {
		panic(err)
	}

	go geoPosition(loc, locCh, &wg)
	locations := <-locCh

	fmt.Println(locations)
	fmt.Println("\nВведите номер предложенных локаций начиная с 0")

	var num int
	_, err = fmt.Scanln(&num)
	if err != nil {
		panic(err)
	}

	go getWeather(locations[num], &weather, &wg)
	go getPlaces(locations[num], &place, &wg)

	wg.Wait()
	fmt.Println(locations[num])
	fmt.Println(weather)
	for pl := range place {
		fmt.Println(place[pl])
	}

}
