package main

type AbstractAPIResponse struct {
	Base          string             `json:"base"`
	LastUpdated   int64              `json:"last_updated"`
	ExchangeRates map[string]float64 `json:"exchange_rates"`
}

type ExchangeRatesAPIResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

type Config struct {
	AbstractAPIKey      string `json:"AbstractAPIKey"`
	ExchangeRatesAPIKey string `json:"ExchangeRatesAPIKey"`
}
