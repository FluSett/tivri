package main

import (
	"log"
	"tivri/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer server.Close()
	err = server.Start()
	if err != nil {
		log.Fatalf("application error: %v", err)
	}
}
