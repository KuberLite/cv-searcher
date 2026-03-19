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
	"github.com/KuberLite/cv-searcher/internal/kafka"
	"github.com/KuberLite/cv-searcher/internal/meilisearch"
	"github.com/KuberLite/cv-searcher/internal/qdrant"
	"github.com/KuberLite/cv-searcher/internal/vectorizer"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	meiliClient := meilisearch.New(cfg.MeiliSearchURL, "products")
	vectorizerClient := vectorizer.New(cfg.VectorizerURL)
	qdrantClient, err := qdrantclient.New(cfg.QDRantURL.URL, cfg.QDRantURL.Port, "products")
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}
	if err = qdrantClient.InitCollection(ctx, 384); err != nil {
		log.Fatalf("Failed to initialize QDRant: %v", err)
	}

	err = meiliClient.ConfigureIndex(ctx)
	if err != nil {
		log.Fatal("Warning: failed to configure index: %w", err)
	}

	mux := http.NewServeMux()
	searchHandler := handler.NewSearchHandler(meiliClient)
	mux.HandleFunc("/api/v1/search", searchHandler.Search)

	srv := &http.Server{
		Addr:         cfg.HttpPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.HttpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	consumer := kafka.New(cfg, meiliClient, vectorizerClient, qdrantClient)
	go func() {
		log.Println("Starting Kafka consumer")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer stopped with error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Service stopped")
}
