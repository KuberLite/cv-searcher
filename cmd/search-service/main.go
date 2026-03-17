package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KuberLite/cv-searcher/internal/config"
	"github.com/KuberLite/cv-searcher/internal/handler"
	"github.com/KuberLite/cv-searcher/internal/meilisearch"
)

func main() {
	time.Sleep(2 * time.Second)
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

	client := meilisearch.New(cfg.MeiliSearchURL, "products")
	go func() {
		searchHandler := handler.NewSearchHandler(client)
		http.HandleFunc("/api/v1/search", searchHandler.Search)
		log.Println("Staring HTTP server")
		if err := http.ListenAndServe(cfg.HttpPort, nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Service stopped")
}
