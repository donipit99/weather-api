package models

type WeatherTodayParams struct {
	Lat  float64
	Lon  float64
	Lang string
}

type WeatherResult struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WeatherCode int     `json:"weathercode"`
	} `json:"current_weather"`
}
