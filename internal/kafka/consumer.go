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

type Consumer struct {
	reader *kafka.Reader
}

func New(cfg config.Config) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroup,
	})
	return &Consumer{
		reader: reader,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	defer c.reader.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				continue
			}
			data := model.ProductEvent{}
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				log.Printf("JSON parse error: %v", err)
				continue
			}
			fmt.Println(string(data.Action) + " " + string(data.Payload.Description))
		}
	}
}
