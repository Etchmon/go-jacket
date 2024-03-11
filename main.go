package main

import (
    "fmt"
    "io"
    "net/http"
	"encoding/json"
	"time"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
    "os"
)

type Weather struct {
	Location struct {
		Name string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempF float64 `json:"temp_f"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Hour []struct {
				TimeEpoch int64 `json:"time_epoch"`
				TempF float64 `json:"temp_f"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain float64 `json:"chance_of_rain"`
				AirQuality float64 `json:"air_quality"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func main() {
	fmt.Println("Welcome!")
	apiKey := os.Getenv("WEATHER_API_KEY")
    if apiKey == "" {
        fmt.Println("Missing WEATHER_API_KEY environment variable")
        os.Exit(1)
    }
	res,err := http.Get("http://api.weatherapi.com/v1/forecast.json?key=" + apiKey + "&q=auto:ip&days=1&aqi=no&alerts=no")
	if err != nil {
		panic(err)
	}
	// close body of response
	defer res.Body.Close()

	// check if the response is 200 OK
	if res.StatusCode != 200 {
		fmt.Println("Status code:", res.StatusCode)
		panic("Weather API not available")
	}

	// read the body of the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var weather Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		panic(err)
	}
	
	location, current, hours := weather.Location, weather.Current, weather.Forecast.Forecastday[0].Hour

	// Do you need a jacket?
	var jacket string
	needJacket := false

	for _, hour := range hours {
    	date := time.Unix(hour.TimeEpoch, 0)
    	hourOfDay := date.Hour()

    	if (hourOfDay >= 8 && hourOfDay <= 22 && hour.TempF < 60) || hour.ChanceOfRain > 40 {
        	needJacket = true
        	break
    	}
	}

	if needJacket {
    	jacket = color.RedString("You might need a jacket")
	} else {
    	jacket = "You don't need a jacket"
	}

	fmt.Println(jacket)

	fmt.Printf(
	"%s, %s: %.0fF, %s\n", 
	location.Name, 
	location.Country, 
	current.TempF, 
	current.Condition.Text,
    )

	
	// Create a new table
    t := table.NewWriter()
    t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)

    // Add the headers
    t.AppendHeader(table.Row{"Time", "Temperature", "Chance of Rain", "Condition"})

	// Add the rows
    for _, hour := range hours {
        date := time.Unix(hour.TimeEpoch, 0)

        if date.Before(time.Now()) {
            continue
        }

		// Determine the chance of rain string
    	var chanceOfRainStr string
    	if hour.ChanceOfRain < 40 {
        	chanceOfRainStr = fmt.Sprintf("%.0f%%", hour.ChanceOfRain)
    	} else {
        	chanceOfRainStr = color.RedString("%.0f%%", hour.ChanceOfRain)
    	}

        t.AppendRow(table.Row{
            date.Format("15:04"),
            fmt.Sprintf("%.0fF", hour.TempF),
            chanceOfRainStr,
            hour.Condition.Text,
        })
    }

    // Render the table
    t.Render()

}