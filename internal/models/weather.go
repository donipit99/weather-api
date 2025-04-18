package models

type WeatherTodayParams struct {
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

var WeatherCodeMap = map[int]string{ // переменная WeatherCodeMap (словарь), сопостовляет код погоды и описание
	0:  "Ясно",
	1:  "Преимущественно ясно",
	2:  "Переменная облачность",
	3:  "Облачно",
	45: "Туман",
	48: "Инейный туман",
	51: "Легкая морось",
	53: "Умеренная морось",
	55: "Сильная морось",
	61: "Небольшой дождь",
	63: "Умеренный дождь",
	65: "Сильный дождь",
	71: "Небольшой снег",
	73: "Умеренный снег",
	75: "Сильный снег",
	95: "Гроза",
}

func GetWeatherDescription(code int) string { // функция GetWeatherDescription принимает код погоды (int) и возвращает описание (string)
	if desc, ok := WeatherCodeMap[code]; ok { // если код есть в словаре WeatherCodeMap,
		return desc // то возвращает описание
	}
	return "Неизвестно" //  если нет - неизвестно
}
