package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	nominatim "github.com/doppiogancio/go-nominatim"
	"github.com/hectormalot/omgo"
	"github.com/maltegrosse/go-geoclue2"
)

/*
WMO Weather interpretation codes (WW)
Code	Description
0	Clear sky
1, 2, 3	Mainly clear, partly cloudy, and overcast
45, 48	Fog and depositing rime fog
51, 53, 55	Drizzle: Light, moderate, and dense intensity
56, 57	Freezing Drizzle: Light and dense intensity
61, 63, 65	Rain: Slight, moderate and heavy intensity
66, 67	Freezing Rain: Light and heavy intensity
71, 73, 75	Snow fall: Slight, moderate, and heavy intensity
77	Snow grains
80, 81, 82	Rain showers: Slight, moderate, and violent
85, 86	Snow showers slight and heavy
95 *	Thunderstorm: Slight or moderate
96, 99 *	Thunderstorm with slight and heavy hail

(*) Thunderstorm forecast with hail is only available in Central Europe
*/

// WMOWeatherCodes maps WMO weather code integers to their descriptions
var WMOWeatherCodes = map[float64]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	56: "Light freezing drizzle",
	57: "Dense freezing drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	66: "Light freezing rain",
	67: "Heavy freezing rain",
	71: "Slight snow fall",
	73: "Moderate snow fall",
	75: "Heavy snow fall",
	77: "Snow grains",
	80: "Slight rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	85: "Slight snow showers",
	86: "Heavy snow showers",
	95: "Thunderstorm",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

var ErrLocationNotAccurate = errors.New("location service is not accurate enough")

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, err := omgo.NewClient()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create client: %s\n", err)
		os.Exit(1)
	}
	lat, long, err := getLocation()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get latitude/longitude: %s\n", err)
		os.Exit(1)
	}
	address, err := nominatim.ReverseGeocode(lat, long, "german")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to reverse geocode: %s\n", err)
		os.Exit(1)
	}

	location, err := omgo.NewLocation(lat, long)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to set location: %s\n", err)
		os.Exit(1)
	}
	current, err := client.CurrentWeather(ctx, location, &omgo.Options{
		Timezone: "Europe/Berlin",
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get current weather: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Weather conditions in %s: %s with a temperature of %.1f°C\n", address.City,
		WMOWeatherCodes[current.WeatherCode], current.Temperature)
}

func getLocation() (float64, float64, error) {
	gcm, err := geoclue2.NewGeoclueManager()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create geoclue manager: %w", err)
	}

	client, err := gcm.GetClient()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get client: %w", err)
	}
	if err = client.SetDesktopId("waybar-weather"); err != nil {
		return 0, 0, fmt.Errorf("failed to set desktop id: %w", err)
	}

	if err = client.SetRequestedAccuracyLevel(geoclue2.GClueAccuracyLevelStreet); err != nil {
		return 0, 0, fmt.Errorf("failed to set requested accuracy level: %w", err)
	}

	// Get RequestedAccuracyLevel
	level, err := client.GetRequestedAccuracyLevel()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get requested accuracy level: %w", err)
	}
	fmt.Printf("Accuary: %s\n", level.String())
	if level < 4 {
		return 0, 0, ErrLocationNotAccurate
	}

	if err = client.Start(); err != nil {
		return 0, 0, fmt.Errorf("failed to start client: %w", err)
	}

	// create new Instance of Geoclue Location
	location, err := client.GetLocation()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get location: %w", err)
	}

	latitude, err := location.GetLatitude()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get latitude: %w", err)
	}

	// get longitude
	longitude, err := location.GetLongitude()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get longitude: %w", err)
	}

	return latitude, longitude, nil
}
