package application

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hanzalahimran7/MicroserviceInGo/handler"
	"github.com/hanzalahimran7/MicroserviceInGo/respository/order"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Route("/orders", a.loadOrderRoutes)
	a.router = router
}

type apiFunc func(http.ResponseWriter, *http.Request) (error, int)

type APIError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func httpHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err, status_code := f(w, r)
		if err != nil {
			WriteJSON(w, status_code, APIError{Error: err.Error()})
		}
	}
}
func (a *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{Client: a.rdb},
	}
	router.Post("/", httpHandlerFunc(orderHandler.Create))
	router.Get("/", httpHandlerFunc(orderHandler.List))
	router.Get("/{id}", orderHandler.GetById)
	router.Delete("/{id}", orderHandler.DeleteById)
	router.Put("/{id}", orderHandler.PutById)
}
