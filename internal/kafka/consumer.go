package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/KuberLite/cv-searcher/internal/config"
	"github.com/KuberLite/cv-searcher/internal/model"
	"github.com/segmentio/kafka-go"
)

type ProductIndexer interface {
	IndexProduct(ctx context.Context, product model.Product) error
	DeleteProduct(ctx context.Context, product model.Product) error
}

type Consumer struct {
	reader  *kafka.Reader
	indexer ProductIndexer
}

func New(cfg config.Config, indexer ProductIndexer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroup,
	})
	return &Consumer{
		reader:  reader,
		indexer: indexer,
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
		return c.indexer.IndexProduct(ctx, event.Payload)
	case "delete":
		return c.indexer.DeleteProduct(ctx, event.Payload)
	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}
