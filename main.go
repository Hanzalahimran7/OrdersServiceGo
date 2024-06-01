package main

import (
	"context"
	"log"

	"github.com/hanzalahimran7/MicroserviceInGo/application"
)

func main() {
	app := application.NewApp()
	if err := app.Start(context.TODO()); err != nil {
		log.Fatal("Error starting the server", err)
	}
}
