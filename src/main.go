package main

import (
	"log"
)

func main() {
	app, err := initialize()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := app.createTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	app.SetupRoutes()

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
