package main

import (
	"os"
	"net/http"
	"log"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/rs/cors"
)

const LowRatio = 0.2
const CrashRatio = 0.8
const PriceIncrement = 0.04
const ClockPeriodMinutes = 5

type Menu struct {
	Items []Product
}

type Product struct {
	ID    int
	Name  string
	Price int
	Trend string
}

type itemResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	price int
}

type menuResponse struct {
	Items []itemResponse `json:"items"`
}

type priceResponse struct {
	ID      int    `json:"id"`
	Low     string `json:"low"`
	High    string `json:"high"`
	Current string `json:"current"`
	Trend   string `json:"trend"`
}

type pricesResponse struct {
	Prices []priceResponse `json:"prices"`
}

var fakeMenu Menu

func main() {

	fakeMenu = Menu{
		Items: []Product{
			{ID: 1, Name: "Stella", Price: 540},
			{ID: 2, Name: "Carlsberg", Price: 480},
			{ID: 3, Name: "Coors Light", Price: 420},
			{ID: 4, Name: "Carling", Price: 480},
			{ID: 5, Name: "Budweiser", Price: 480},
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	r.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		var m menuResponse

		for _, product := range fakeMenu.Items {
			m.Items = append(m.Items, itemResponse{
				ID:   product.ID,
				Name: product.Name,
			})
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}).Methods(http.MethodGet)

	r.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
		var p pricesResponse

		for _, product := range fakeMenu.Items {
			p.Prices = append(p.Prices, newPriceResp(product))
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}).Methods(http.MethodGet)

	c := cors.AllowAll()

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), c.Handler(r))
	if err != nil {
		log.Fatal(err)
	}
}

func newPriceResp(product Product) priceResponse {
	return priceResponse{
		ID:      product.ID,
		Low:     toMoney(product.Low()),
		High:    toMoney(product.High()),
		Current: toMoney(product.Current()),
		Trend:   product.Trend,
	}
}

func (product Product) Current() int {
	return product.Low()
}

func (product Product) High() int {
	return product.Low()
}

func (product Product) Low() int {
	return int(float64(product.Price) * LowRatio)
}

func toMoney(amount int) string {
	return fmt.Sprintf("£%.2f", float64(amount)/100.0)
}
