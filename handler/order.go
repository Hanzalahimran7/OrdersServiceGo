package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hanzalahimran7/MicroserviceInGo/model"
	"github.com/hanzalahimran7/MicroserviceInGo/respository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) (error, int) {
	//anynomous struct for what to expect from request
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("INVALID REQUEST BODY"), http.StatusBadRequest
	}
	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		CreatedAt:  &now,
		LineItems:  body.LineItems,
	}
	if err := o.Repo.Insert(r.Context(), order); err != nil {
		return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
	}
	WriteJSON(w, http.StatusCreated, order)
	return nil, 0
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) (error, int) {
	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		cursor = "0"
	}
	cur, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil {
		return fmt.Errorf("INVALID CURSOR: %v", cursor), http.StatusBadRequest
	}
	const size = 50
	res, err := o.Repo.ListOrders(r.Context(), order.FindAllPage{
		Offset: uint(cur),
		Size:   size,
	})
	if err != nil {
		return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
	}
	var body struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	body.Items = res.Orders
	body.Next = res.Cursor
	WriteJSON(w, http.StatusCreated, body)
	return nil, 0

}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) (error, int) {
	idParam := chi.URLParam(r, "id")
	cur, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fmt.Errorf("INVALID CURSOR: %v", idParam), http.StatusBadRequest
	}
	order, err := o.Repo.FindByID(r.Context(), cur)
	if err != nil {
		if err.Error() == "order does not exist" {
			return fmt.Errorf("NOT FOUND"), http.StatusNotFound
		} else {
			log.Println(err.Error())
			return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
		}
	}
	WriteJSON(w, http.StatusCreated, order)
	return nil, 0
}

func (o *Order) PutById(w http.ResponseWriter, r *http.Request) (error, int) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("BAD REQUEST"), http.StatusBadRequest
	}
	idParam := chi.URLParam(r, "id")
	cur, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fmt.Errorf("INVALID CURSOR: %v", idParam), http.StatusBadRequest
	}
	order, err := o.Repo.FindByID(r.Context(), cur)
	if err != nil {
		if err.Error() == "order does not exist" {
			return fmt.Errorf("NOT FOUND"), http.StatusNotFound
		} else {
			log.Println(err.Error())
			return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
		}
	}
	now := time.Now()
	switch body.Status {
	case "shipped":
		if order.ShippedAt != nil {
			return fmt.Errorf("ITEM ALREADY SHIPPED"), http.StatusBadRequest
		}
		order.ShippedAt = &now
	case "completed":
		if order.CompletedAt != nil {
			return fmt.Errorf("ITEM ALREADY Completed"), http.StatusBadRequest
		}
		order.CompletedAt = &now
	default:
		return fmt.Errorf("INVALID STATUS"), http.StatusBadRequest
	}
	err = o.Repo.UpdateOrder(r.Context(), order)
	if err != nil {
		return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
	}
	WriteJSON(w, http.StatusCreated, order)
	return nil, 0
}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) (error, int) {
	idParam := chi.URLParam(r, "id")
	cur, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return fmt.Errorf("INVALID CURSOR: %v", idParam), http.StatusBadRequest
	}

	err = o.Repo.DeleteById(r.Context(), cur)
	log.Println(err)
	if err != nil {
		if err.Error() == "order does not exist" {
			return fmt.Errorf("NOT FOUND"), http.StatusNotFound
		} else {
			log.Println(err.Error())
			return fmt.Errorf("INTERNAL SERVER ERROR"), http.StatusInternalServerError
		}
	}
	w.WriteHeader(http.StatusOK)
	return nil, 0
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
