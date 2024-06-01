package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/hanzalahimran7/MicroserviceInGo/application"
)

func main() {
	app := application.NewApp()
	//Graceful shutdown in case of signit (ctrl + c)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		log.Fatal("Error starting the server", err)
	}
}
