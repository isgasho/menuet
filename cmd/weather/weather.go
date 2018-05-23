package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/caseymrm/menuet"
)

func temperature(woeid string) (temp, unit, text string) {
	url := "https://query.yahooapis.com/v1/public/yql?format=json&q=select%20item.condition%20from%20weather.forecast%20where%20woeid%20%3D%20" + woeid
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	var response struct {
		Query struct {
			Results struct {
				Channel struct {
					Item struct {
						Condition struct {
							Temp string `json:"temp"`
							Text string `json:"text"`
						} `json:"condition"`
					} `json:"item"`
					Units struct {
						Temperature string `json:"temperature"`
					} `json:"units"`
				} `json:"channel"`
			} `json:"results"`
		} `json:"query"`
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		log.Fatal(err)
	}
	return response.Query.Results.Channel.Item.Condition.Temp, response.Query.Results.Channel.Units.Temperature, response.Query.Results.Channel.Item.Condition.Text
}

var woeids = map[int]string{
	2442047: "Los Angeles",
	2487956: "San Francisco",
	2459115: "New York",
}

func setWeather() {
	temp, unit, text := temperature(menuet.Defaults().String("loc"))
	menuet.App().SetMenuState(&menuet.MenuState{
		Title: fmt.Sprintf("%s°%s and %s", temp, unit, text),
		Items: menuItems(),
	})
}

func menuItems() []menuet.MenuItem {
	items := []menuet.MenuItem{}
	for woeid, name := range woeids {
		items = append(items, menuet.MenuItem{
			Text:     name,
			Callback: strconv.Itoa(woeid),
			State:    strconv.Itoa(woeid) == menuet.Defaults().String("loc"),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Callback < items[j].Callback
	})
	return items
}

func hourlyWeather() {
	for {
		setWeather()
		time.Sleep(time.Hour)
	}
}

func handleClicks(callback chan string) {
	for woeid := range callback {
		menuet.Defaults().SetString("loc", woeid)
		setWeather()
		num, err := strconv.Atoi(woeid)
		if err != nil {
			log.Printf("Atoi: %v", err)
		}
		menuet.App().Notification(menuet.Notification{
			Title:    "Location changed",
			Subtitle: "Did you move?",
			Message:  "Now showing weather for " + woeids[num],
		})
	}
}

func main() {
	// Load the location from last time
	woeid := menuet.Defaults().String("loc")
	if woeid == "" {
		menuet.Defaults().SetString("loc", "2442047")
	}

	// Start the hourly check, and set the first value
	go hourlyWeather()

	// Configure the application
	menuet.App().Label = "com.github.caseymrm.menuet.weather"

	// Hook up the on-click to populate the menu
	menuet.App().MenuOpened = menuItems

	// Set up the click channel
	clickChannel := make(chan string)
	menuet.App().Clicked = clickChannel
	go handleClicks(clickChannel)

	// Run the app (does not return)
	menuet.App().RunApplication()
}
