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
	"sync"
)

const LowRatio = 0.2
const CrashRatio = 0.8
const PriceIncrement = 0.04
const ClockPeriodMinutes = 1

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
	lock         sync.RWMutex
	reset        chan struct{}
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
	Crash  *int            `json:"crash"`
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

var crash struct {
	ID   *int
	lock sync.RWMutex
}

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

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	r.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		var m menuResponse

		for _, product := range fakeMenu.Items {
			product.lock.RLock()
			m.Items = append(m.Items, itemResponse{
				ID:   product.ID,
				Name: product.Name,
			})
			product.lock.RUnlock()
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}).Methods(http.MethodGet)

	r.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
		var p pricesResponse

		for _, product := range fakeMenu.Items {
			product.lock.RLock()
			p.Prices = append(p.Prices, newPriceResp(*product))
			product.lock.RUnlock()
		}

		crash.lock.RLock()
		p.Crash = crash.ID
		crash.lock.RUnlock()

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

			menuProduct.IncrPrice()
		}

		w.WriteHeader(http.StatusNoContent)
	})

	c := cors.AllowAll()

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), c.Handler(r))
	if err != nil {
		log.Fatal(err)
	}
}

func NewProduct(ID int, name string, price int) *Product {
	initialPrice := int(float64(price) * LowRatio)
	product := &Product{
		ID:           ID,
		Name:         name,
		BasePrice:    price,
		lowPrice:     initialPrice,
		currentPrice: initialPrice,
		highPrice:    initialPrice,
		reset:        make(chan struct{}),
	}

	go product.Run()

	return product
}

func (product *Product) Run() {
	timer := time.NewTimer(ClockPeriodMinutes * time.Minute)
	select {
	case <-timer.C:
		product.DecrPrice()
	case <-product.reset:
	}
	go product.Run()
}

func (product *Product) minPrice() int {
	return int(float64(product.BasePrice) * LowRatio)
}

func (product *Product) maxPrice() int {
	return int(float64(product.BasePrice) * CrashRatio)
}

func (product *Product) IncrPrice() {
	product.lock.Lock()
	defer product.lock.Unlock()

	product.reset <- struct{}{}

	newPrice := int(math.Ceil(float64(product.currentPrice) * (1.0 + PriceIncrement)))

	if (newPrice > product.maxPrice()) {
		product.currentPrice = product.minPrice()
		crash.lock.Lock()
		crash.ID = &product.ID
		crash.lock.Unlock()

		go func() {
			select {
			case <-time.After(2 * time.Second):
				crash.lock.Lock()
				if *crash.ID == product.ID {
					crash.ID = nil
				}
				crash.lock.Unlock()
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
	product.lock.Lock()
	defer product.lock.Unlock()

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
