package main

import (
	"os"
	"net/http"
	"log"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
)

type item struct{
	Name string `json:"name"`
}

type menu struct{
	Items []item `json:"items"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		menu := menu{
			Items: []item{
				{Name:"Stella"},
				{Name:"Carlsberg"},
				{Name:"Coors Light"},
				{Name:"Carling"},
				{Name:"Budweiser"},
			},
		}

		json.NewEncoder(w).Encode(menu)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		log.Fatal(err)
	}
}
