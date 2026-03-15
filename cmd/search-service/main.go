package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KuberLite/cv-searcher/internal/config"
	"github.com/KuberLite/cv-searcher/internal/kafka"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	consumer := kafka.New(cfg)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("Service stopped")
}
