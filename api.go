package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
)

type APIServer struct {
	listenerAddress string
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Error string
}

func NewApiServer(listenerAddress string) *APIServer {
	return &APIServer{listenerAddress: listenerAddress}
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)

		if err != nil {
			log.Printf("Encountered an error: %e\n", err)
			WriteJson(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		}
	}
}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

const SubscribersFileName = "subscribers.txt"

var config Config
var subscribers []string

func loadSubscribers() error {
	data, err := ioutil.ReadFile(SubscribersFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	subscribers = strings.Split(string(data), "\n")
	return nil
}

func handleRate(w http.ResponseWriter, r *http.Request) error {
	err := checkHttpMethod("GET", r, w)
	if err != nil {
		return err
	}

	finalRate, err := GetFinalUahRate()

	if err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, finalRate)
}

func handleSubscribe(w http.ResponseWriter, r *http.Request) error {
	err := checkHttpMethod("POST", r, w)
	if err != nil {
		return err
	}

	email := r.FormValue("email")
	_, err = mail.ParseAddress(email)
	if err != nil {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return err
	}

	for _, subscriber := range subscribers {
		if subscriber == email {
			http.Error(w, "Email already subscribed", http.StatusBadRequest)
			return fmt.Errorf("email already subscribed")
		}
	}

	subscribers = append(subscribers, email)

	if err := saveSubscribersToFile(); err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, "Subscribed successfully")
}

func checkHttpMethod(methodName string, r *http.Request, w http.ResponseWriter) error {
	if r.Method != methodName {
		return WriteJson(w, http.StatusBadRequest, "This HTTP method is not allowed: %s")
	}
	return nil
}

func saveSubscribersToFile() error {
	data := strings.Join(subscribers, "\n")
	return ioutil.WriteFile(SubscribersFileName, []byte(data), 0644)
}

func handleSendEmails(w http.ResponseWriter, r *http.Request) error {
	err := checkHttpMethod("POST", r, w)
	if err != nil {
		return err
	}

	finalRate := fmt.Sprint(GetFinalUahRate())
	for _, subscriber := range subscribers {
		err := SendEmailOAUTH2(subscriber, finalRate)
		if err != nil {
			return err
		}
	}
	return WriteJson(w, http.StatusOK, "Emails sent successfully")
}

func GetFinalUahRate() (float64, error) {
	ch := make(chan float64)

	var apiCallErrors error = nil

	go func() {
		err := getRateFromAbstractAPI(ch)
		if err != nil {
			apiCallErrors = err
			return
		}
	}()

	go func() {
		err := getRateFromExchangeRatesAPI(ch)
		if err != nil {
			apiCallErrors = err
			return
		}
	}()

	rate1 := <-ch
	rate2 := <-ch

	finalRate := rate1 * rate2
	return finalRate, apiCallErrors
}

func getRateFromAbstractAPI(ch chan<- float64) error {

	resp, err := http.Get(getAbstractApiLink())

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response AbstractAPIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	rate, ok := response.ExchangeRates["EUR"]
	if !ok {
		log.Fatal("Rate not found")
	}

	ch <- rate

	return nil
}

func getRateFromExchangeRatesAPI(ch chan<- float64) error {
	resp, err := http.Get(fmt.Sprintf(getExchangeApiLink()))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response ExchangeRatesAPIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	rate, ok := response.Rates["UAH"]
	if !ok {
		log.Fatal("Rate not found")
	}

	ch <- rate
	return nil
}

func getAbstractApiLink() string {

	return fmt.Sprintf(
		"https://exchange-rates.abstractapi.com/v1/live/?api_key=%s&base=BTC&target=EUR",
		config.AbstractAPIKey)
}

func getExchangeApiLink() string {
	return fmt.Sprintf(
		"http://api.exchangeratesapi.io/v1/latest?access_key=%s",
		config.ExchangeRatesAPIKey)
}
func loadConfiguration() {

	config.AbstractAPIKey = os.Getenv("ABSTRACT_API_KEY")
	if config.AbstractAPIKey == "" {
		log.Fatal("ABSTRACT_API_KEY environment variable is not set.")
	}

	config.ExchangeRatesAPIKey = os.Getenv("EXCHANGE_RATES_API_KEY")
	if config.AbstractAPIKey == "" {
		log.Fatal("EXCHANGE_RATES_API_KEY environment variable is not set.")
	}

}

func loggingMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}

func (s *APIServer) Run() {
	err := loadSubscribers()
	if err != nil {
		log.Fatal("Couldn't load subscribers")
	}

	loadConfiguration()
	SetUpOAuthGmailService()

	router := mux.NewRouter()

	router.HandleFunc("/rate", loggingMiddleware(makeHTTPHandleFunc(handleRate)))
	router.HandleFunc("/subscribe", loggingMiddleware(makeHTTPHandleFunc(handleSubscribe)))
	router.HandleFunc("/sendEmails", loggingMiddleware(makeHTTPHandleFunc(handleSendEmails)))

	http.ListenAndServe(s.listenerAddress, router)
}
