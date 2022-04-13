package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Temperature float64

func (t Temperature) Fahrenheit() float64 {
	return (float64(t)-273.15)*(9.0/5.0) + 32.0
}

//"Conditions" struct contains the desire value
type Conditions struct {
	weather     string
	temperature Temperature
	pressure    float64
	humidity    float64
	windSpeed   float64
	windAngle   float64
}

//"OWMresponse" struct contains the desired data abstruct from json file
type OWMResponse struct { //change the struct formation here to match needed value in jason
	Weather []struct {
		Main string
	}
	Main struct {
		Temp     Temperature
		Pressure float64
		Humidity float64
	}
	Wind struct {
		Speed float64
		Angle float64 `json:"deg"`
	}
}

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(key string) *Client {
	return &Client{
		APIKey:  key,
		BaseURL: "https://api.openweathermap.org",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c Client) FormatURL(location string) string {
	location = url.QueryEscape(location)
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", c.BaseURL, location, c.APIKey)
}

//get the http request and converted it into a "Conditions" stuct
func (c *Client) GetWeather(location string) (Conditions, error) {
	URL := c.FormatURL(location)
	resp, err := c.HTTPClient.Get(URL)
	if err != nil {
		return Conditions{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Conditions{}, fmt.Errorf("could not find location: %s ", location)
	}
	if resp.StatusCode != http.StatusOK {
		return Conditions{}, fmt.Errorf("unexpected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Conditions{}, err
	}
	conditions, err := ParseResponse(data)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

//This is the where the converstion take place (Json format-> "Conditions")
func ParseResponse(data []byte) (Conditions, error) {
	var resp OWMResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Conditions{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Weather) < 1 {
		return Conditions{}, fmt.Errorf("invalid API response %s: require at least one weather element", data)
	}
	conditions := Conditions{
		weather:     resp.Weather[0].Main,
		temperature: resp.Main.Temp,
		pressure:    resp.Main.Pressure,
		humidity:    resp.Main.Humidity,
		windSpeed:   resp.Wind.Speed,
		windAngle:   resp.Wind.Angle,
	}
	return conditions, nil
}

//this function has not been used
func FormatURL(baseURL, location, key string) string {
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", baseURL, location, key)
}

func Get(location, key string) (Conditions, error) {
	c := NewClient(key)
	conditions, err := c.GetWeather(location)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

func RunCLI() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s LOCATION\nExample: %[1]s London,UK\n", os.Args[0])
		os.Exit(1)
	}
	location := os.Args[1]
	key := os.Getenv("OPENWEATHERMAP_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "Please set the environment variable OPENWEATHERMAP_API_KEY")
		os.Exit(1)
	}
	conditions, err := Get(location, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Weather:%s Temperature:%.1fº Humidity:%.1f Pressure%.1f\n", conditions.weather, conditions.temperature.Fahrenheit(), conditions.humidity, conditions.pressure)
	fmt.Printf("WindSpeed:%.1f WindAngle:%.1f\n", conditions.windSpeed, conditions.windAngle)
}
