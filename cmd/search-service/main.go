package main

import (
	"context"
	"errors"
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
	"github.com/KuberLite/cv-searcher/internal/repo"
	"github.com/KuberLite/cv-searcher/internal/vectorizer"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := repo.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	goose.SetDialect("postgres")
	if err := goose.Up(database.DB(), "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	defer database.Close()

	meiliClient := meilisearch.New(cfg.MeiliSearchURL, "products")
	vectorizerClient := vectorizer.New(cfg.VectorizerURL)
	qdrantClient, err := qdrantclient.New(cfg.QDRantURL.URL, cfg.QDRantURL.Port, "products")
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}
	if err = qdrantClient.InitCollection(ctx, 1024); err != nil {
		log.Fatalf("Failed to initialize QDRant: %v", err)
	}

	if err := meiliClient.ConfigureIndex(ctx); err != nil {
		log.Fatalf("Failed to configure Meilisearch index: %v", err)
	}

	mux := http.NewServeMux()
	searchHandler := handler.NewSearchHandler(meiliClient, vectorizerClient, qdrantClient)
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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
