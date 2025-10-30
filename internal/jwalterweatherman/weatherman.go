// Package weatherman provides JSON weather data for testing
package weatherman

type WeatherData struct {
	TemperatureC float64 `json:"temperature_c"`
	Humidity     float64 `json:"humidity"`
	Condition     string    `json:"condition"`
}

func GetTestData() *WeatherData {
	return &WeatherData{
		TemperatureC: 22.5,
		Humidity:     65.0,
		Condition:     "partly_cloudy",
	}
}