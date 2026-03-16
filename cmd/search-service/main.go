package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KuberLite/cv-searcher/internal/config"
	"github.com/KuberLite/cv-searcher/internal/meilisearch"
	"github.com/KuberLite/cv-searcher/internal/model"
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
		//if err := seedTestData(client); err != nil {
		//	log.Fatal(err)
		//}

		product := model.Product{
			ID:          "1",
			Name:        "iPhone 15 Pro",
			Description: "Смартфон от Apple с процессором A17 Pro",
		}

		err := client.DeleteProduct(context.Background(), product)

		if err != nil {
			log.Fatalln(fmt.Errorf("delete error: %w", err))
		}
	}()

	<-ctx.Done()
	log.Println("Service stopped")
}
