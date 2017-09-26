package main

import (
	"os"
	"net/http"
	"log"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/rs/cors"
	"math"
	"time"
)

const LowRatio = 0.2
const CrashRatio = 0.8
const PriceIncrement = 0.04
const ClockPeriodMinutes = 5

const TrendUp = "up"
const TrendDown = "down"

type Menu struct {
	Items []*Product
}

type Product struct {
	ID           int
	Name         string
	BasePrice    int
	lowPrice     int
	currentPrice int
	highPrice    int
	Trend        string
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
	Crash  *int            `json:"crash""`
}

type billEvent struct {
	Bill billEventBill `json:"bill"`
}

type billEventBill struct {
	Products []billEventProduct `json:"products"`
}

type billEventProduct struct {
	ID int `json:"flypayProductId"`
}

var fakeMenu Menu

var crash *int

func main() {

	fakeMenu = Menu{
		Items: []*Product{
			NewProduct(1, "Stella", 540),
			NewProduct(2, "Carlsberg", 480),
			NewProduct(3, "Coors Light", 420),
			NewProduct(4, "Carling", 480),
			NewProduct(5, "Budweiser", 480),
		},
	}

	go priceDecrementer()

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
			p.Prices = append(p.Prices, newPriceResp(*product))
		}

		p.Crash = crash

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	}).Methods(http.MethodGet)

	r.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		var event billEvent
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		for _, product := range event.Bill.Products {
			menuProduct, err := fakeMenu.Product(product.ID)
			if err != nil {
				continue
			}

			log.Printf("Found product, price: %v", menuProduct.currentPrice)
			menuProduct.IncrPrice()
			log.Printf("Incr product, price: %v", menuProduct.currentPrice)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	c := cors.AllowAll()

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), c.Handler(r))
	if err != nil {
		log.Fatal(err)
	}
}

func priceDecrementer() {
	ticker := time.NewTicker(time.Minute * ClockPeriodMinutes)
	for range ticker.C {
		for _, product := range fakeMenu.Items {
			product.DecrPrice()
		}
	}
}

func NewProduct(ID int, name string, price int) *Product {
	initialPrice := int(float64(price) * LowRatio)
	return &Product{
		ID:           ID,
		Name:         name,
		BasePrice:    price,
		lowPrice:     initialPrice,
		currentPrice: initialPrice,
		highPrice:    initialPrice,
	}
}

func (product *Product) minPrice() int {
	return int(float64(product.BasePrice) * LowRatio)
}

func (product *Product) maxPrice() int {
	return int(float64(product.BasePrice) * CrashRatio)
}

func (product *Product) IncrPrice() {
	newPrice := int(math.Ceil(float64(product.currentPrice) * (1.0 + PriceIncrement)))

	if (newPrice > product.maxPrice()) {
		product.currentPrice = product.minPrice()
		crash = &product.ID
		go func() {
			select {
			case <-time.After(2 * time.Second):
				if *crash == product.ID {
					crash = nil
				}
			}
		}()
		product.Trend = TrendDown
		return
	}

	product.currentPrice = newPrice
	product.Trend = TrendUp

	if product.currentPrice > product.highPrice {
		product.highPrice = product.currentPrice
	}
}

func (product *Product) DecrPrice() {
	newPrice := int(math.Floor(float64(product.currentPrice) * (1 - PriceIncrement)))

	minPrice := product.minPrice()
	if newPrice < minPrice {
		product.currentPrice = product.minPrice()
		product.Trend = ""
		return
	}

	product.currentPrice = newPrice
	product.Trend = TrendDown
}

func (menu Menu) Product(productID int) (*Product, error) {
	for _, product := range menu.Items {
		if product.ID == productID {
			return product, nil
		}
	}

	return nil, fmt.Errorf("product %d not found", productID)
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
	return product.currentPrice
}

func (product Product) High() int {
	return product.highPrice
}

func (product Product) Low() int {
	return product.lowPrice
}

func toMoney(amount int) string {
	return fmt.Sprintf("£%.2f", float64(amount)/100.0)
}
