package dto

type GetWeatherTodayParams struct {
	Lat float64
	Lon float64
}

type CurrentWeather struct {
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weathercode"`
	WeatherDesc string  `json:"weather_description"`
}

type WeatherResult struct {
	CurrentWeather CurrentWeather `json:"current_weather"`
}
