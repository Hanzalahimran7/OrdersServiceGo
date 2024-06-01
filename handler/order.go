package handler

import (
	"log"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("Create Order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	log.Println("Listing Orders")
}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	log.Println("Get Order by ID")
}

func (o *Order) PutById(w http.ResponseWriter, r *http.Request) {
	log.Println("Put Order by ID")
}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	log.Println("Delete Order by ID")
}
