package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func NewApp() *App {
	app := &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}
	fmt.Println(a.rdb.Options().Addr)
	if err := a.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("CANNOT CONNECT TO REDIS")
	}

	defer func() {
		log.Println("Closing the redis connection")
		a.rdb.Close()
	}()

	log.Println("Connected to Redis")
	log.Println("Starting API Server")
	ch := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- fmt.Errorf("ERROR RUNNING THE SERVER")
		}
		close(ch)
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		log.Println("Shutting down the server")
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
