package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/KuberLite/cv-searcher/internal/config"
	"github.com/KuberLite/cv-searcher/internal/model"
	"github.com/KuberLite/cv-searcher/internal/qdrant"
	"github.com/KuberLite/cv-searcher/internal/vectorizer"
	"github.com/segmentio/kafka-go"
)

type ProductIndexer interface {
	IndexProduct(ctx context.Context, product model.Product) error
	DeleteProduct(ctx context.Context, product model.Product) error
}

type Consumer struct {
	reader       *kafka.Reader
	indexer      ProductIndexer
	vectorizer   *vectorizer.Client
	qdrantClient *qdrantclient.Client
}

func New(cfg config.Config, indexer ProductIndexer, vectorizer *vectorizer.Client, qdrantClient *qdrantclient.Client) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroup,
	})
	return &Consumer{
		reader:       reader,
		indexer:      indexer,
		vectorizer:   vectorizer,
		qdrantClient: qdrantClient,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	defer c.reader.Close()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("Kafka read error: %v", err)
			continue
		}

		var event model.ProductEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal product event: %v", err)
			continue
		}

		if err := c.processEvent(ctx, event); err != nil {
			log.Printf("Failed to process event (action=%s, id=%s): %v",
				event.Action, event.Payload.ID, err)
		}
	}
}

func (c *Consumer) processEvent(ctx context.Context, event model.ProductEvent) error {
	switch event.Action {
	case "create", "update":
		if err := c.indexer.IndexProduct(ctx, event.Payload); err != nil {
			return err
		}

		if c.vectorizer != nil && c.qdrantClient != nil {
			vector, err := c.vectorizer.GetProductVector(
				ctx,
				event.Payload.Name,
				event.Payload.Description,
			)
			if err == nil {
				payload := map[string]interface{}{
					"name":        event.Payload.Name,
					"description": event.Payload.Description,
					"brand":       event.Payload.Brand,
				}
				_ = c.qdrantClient.IndexProduct(ctx, event.Payload.ID, vector, payload)
			}
		}

		return nil
	case "delete":
		if err := c.indexer.DeleteProduct(ctx, event.Payload); err != nil {
			return err
		}

		if c.qdrantClient != nil {
			_ = c.qdrantClient.DeleteProduct(ctx, event.Payload.ID)
		}
		return nil
	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}
