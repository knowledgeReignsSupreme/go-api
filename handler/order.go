package handler

import (
	"fmt"
	"net/http"
)

type Order struct {}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List orders")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get order By ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update order By ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete order By ID")
}