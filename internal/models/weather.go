package models

type WeatherTodayParams struct {
	Lat float64
	Lon float64
}

type CurrentWeather struct {
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weathercode"`
}

type WeatherResult struct {
	CurrentWeather CurrentWeather `json:"current_weather"`
}
