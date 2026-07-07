package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tivri/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	defer server.Close()

	err = server.Start()
	if err != nil {
		log.Fatalf("application error: %v", err)
	}
}
